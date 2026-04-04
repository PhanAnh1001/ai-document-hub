"""
Document CRUD endpoints: upload, list, get, delete.
"""
import os
import uuid
import logging
from typing import Optional

import aiofiles
from fastapi import APIRouter, Depends, HTTPException, UploadFile, File, BackgroundTasks, Query
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func

from app.database import get_db
from app.config import settings
from app.models.db_models import Document, DocumentStatus
from app.models.schemas import DocumentUploadResponse, DocumentResponse, DocumentListResponse
from app.worker import process_document_background

logger = logging.getLogger(__name__)
router = APIRouter()


@router.post("/upload", response_model=DocumentUploadResponse, status_code=201)
async def upload_document(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    db: AsyncSession = Depends(get_db),
):
    """Upload a document file and trigger background processing."""
    # Read file content to check size
    content = await file.read()
    file_size = len(content)

    if file_size > settings.MAX_UPLOAD_SIZE:
        raise HTTPException(status_code=413, detail="File too large. Max size is 10MB.")

    # Ensure upload directory exists
    os.makedirs(settings.UPLOAD_DIR, exist_ok=True)

    # Generate unique filename
    doc_id = str(uuid.uuid4())
    ext = os.path.splitext(file.filename or "")[1] or ".bin"
    saved_filename = f"{doc_id}{ext}"
    file_path = os.path.join(settings.UPLOAD_DIR, saved_filename)

    # Save file to disk
    async with aiofiles.open(file_path, "wb") as f:
        await f.write(content)

    # Determine MIME type
    mime_type = file.content_type or "application/octet-stream"

    # Create DB record
    doc = Document(
        id=doc_id,
        user_id="00000000-0000-0000-0000-000000000001",  # default user (auth not required for MVP)
        filename=saved_filename,
        original_filename=file.filename or saved_filename,
        file_path=file_path,
        file_size=float(file_size),
        mime_type=mime_type,
        status=DocumentStatus.UPLOADED.value,
    )
    db.add(doc)
    await db.flush()

    # Queue background processing
    background_tasks.add_task(process_document_background, doc_id, file_path, mime_type)

    return DocumentUploadResponse(
        doc_id=doc_id,
        filename=saved_filename,
        status=DocumentStatus.UPLOADED.value,
    )


@router.get("/", response_model=DocumentListResponse)
async def list_documents(
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    status: Optional[str] = Query(None),
    db: AsyncSession = Depends(get_db),
):
    """List documents with optional status filter and pagination."""
    stmt = select(Document)
    if status:
        stmt = stmt.where(Document.status == status)

    # Count total
    count_stmt = select(func.count()).select_from(stmt.subquery())
    total_result = await db.execute(count_stmt)
    total = total_result.scalar() or 0

    # Paginate
    stmt = stmt.offset((page - 1) * page_size).limit(page_size)
    result = await db.execute(stmt)
    docs = result.scalars().all()

    return DocumentListResponse(
        items=[DocumentResponse.model_validate(doc) for doc in docs],
        total=total,
        page=page,
        page_size=page_size,
    )


@router.get("/{doc_id}", response_model=DocumentResponse)
async def get_document(doc_id: str, db: AsyncSession = Depends(get_db)):
    """Get document details by ID."""
    result = await db.execute(select(Document).where(Document.id == doc_id))
    doc = result.scalar_one_or_none()
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")
    return DocumentResponse.model_validate(doc)


@router.delete("/{doc_id}", status_code=204)
async def delete_document(doc_id: str, db: AsyncSession = Depends(get_db)):
    """Delete a document and its associated file."""
    result = await db.execute(select(Document).where(Document.id == doc_id))
    doc = result.scalar_one_or_none()
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")

    # Remove file from disk (best-effort)
    if doc.file_path and os.path.exists(doc.file_path):
        try:
            os.remove(doc.file_path)
        except OSError as e:
            logger.warning(f"Could not delete file {doc.file_path}: {e}")

    await db.delete(doc)
    await db.flush()
