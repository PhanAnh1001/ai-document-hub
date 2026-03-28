"""
Evaluation endpoints: run metrics, get aggregate stats.
"""
import logging
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from app.database import get_db
from app.config import settings
from app.models.db_models import Document
from app.models.schemas import EvalResponse, EvalStatsResponse, EvalRunRequest
from app.services.evaluation_service import EvaluationService
from app.services.rag_service import RAGService
from app.services.llm_client import LLMClient

logger = logging.getLogger(__name__)
router = APIRouter()

# In-memory eval history (simple, no DB table for evals in MVP)
_eval_history: list[dict] = []


def get_eval_service() -> EvaluationService:
    llm = LLMClient(settings)
    rag = RAGService(llm, settings)
    return EvaluationService(llm, rag)


@router.get("/stats", response_model=EvalStatsResponse)
async def get_eval_stats():
    """Get aggregate evaluation statistics."""
    if not _eval_history:
        return EvalStatsResponse(
            total_evaluations=0,
            faithfulness=0.0,
            answer_relevancy=0.0,
            context_precision=0.0,
            extraction_accuracy=None,
        )

    n = len(_eval_history)
    avg_faith = sum(e["faithfulness"] for e in _eval_history) / n
    avg_rel = sum(e["answer_relevancy"] for e in _eval_history) / n
    avg_prec = sum(e["context_precision"] for e in _eval_history) / n

    acc_values = [e["extraction_accuracy"] for e in _eval_history if e.get("extraction_accuracy") is not None]
    avg_acc = sum(acc_values) / len(acc_values) if acc_values else None

    return EvalStatsResponse(
        total_evaluations=n,
        faithfulness=round(avg_faith, 4),
        answer_relevancy=round(avg_rel, 4),
        context_precision=round(avg_prec, 4),
        extraction_accuracy=round(avg_acc, 4) if avg_acc is not None else None,
    )


@router.post("/run", response_model=EvalResponse)
async def run_evaluation(
    request: EvalRunRequest,
    db: AsyncSession = Depends(get_db),
    eval_service: EvaluationService = Depends(get_eval_service),
):
    """Run evaluation metrics on a document."""
    # Verify document exists
    result = await db.execute(select(Document).where(Document.id == request.doc_id))
    doc = result.scalar_one_or_none()
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")

    try:
        metrics = await eval_service.evaluate_document(
            request.doc_id,
            request.question,
            request.expected_answer,
            db,
        )
        _eval_history.append(metrics)

        from datetime import datetime
        return EvalResponse(
            faithfulness=metrics["faithfulness"],
            answer_relevancy=metrics["answer_relevancy"],
            context_precision=metrics["context_precision"],
            extraction_accuracy=metrics.get("extraction_accuracy"),
            doc_id=request.doc_id,
            evaluated_at=datetime.utcnow(),
        )
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Evaluation failed: {e}")
        raise HTTPException(status_code=500, detail=f"Evaluation failed: {e}")
