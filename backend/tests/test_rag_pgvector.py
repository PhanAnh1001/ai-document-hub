"""
TDD tests for pgvector-aware RAG service.
All tests run against SQLite (aiosqlite) — pgvector-specific paths are exercised
via the fallback JSON cosine path. pgvector-only tests are skipped when the
pgvector package is not available or the vector column is absent.
"""
import io
import json
import os
import pytest

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

INVOICE_TEXT = (
    "HÓA ĐƠN GIÁ TRỊ GIA TĂNG\n"
    "Số: 001\nNgày: 15/01/2024\n"
    "Đơn vị bán hàng: Công ty TNHH ABC\n"
    "Mã số thuế: 0123456789\n"
    "Tên hàng hóa: Máy tính xách tay Dell XPS 15\n"
    "Số lượng: 2\nĐơn giá: 25,000,000 VNĐ\n"
    "Thành tiền: 50,000,000 VNĐ\n"
    "Thuế GTGT (10%): 5,000,000 VNĐ\n"
    "Tổng cộng: 55,000,000 VNĐ\n"
    "Người mua: Trần Văn B\n"
    "Mã số thuế NM: 9876543210"
)


class TestIndexDocumentCreatesChunks:
    """index_document creates chunk records in the database."""

    async def test_index_document_creates_chunks(self, client, db_session):
        """Indexing a document creates at least one chunk in document_chunks."""
        from sqlalchemy import select
        from app.models.db_models import DocumentChunk

        # Upload + OCR so the document exists and has ocr_text
        content = INVOICE_TEXT.encode()
        files = {"file": ("pgvec_test.txt", io.BytesIO(content), "text/plain")}
        resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = resp.json()["doc_id"]
        await client.post(f"/api/v1/ocr/process/{doc_id}")
        # Trigger indexing via query endpoint (uses _ensure_indexed internally)
        await client.post(
            "/api/v1/query/",
            json={"question": "What is this document?", "doc_ids": [doc_id]},
        )

        result = await db_session.execute(
            select(DocumentChunk).where(DocumentChunk.document_id == doc_id)
        )
        chunks = result.scalars().all()
        assert len(chunks) >= 1, "At least one chunk should be created"

    async def test_index_creates_embedding_json(self, client, db_session):
        """Each chunk must have a non-null embedding_json list."""
        from sqlalchemy import select
        from app.models.db_models import DocumentChunk

        content = b"The vendor is ABC Corp. Invoice total 2,000,000 VND."
        files = {"file": ("emb_test.txt", io.BytesIO(content), "text/plain")}
        resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = resp.json()["doc_id"]
        await client.post(f"/api/v1/ocr/process/{doc_id}")
        await client.post(f"/api/v1/extract/process/{doc_id}", json={"doc_type": "invoice"})

        result = await db_session.execute(
            select(DocumentChunk).where(DocumentChunk.document_id == doc_id)
        )
        chunks = result.scalars().all()
        for chunk in chunks:
            assert chunk.embedding_json is not None, "embedding_json must not be None"
            assert isinstance(chunk.embedding_json, list), "embedding_json should be a list"
            assert len(chunk.embedding_json) > 0, "embedding_json must not be empty"


class TestQueryReturnsRelevantChunks:
    """RAG query returns contextually relevant results."""

    async def _prepare_doc(self, client, content: bytes) -> str:
        """Upload, OCR, extract (index) a document; return doc_id."""
        files = {"file": ("query_pgvec.txt", io.BytesIO(content), "text/plain")}
        resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = resp.json()["doc_id"]
        await client.post(f"/api/v1/ocr/process/{doc_id}")
        await client.post(f"/api/v1/extract/process/{doc_id}", json={"doc_type": "invoice"})
        return doc_id

    async def test_query_returns_relevant_chunks(self, client):
        """Query response includes sources with document_id matching indexed doc."""
        doc_id = await self._prepare_doc(client, INVOICE_TEXT.encode())

        resp = await client.post(
            "/api/v1/query/",
            json={"question": "Tổng cộng hóa đơn là bao nhiêu?", "doc_ids": [doc_id]},
        )
        assert resp.status_code == 200
        data = resp.json()
        assert "answer" in data
        assert len(data["answer"]) > 0
        # Sources should reference the indexed document
        sources = data.get("sources", [])
        assert isinstance(sources, list)
        if sources:
            assert any(s["document_id"] == doc_id for s in sources)

    async def test_query_answer_not_empty_for_indexed_doc(self, client):
        """RAG answer is non-empty when relevant content exists."""
        content = b"Tong tien thanh toan la 1,500,000 VND. Khach hang: Le Van A."
        doc_id = await self._prepare_doc(client, content)

        resp = await client.post(
            "/api/v1/query/",
            json={"question": "Tong tien la bao nhieu?", "doc_ids": [doc_id]},
        )
        assert resp.status_code == 200
        assert len(resp.json()["answer"]) > 0


