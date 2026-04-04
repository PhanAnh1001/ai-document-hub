"""
Evaluation service: compute RAG metrics (RAGAS-style) and field-level extraction accuracy.
Uses heuristic metrics when RAGAS is not available.
"""
import logging
import re
from typing import Any

logger = logging.getLogger(__name__)


class EvaluationService:
    def __init__(self, llm_client, rag_service):
        self.llm = llm_client
        self.rag = rag_service

    async def evaluate_document(
        self,
        doc_id: str,
        question: str | None,
        expected_answer: str | None,
        db_session,
    ) -> dict[str, Any]:
        """
        Run evaluation on a document.
        Returns faithfulness, answer_relevancy, context_precision, extraction_accuracy.
        """
        from sqlalchemy import select
        from app.models.db_models import Document, DocumentChunk

        # Fetch document
        result = await db_session.execute(
            select(Document).where(Document.id == doc_id)
        )
        doc = result.scalar_one_or_none()
        if doc is None:
            raise ValueError(f"Document {doc_id} not found")

        # Default question if not provided
        if not question:
            question = "What is this document about?"

        # Run RAG query
        rag_result = await self.rag.query(question, [doc_id], db_session)
        answer = rag_result.get("answer", "")
        sources = rag_result.get("sources", [])

        # Compute heuristic metrics
        faithfulness = self._compute_faithfulness(answer, sources)
        answer_relevancy = self._compute_answer_relevancy(answer, question)
        context_precision = self._compute_context_precision(sources, question)

        # Extraction accuracy if expected answer provided
        extraction_accuracy = None
        if expected_answer:
            extraction_accuracy = self._compute_extraction_accuracy(answer, expected_answer)

        return {
            "faithfulness": faithfulness,
            "answer_relevancy": answer_relevancy,
            "context_precision": context_precision,
            "extraction_accuracy": extraction_accuracy,
            "doc_id": doc_id,
        }

    def _compute_faithfulness(self, answer: str, sources: list[dict]) -> float:
        """
        Heuristic: check if answer words appear in source chunks.
        Score = fraction of answer words found in context.
        """
        if not answer or not sources:
            return 0.5

        context = " ".join(s.get("text_snippet", "") for s in sources).lower()
        answer_words = set(re.findall(r"\w+", answer.lower()))
        if not answer_words:
            return 0.5

        # Exclude common stop words
        stop_words = {"the", "a", "an", "is", "are", "was", "were", "in", "on", "at",
                      "to", "for", "of", "and", "or", "but", "tôi", "không", "có", "này"}
        meaningful_words = answer_words - stop_words
        if not meaningful_words:
            return 0.8

        found = sum(1 for w in meaningful_words if w in context)
        return min(1.0, found / len(meaningful_words))

    def _compute_answer_relevancy(self, answer: str, question: str) -> float:
        """
        Heuristic: check if question keywords appear in answer.
        """
        if not answer or not question:
            return 0.5

        # Special case: "no info found" answer
        no_info_phrases = ["không tìm thấy", "no information", "i don't know", "not found"]
        if any(phrase in answer.lower() for phrase in no_info_phrases):
            return 0.3

        question_words = set(re.findall(r"\w+", question.lower()))
        stop_words = {"what", "who", "when", "where", "how", "is", "the", "a", "an", "?"}
        q_keywords = question_words - stop_words
        if not q_keywords:
            return 0.7

        answer_lower = answer.lower()
        found = sum(1 for w in q_keywords if w in answer_lower)
        return min(1.0, 0.3 + 0.7 * found / len(q_keywords))

    def _compute_context_precision(self, sources: list[dict], question: str) -> float:
        """
        Heuristic: fraction of retrieved chunks relevant to question.
        """
        if not sources:
            return 0.0

        q_words = set(re.findall(r"\w+", question.lower()))
        relevant = 0
        for source in sources:
            snippet = source.get("text_snippet", "").lower()
            chunk_words = set(re.findall(r"\w+", snippet))
            if q_words & chunk_words:  # intersection
                relevant += 1

        return relevant / len(sources)

    def _compute_extraction_accuracy(self, answer: str, expected: str) -> float:
        """
        Simple token-level F1 between answer and expected.
        """
        pred_tokens = set(re.findall(r"\w+", answer.lower()))
        gold_tokens = set(re.findall(r"\w+", expected.lower()))
        if not gold_tokens:
            return 1.0 if not pred_tokens else 0.0
        if not pred_tokens:
            return 0.0
        tp = len(pred_tokens & gold_tokens)
        precision = tp / len(pred_tokens)
        recall = tp / len(gold_tokens)
        if precision + recall == 0:
            return 0.0
        f1 = 2 * precision * recall / (precision + recall)
        return round(f1, 4)

    def get_empty_stats(self) -> dict:
        """Return zero stats when no evaluations exist."""
        return {
            "total_evaluations": 0,
            "faithfulness": 0.0,
            "answer_relevancy": 0.0,
            "context_precision": 0.0,
            "extraction_accuracy": None,
        }
