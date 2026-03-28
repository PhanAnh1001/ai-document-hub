"""
RAG Q&A endpoints: ask questions, get answers with citations.
"""
import logging
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, desc

from app.database import get_db
from app.config import settings
from app.models.db_models import Document, DocumentChunk, ChatHistory
from app.models.schemas import QueryRequest, QueryResponse, Citation, ChatHistoryItem
from app.services.rag_service import RAGService
from app.services.llm_client import LLMClient

logger = logging.getLogger(__name__)
router = APIRouter()


def get_rag_service() -> RAGService:
    llm = LLMClient(settings)
    return RAGService(llm, settings)


@router.post("/", response_model=QueryResponse)
async def query_documents(
    request: QueryRequest,
    db: AsyncSession = Depends(get_db),
    rag_service: RAGService = Depends(get_rag_service),
):
    """Ask a question and get a RAG-powered answer with citations."""
    # Ensure documents are indexed (auto-index docs with OCR text if not yet chunked)
    await _ensure_indexed(request.doc_ids, db, rag_service)

    result = await rag_service.query(request.question, request.doc_ids, db)

    # Save to chat history (best-effort, no user auth yet)
    try:
        history = ChatHistory(
            user_id="00000000-0000-0000-0000-000000000001",
            question=request.question,
            answer=result["answer"],
            context_chunks=[s["chunk_id"] for s in result["sources"]],
        )
        db.add(history)
        await db.flush()
    except Exception as e:
        logger.warning(f"Could not save chat history: {e}")

    sources = [
        Citation(
            chunk_id=s["chunk_id"],
            document_id=s["document_id"],
            text_snippet=s["text_snippet"],
            chunk_index=s["chunk_index"],
        )
        for s in result["sources"]
    ]

    return QueryResponse(
        answer=result["answer"],
        sources=sources,
        question=result["question"],
    )


@router.get("/history", response_model=list[ChatHistoryItem])
async def get_chat_history(
    limit: int = 20,
    db: AsyncSession = Depends(get_db),
):
    """Get recent chat history."""
    result = await db.execute(
        select(ChatHistory)
        .order_by(desc(ChatHistory.created_at))
        .limit(limit)
    )
    items = result.scalars().all()
    return [ChatHistoryItem.model_validate(item) for item in items]


async def _ensure_indexed(doc_ids: list[str] | None, db: AsyncSession, rag_service: RAGService):
    """Auto-index documents that have OCR text but no chunks yet."""
    if doc_ids:
        stmt = select(Document).where(Document.id.in_(doc_ids))
    else:
        stmt = select(Document)

    result = await db.execute(stmt)
    docs = result.scalars().all()

    for doc in docs:
        if not doc.ocr_text:
            continue
        # Check if already chunked
        chunk_count = await db.execute(
            select(DocumentChunk).where(DocumentChunk.document_id == doc.id).limit(1)
        )
        if chunk_count.first() is None:
            await rag_service.index_document(str(doc.id), doc.ocr_text, db)
