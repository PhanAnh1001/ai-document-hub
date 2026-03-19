"""
TDD tests for OCR endpoints.
"""
import io
import pytest


class TestOCR:
    async def _upload_doc(self, client, content=b"OCR test content", filename="ocr.txt"):
        """Helper: upload a doc and return doc_id."""
        files = {"file": (filename, io.BytesIO(content), "text/plain")}
        resp = await client.post("/api/v1/documents/upload", files=files)
        assert resp.status_code == 201
        return resp.json()["doc_id"]

    async def test_ocr_text_file(self, client):
        """OCR on a text file returns text + confidence score."""
        doc_id = await self._upload_doc(client, b"Hello, this is OCR text.")
        response = await client.post(f"/api/v1/ocr/process/{doc_id}")
        assert response.status_code == 200
        data = response.json()
        assert "text" in data
        assert len(data["text"]) > 0
        assert "confidence" in data
        assert 0.0 <= data["confidence"] <= 1.0

    async def test_ocr_returns_provider(self, client):
        """OCR result includes provider field; in test env it should be 'mock'."""
        doc_id = await self._upload_doc(client, b"Provider test content.")
        response = await client.post(f"/api/v1/ocr/process/{doc_id}")
        assert response.status_code == 200
        data = response.json()
        assert "provider" in data
        # In test environment with no real OCR, provider is mock
        assert data["provider"] in ("paddle", "google_vision", "mock", "text")

    async def test_ocr_updates_document_status(self, client):
        """After OCR processing, document status should be ocr_done."""
        doc_id = await self._upload_doc(client, b"Status update test.")
        await client.post(f"/api/v1/ocr/process/{doc_id}")

        # Check document status updated
        doc_resp = await client.get(f"/api/v1/documents/{doc_id}")
        assert doc_resp.status_code == 200
        doc = doc_resp.json()
        assert doc["status"] in ("ocr_done", "ocr_processing", "extracting", "extracted", "indexed")

    async def test_ocr_result_endpoint(self, client):
        """GET /ocr/{doc_id}/result returns OCR data after processing."""
        doc_id = await self._upload_doc(client, b"Result endpoint test.")
        # First process
        await client.post(f"/api/v1/ocr/process/{doc_id}")
        # Then get result
        response = await client.get(f"/api/v1/ocr/{doc_id}/result")
        assert response.status_code == 200
        data = response.json()
        assert "text" in data
        assert "doc_id" in data

    async def test_ocr_nonexistent_doc(self, client):
        """OCR on unknown doc_id → 404."""
        response = await client.post("/api/v1/ocr/process/00000000-0000-0000-0000-000000000099")
        assert response.status_code == 404
