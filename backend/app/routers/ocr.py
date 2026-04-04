"""
OCR endpoints: trigger OCR processing, get results.
"""
import logging
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.database import get_db
from app.config import settings
from app.models.db_models import Document, DocumentStatus
from app.models.schemas import OCRResponse
from app.services.ocr_service import OCRService

logger = logging.getLogger(__name__)
router = APIRouter()


def get_ocr_service() -> OCRService:
    return OCRService(settings)


@router.post("/process/{doc_id}", response_model=OCRResponse)
async def process_ocr(
    doc_id: str,
    db: AsyncSession = Depends(get_db),
    ocr_service: OCRService = Depends(get_ocr_service),
):
    """Trigger OCR processing for a document."""
    result = await db.execute(select(Document).where(Document.id == doc_id))
    doc = result.scalar_one_or_none()
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")

    # Update status to processing
    doc.status = DocumentStatus.OCR_PROCESSING.value
    await db.flush()

    try:
        text, confidence, provider = await ocr_service.extract_text(
            doc.file_path, doc.mime_type or "text/plain"
        )

        doc.ocr_text = text
        doc.ocr_confidence = confidence
        doc.status = DocumentStatus.OCR_DONE.value
        await db.flush()

        return OCRResponse(
            doc_id=doc_id,
            text=text,
            confidence=confidence,
            provider=provider,
            status=doc.status,
        )
    except Exception as e:
        logger.error(f"OCR failed for doc {doc_id}: {e}")
        doc.status = DocumentStatus.FAILED.value
        doc.error_message = str(e)
        await db.flush()
        raise HTTPException(status_code=500, detail=f"OCR processing failed: {e}")


@router.get("/{doc_id}/result", response_model=OCRResponse)
async def get_ocr_result(doc_id: str, db: AsyncSession = Depends(get_db)):
    """Get OCR result for a document."""
    result = await db.execute(select(Document).where(Document.id == doc_id))
    doc = result.scalar_one_or_none()
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")

    if doc.status == DocumentStatus.UPLOADED.value:
        raise HTTPException(status_code=400, detail="OCR not yet processed")

    return OCRResponse(
        doc_id=doc_id,
        text=doc.ocr_text or "",
        confidence=doc.ocr_confidence or 0.0,
        provider="unknown",
        status=doc.status,
    )
