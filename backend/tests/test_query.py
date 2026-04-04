"""
TDD tests for RAG Q&A endpoints.
"""
import io
import pytest


class TestQuery:
    async def _prepare_indexed_doc(self, client, content=b"The total amount is 1,000,000 VND."):
        """Helper: upload, OCR, extract, index a document."""
        files = {"file": ("rag_test.txt", io.BytesIO(content), "text/plain")}
        resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = resp.json()["doc_id"]
        await client.post(f"/api/v1/ocr/process/{doc_id}")
        await client.post(f"/api/v1/extract/process/{doc_id}", json={"doc_type": "invoice"})
        return doc_id

    async def test_query_returns_answer(self, client):
        """Asking a question returns a non-empty answer."""
        doc_id = await self._prepare_indexed_doc(client)

        response = await client.post(
            "/api/v1/query/",
            json={"question": "What is the total amount?", "doc_ids": [doc_id]},
        )
        assert response.status_code == 200
        data = response.json()
        assert "answer" in data
        assert len(data["answer"]) > 0
        assert data["question"] == "What is the total amount?"

    async def test_query_includes_sources(self, client):
        """Query response includes sources/citations list."""
        doc_id = await self._prepare_indexed_doc(client)

        response = await client.post(
            "/api/v1/query/",
            json={"question": "Tell me about this document", "doc_ids": [doc_id]},
        )
        assert response.status_code == 200
        data = response.json()
        assert "sources" in data
        assert isinstance(data["sources"], list)

    async def test_query_empty_corpus(self, client):
        """Query with no indexed documents returns appropriate message."""
        response = await client.post(
            "/api/v1/query/",
            json={
                "question": "What documents exist?",
                "doc_ids": ["00000000-0000-0000-0000-000000000000"],
            },
        )
        assert response.status_code == 200
        data = response.json()
        assert "answer" in data
        # Should indicate no relevant info found
        assert len(data["answer"]) > 0

    async def test_query_history(self, client):
        """GET /query/history returns list of past queries."""
        # Make a query first
        await self._prepare_indexed_doc(client)
        await client.post(
            "/api/v1/query/",
            json={"question": "History test question?"},
        )

        response = await client.get("/api/v1/query/history")
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)