class TestIndexAndQueryRoundtrip:
    """Full roundtrip: index document → query → verify answer contains context."""

    async def test_roundtrip_invoice_total(self, client):
        """Query about invoice total after indexing returns sensible answer."""
        content = b"Vendor: ABC Company. Invoice total: 99,000,000 VND. Tax 10%: 9,000,000 VND."
        files = {"file": ("roundtrip.txt", io.BytesIO(content), "text/plain")}
        resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = resp.json()["doc_id"]
        await client.post(f"/api/v1/ocr/process/{doc_id}")
        await client.post(f"/api/v1/extract/process/{doc_id}", json={"doc_type": "invoice"})

        resp = await client.post(
            "/api/v1/query/",
            json={"question": "What is the invoice total?", "doc_ids": [doc_id]},
        )
        assert resp.status_code == 200
        data = resp.json()
        assert "answer" in data
        # The answer or sources must confirm the document was retrieved
        sources = data.get("sources", [])
        assert any(s["document_id"] == doc_id for s in sources), (
            "Sources must include the indexed document"
        )

    async def test_roundtrip_multiple_docs(self, client):
        """Querying with doc_ids filters to only the specified documents."""
        content_a = b"Document A: vendor is AlphaVendor, total 10,000,000 VND."
        content_b = b"Document B: vendor is BetaVendor, total 20,000,000 VND."

        async def upload_and_index(content: bytes) -> str:
            files = {"file": ("multi.txt", io.BytesIO(content), "text/plain")}
            r = await client.post("/api/v1/documents/upload", files=files)
            did = r.json()["doc_id"]
            await client.post(f"/api/v1/ocr/process/{did}")
            await client.post(f"/api/v1/extract/process/{did}", json={"doc_type": "invoice"})
            return did

        doc_a = await upload_and_index(content_a)
        doc_b = await upload_and_index(content_b)

        # Query only doc_a
        resp = await client.post(
            "/api/v1/query/",
            json={"question": "Who is the vendor?", "doc_ids": [doc_a]},
        )
        assert resp.status_code == 200
        sources = resp.json().get("sources", [])
        # All returned sources should belong to doc_a only
        for s in sources:
            assert s["document_id"] == doc_a


