"""
RAG service: chunk + embed documents, retrieve relevant chunks, generate answers.
Uses BGE-M3 embeddings (or mock), stores in DB as JSON arrays.
"""
import json
import logging
import math
from typing import Any

logger = logging.getLogger(__name__)

# Chunk size in characters
CHUNK_SIZE = 500
CHUNK_OVERLAP = 50


class RAGService:
    def __init__(self, llm_client, settings):
        self.llm = llm_client
        self.settings = settings
        self._embedder = None

    def _get_embedder(self):
        """Lazy-load sentence-transformers model."""
        if self._embedder is None:
            try:
                from sentence_transformers import SentenceTransformer
                self._embedder = SentenceTransformer(self.settings.EMBEDDING_MODEL)
                logger.info(f"Loaded embedding model: {self.settings.EMBEDDING_MODEL}")
            except (ImportError, Exception) as e:
                logger.warning(f"Could not load embedder: {e}, using mock embeddings")
        return self._embedder

    def _embed_text(self, text: str) -> list[float]:
        """Embed text; fall back to deterministic mock if no model."""
        embedder = self._get_embedder()
        if embedder:
            try:
                vec = embedder.encode(text, normalize_embeddings=True)
                return vec.tolist()
            except Exception as e:
                logger.warning(f"Embedding failed: {e}")
        return self._mock_embed(text)

    def _mock_embed(self, text: str) -> list[float]:
        """Deterministic mock embedding based on character hash."""
        dim = self.settings.EMBEDDING_DIM
        vec = [0.0] * dim
        for i, char in enumerate(text[:dim]):
            vec[i % dim] += ord(char) / 1000.0
        # Normalize
        magnitude = math.sqrt(sum(x * x for x in vec)) or 1.0
        return [x / magnitude for x in vec]

    def _chunk_text(self, text: str) -> list[str]:
        """Split text into overlapping chunks."""
        if not text:
            return []
        chunks = []
        start = 0
        while start < len(text):
            end = start + CHUNK_SIZE
            chunks.append(text[start:end])
            start += CHUNK_SIZE - CHUNK_OVERLAP
            if start >= len(text):
                break
        return chunks

    def _cosine_similarity(self, a: list[float], b: list[float]) -> float:
        """Compute cosine similarity between two vectors."""
        if not a or not b or len(a) != len(b):
            return 0.0
        dot = sum(x * y for x, y in zip(a, b))
        mag_a = math.sqrt(sum(x * x for x in a))
        mag_b = math.sqrt(sum(x * x for x in b))
        if mag_a == 0 or mag_b == 0:
            return 0.0
        return dot / (mag_a * mag_b)

    async def index_document(self, doc_id: str, text: str, db_session) -> list[dict]:
        """
        Chunk text, embed chunks, store in document_chunks table.
        Stores embedding in both embedding_json (always) and embedding (pgvector, if available).
        Returns list of chunk dicts.
        """
        from app.models.db_models import DocumentChunk, PGVECTOR_AVAILABLE

        chunks = self._chunk_text(text)
        if not chunks:
            return []

        chunk_records = []
        for i, chunk_text in enumerate(chunks):
            embedding = self._embed_text(chunk_text)
            chunk_data: dict = {
                "document_id": doc_id,
                "chunk_index": float(i),
                "chunk_text": chunk_text,
                "embedding_json": embedding,
            }
            # Store in native pgvector column when available
            if PGVECTOR_AVAILABLE:
                chunk_data["embedding"] = embedding  # pgvector handles list→vector conversion

            chunk = DocumentChunk(**chunk_data)
            db_session.add(chunk)
            chunk_records.append({
                "chunk_index": i,
                "chunk_text": chunk_text,
                "embedding": embedding,
            })

        await db_session.flush()
        return chunk_records

    async def query(
        self, question: str, doc_ids: list[str] | None, db_session, top_k: int = 3
    ) -> dict[str, Any]:
        """
        Embed question, find top-k relevant chunks, generate answer with LLM.
        Returns dict with answer, sources, question.
        """
        from sqlalchemy import select
        from app.models.db_models import DocumentChunk

        # Retrieve candidate chunks
        stmt = select(DocumentChunk)
        if doc_ids:
            stmt = stmt.where(DocumentChunk.document_id.in_(doc_ids))
        result = await db_session.execute(stmt)
        all_chunks = result.scalars().all()

        if not all_chunks:
            answer = "Tôi không tìm thấy thông tin này trong tài liệu."
            return {"answer": answer, "sources": [], "question": question}

        # Rank by cosine similarity
        question_vec = self._embed_text(question)
        ranked = []
        for chunk in all_chunks:
            if chunk.embedding_json:
                sim = self._cosine_similarity(question_vec, chunk.embedding_json)
            else:
                sim = 0.0
            ranked.append((sim, chunk))

        ranked.sort(key=lambda x: x[0], reverse=True)
        top_chunks = ranked[:top_k]

        # Build context
        context_parts = [chunk.chunk_text for _, chunk in top_chunks]
        context = "\n\n---\n\n".join(context_parts)

        from app.models.prompts import RAG_GENERATION_PROMPT
        prompt = RAG_GENERATION_PROMPT.format(context=context, question=question)
        answer = await self.llm.complete(prompt)

        sources = [
            {
                "chunk_id": str(chunk.id),
                "document_id": str(chunk.document_id),
                "text_snippet": chunk.chunk_text[:200],
                "chunk_index": int(chunk.chunk_index),
            }
            for _, chunk in top_chunks
        ]

        return {"answer": answer, "sources": sources, "question": question}
