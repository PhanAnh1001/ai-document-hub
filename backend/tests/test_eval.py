"""
TDD tests for evaluation endpoints.
"""
import io
import pytest


class TestEval:
    async def test_eval_stats_empty(self, client):
        """GET /eval/stats with no evaluations returns zero stats."""
        response = await client.get("/api/v1/eval/stats")
        assert response.status_code == 200
        data = response.json()
        assert "total_evaluations" in data
        assert data["total_evaluations"] >= 0

    async def test_eval_stats_format(self, client):
        """Eval stats response has required fields."""
        response = await client.get("/api/v1/eval/stats")
        assert response.status_code == 200
        data = response.json()
        assert "faithfulness" in data
        assert "answer_relevancy" in data
        assert "context_precision" in data

    async def test_eval_run_returns_metrics(self, client):
        """Running eval on a document returns metric scores."""
        # Upload and prepare a document
        content = b"The vendor is ABC Corp. Total: 2,000,000 VND."
        files = {"file": ("eval_test.txt", io.BytesIO(content), "text/plain")}
        upload_resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = upload_resp.json()["doc_id"]
        await client.post(f"/api/v1/ocr/process/{doc_id}")

        response = await client.post(
            "/api/v1/eval/run",
            json={
                "doc_id": doc_id,
                "question": "Who is the vendor?",
                "expected_answer": "ABC Corp",
            },
        )
        assert response.status_code == 200
        data = response.json()
        assert "faithfulness" in data
        assert "answer_relevancy" in data
        assert "context_precision" in data
        assert 0.0 <= data["faithfulness"] <= 1.0
        assert 0.0 <= data["answer_relevancy"] <= 1.0

    async def test_eval_run_nonexistent_doc(self, client):
        """Eval run on unknown doc → 404."""
        response = await client.post(
            "/api/v1/eval/run",
            json={"doc_id": "00000000-0000-0000-0000-000000000099"},
        )
        assert response.status_code == 404