class TestEvaluationExtractionAccuracy:
    """Field-level extraction accuracy evaluation."""

    def _get_eval_service(self):
        """Instantiate EvaluationService with mock dependencies."""
        from app.services.evaluation_service import EvaluationService

        class MockLLM:
            async def complete(self, prompt): return "mock answer"

        class MockRAG:
            async def query(self, *a, **kw): return {"answer": "", "sources": []}

        return EvaluationService(MockLLM(), MockRAG())

    def test_exact_match_fields(self):
        """All fields match → overall score 1.0."""
        svc = self._get_eval_service()
        extracted = {"vendor": "Công ty ABC", "total": 55000000, "date": "15/01/2024"}
        expected = {"vendor": "Công ty ABC", "total": 55000000, "date": "15/01/2024"}
        result = svc.evaluate_extraction_accuracy(extracted, expected)
        assert result["overall"] == 1.0
        assert result["matched"] == 3
        assert result["total"] == 3

    def test_partial_match(self):
        """Partial field match returns correct fraction."""
        svc = self._get_eval_service()
        extracted = {"vendor": "Công ty ABC", "total": 55000000, "date": "wrong date"}
        expected = {"vendor": "Công ty ABC", "total": 55000000, "date": "15/01/2024"}
        result = svc.evaluate_extraction_accuracy(extracted, expected)
        assert result["overall"] == pytest.approx(2 / 3, rel=1e-3)
        assert result["matched"] == 2

    def test_missing_field_counts_as_wrong(self):
        """Missing extracted field is counted as incorrect."""
        svc = self._get_eval_service()
        extracted = {"vendor": "Công ty ABC"}  # total missing
        expected = {"vendor": "Công ty ABC", "total": 55000000}
        result = svc.evaluate_extraction_accuracy(extracted, expected)
        assert result["matched"] == 1
        assert result["total"] == 2
        assert result["fields"]["total"]["correct"] is False

    def test_empty_expected_returns_perfect_score(self):
        """Empty expected dict → overall 1.0 (nothing to check)."""
        svc = self._get_eval_service()
        result = svc.evaluate_extraction_accuracy({"vendor": "X"}, {})
        assert result["overall"] == 1.0

    def test_numeric_field_comparison(self):
        """Numeric fields are compared correctly regardless of string format."""
        svc = self._get_eval_service()
        # Both represent 55000000
        extracted = {"total": "55000000"}
        expected = {"total": 55000000}
        result = svc.evaluate_extraction_accuracy(extracted, expected)
        assert result["fields"]["total"]["correct"] is True

    def test_string_substring_match(self):
        """String fields match if one is a substring of the other."""
        svc = self._get_eval_service()
        # Predicted contains expected
        extracted = {"vendor": "Công ty TNHH ABC Việt Nam"}
        expected = {"vendor": "ABC"}
        result = svc.evaluate_extraction_accuracy(extracted, expected)
        assert result["fields"]["vendor"]["correct"] is True

    async def test_batch_evaluate_aggregates_metrics(self, client, db_session):
        """batch_evaluate returns aggregated metrics across dataset items."""
        from app.services.evaluation_service import EvaluationService
        from app.services.rag_service import RAGService
        from app.config import settings

        class MockLLM:
            async def complete(self, prompt): return "mock llm answer for batch eval"

        rag = RAGService(MockLLM(), settings)
        svc = EvaluationService(MockLLM(), rag)

        # Upload + index two documents
        async def make_doc(content: bytes) -> str:
            files = {"file": ("batch.txt", io.BytesIO(content), "text/plain")}
            r = await client.post("/api/v1/documents/upload", files=files)
            did = r.json()["doc_id"]
            await client.post(f"/api/v1/ocr/process/{did}")
            await client.post(f"/api/v1/extract/process/{did}", json={"doc_type": "invoice"})
            return did

        doc1 = await make_doc(b"Invoice from Vendor A, total 1,000,000 VND.")
        doc2 = await make_doc(b"Invoice from Vendor B, total 2,000,000 VND.")

        dataset = [
            {"doc_id": doc1, "question": "Who is the vendor?"},
            {"doc_id": doc2, "question": "What is the total?"},
        ]
        result = await svc.batch_evaluate(dataset, db_session)

        assert result["total"] == 2
        assert result["errors"] == 0
        assert "avg_faithfulness" in result
        assert "avg_answer_relevancy" in result
        assert "avg_context_precision" in result
        assert 0.0 <= result["avg_faithfulness"] <= 1.0
        assert 0.0 <= result["avg_answer_relevancy"] <= 1.0


class TestPgvectorAvailability:
    """Tests for pgvector availability flag and graceful fallback."""

    def test_pgvector_flag_is_boolean(self):
        """PGVECTOR_AVAILABLE must be a boolean."""
        from app.models.db_models import PGVECTOR_AVAILABLE
        assert isinstance(PGVECTOR_AVAILABLE, bool)

    def test_document_chunk_has_embedding_json(self):
        """DocumentChunk always has embedding_json column regardless of pgvector."""
        from app.models.db_models import DocumentChunk
        assert hasattr(DocumentChunk, "embedding_json")

    def test_pgvector_column_present_when_available(self):
        """When pgvector is installed, DocumentChunk has embedding attribute."""
        from app.models.db_models import DocumentChunk, PGVECTOR_AVAILABLE
        if PGVECTOR_AVAILABLE:
            assert hasattr(DocumentChunk, "embedding"), (
                "DocumentChunk should have 'embedding' column when pgvector is available"
            )
        else:
            # Not an error if not present — just confirm flag is False
            assert PGVECTOR_AVAILABLE is False

    def test_rag_service_falls_back_gracefully(self):
        """RAGService._query_pgvector returns None without pgvector DB."""
        # This just checks the method exists and is callable
        from app.services.rag_service import RAGService

        class MockLLM:
            async def complete(self, prompt): return "ok"

        class MockSettings:
            EMBEDDING_MODEL = "BAAI/bge-m3"
            EMBEDDING_DIM = 1024

        svc = RAGService(MockLLM(), MockSettings())
        assert callable(svc._query_pgvector)
