"""
Background task processor using FastAPI BackgroundTasks (no Redis required).
Runs OCR + extraction + indexing pipeline for uploaded documents.
"""
import asyncio
import logging
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker

from app.config import settings
from app.models.db_models import DocumentStatus

logger = logging.getLogger(__name__)


async def process_document_background(doc_id: str, file_path: str, mime_type: str):
    """
    Full processing pipeline: OCR → extraction → RAG indexing.
    Creates its own DB session since background tasks run outside request context.
    """
    from app.services.ocr_service import OCRService
    from app.services.extraction_service import ExtractionService
    from app.services.rag_service import RAGService
    from app.services.llm_client import LLMClient

    engine = create_async_engine(settings.DATABASE_URL, echo=False)
    session_factory = async_sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)

    try:
        async with session_factory() as db:
            from sqlalchemy import select
            from app.models.db_models import Document

            # Fetch document
            result = await db.execute(select(Document).where(Document.id == doc_id))
            doc = result.scalar_one_or_none()
            if doc is None:
                logger.error(f"Background task: document {doc_id} not found")
                return

            # Step 1: OCR
            try:
                doc.status = DocumentStatus.OCR_PROCESSING.value
                await db.commit()

                ocr_service = OCRService(settings)
                text, confidence, provider = await ocr_service.extract_text(file_path, mime_type)

                doc.ocr_text = text
                doc.ocr_confidence = confidence
                doc.status = DocumentStatus.OCR_DONE.value
                await db.commit()
                logger.info(f"OCR done for {doc_id}: {len(text)} chars, provider={provider}")
            except Exception as e:
                logger.error(f"Background OCR failed for {doc_id}: {e}")
                doc.status = DocumentStatus.FAILED.value
                doc.error_message = f"OCR error: {e}"
                await db.commit()
                return

            # Step 2: Extraction
            try:
                doc.status = DocumentStatus.EXTRACTING.value
                await db.commit()

                llm = LLMClient(settings)
                extraction_service = ExtractionService(llm)
                doc_type = doc.doc_type or "other"
                extracted = await extraction_service.extract(text, doc_type)

                doc.extracted_data = extracted
                doc.status = DocumentStatus.EXTRACTED.value
                await db.commit()
                logger.info(f"Extraction done for {doc_id}")
            except Exception as e:
                logger.warning(f"Background extraction failed for {doc_id}: {e}")
                doc.status = DocumentStatus.OCR_DONE.value
                await db.commit()

            # Step 3: RAG Indexing
            try:
                doc.status = DocumentStatus.INDEXING.value
                await db.commit()

                llm = LLMClient(settings)
                rag_service = RAGService(llm, settings)
                await rag_service.index_document(doc_id, text, db)
                await db.commit()

                doc.status = DocumentStatus.INDEXED.value
                await db.commit()
                logger.info(f"Indexing done for {doc_id}")
            except Exception as e:
                logger.warning(f"Background indexing failed for {doc_id}: {e}")
                doc.status = DocumentStatus.EXTRACTED.value
                await db.commit()

    except Exception as e:
        logger.error(f"Background processing failed for {doc_id}: {e}")
    finally:
        await engine.dispose()
