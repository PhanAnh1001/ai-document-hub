"""
LLM extraction endpoints: trigger extraction, get results.
"""
import logging
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.database import get_db
from app.config import settings
from app.models.db_models import Document, DocumentStatus
from app.models.schemas import ExtractionRequest, ExtractionResponse
from app.services.extraction_service import ExtractionService
from app.services.llm_client import LLMClient

logger = logging.getLogger(__name__)
router = APIRouter()


def get_extraction_service() -> ExtractionService:
    llm = LLMClient(settings)
    return ExtractionService(llm)


@router.post("/process/{doc_id}", response_model=ExtractionResponse)
async def process_extraction(
    doc_id: str,
    request: ExtractionRequest = None,
    db: AsyncSession = Depends(get_db),
    extraction_service: ExtractionService = Depends(get_extraction_service),
):
    """Trigger LLM extraction for a document."""
    if request is None:
        request = ExtractionRequest()

    result = await db.execute(select(Document).where(Document.id == doc_id))
    doc = result.scalar_one_or_none()
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")

    # Determine doc type
    doc_type = request.doc_type or doc.doc_type or "other"

    # Use OCR text or fallback to filename
    text = doc.ocr_text or f"Document: {doc.original_filename}"

    doc.status = DocumentStatus.EXTRACTING.value
    if request.doc_type:
        doc.doc_type = doc_type
    await db.flush()

    try:
        extracted = await extraction_service.extract(text, doc_type)

        doc.extracted_data = extracted
        doc.status = DocumentStatus.EXTRACTED.value
        await db.flush()

        return ExtractionResponse(
            doc_id=doc_id,
            doc_type=doc_type,
            status=doc.status,
            extracted_data=extracted,
        )
    except Exception as e:
        logger.error(f"Extraction failed for doc {doc_id}: {e}")
        doc.status = DocumentStatus.FAILED.value
        doc.error_message = str(e)
        await db.flush()
        raise HTTPException(status_code=500, detail=f"Extraction failed: {e}")


@router.get("/{doc_id}/result", response_model=ExtractionResponse)
async def get_extraction_result(doc_id: str, db: AsyncSession = Depends(get_db)):
    """Get extraction result for a document."""
    result = await db.execute(select(Document).where(Document.id == doc_id))
    doc = result.scalar_one_or_none()
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")

    return ExtractionResponse(
        doc_id=doc_id,
        doc_type=doc.doc_type or "other",
        status=doc.status,
        extracted_data=doc.extracted_data or {},
    )
