"""
TDD tests for LLM extraction endpoints.
"""
import io
import pytest


class TestExtraction:
    async def _upload_and_ocr(self, client, content=b"Extraction test content", filename="doc.txt"):
        """Helper: upload + run OCR, return doc_id."""
        files = {"file": (filename, io.BytesIO(content), "text/plain")}
        resp = await client.post("/api/v1/documents/upload", files=files)
        assert resp.status_code == 201
        doc_id = resp.json()["doc_id"]
        # Run OCR first
        await client.post(f"/api/v1/ocr/process/{doc_id}")
        return doc_id

    async def test_extract_invoice(self, client):
        """Invoice text extraction returns structured JSON with key fields."""
        invoice_text = (
            b"INVOICE #001\n"
            b"Vendor: Cong Ty TNHH ABC\n"
            b"Tax ID: 0123456789\n"
            b"Date: 2024-01-15\n"
            b"Total: 5,000,000 VND\n"
        )
        doc_id = await self._upload_and_ocr(client, invoice_text, "invoice.txt")

        response = await client.post(
            f"/api/v1/extract/process/{doc_id}",
            json={"doc_type": "invoice"},
        )
        assert response.status_code == 200
        data = response.json()
        assert "extracted_data" in data
        assert isinstance(data["extracted_data"], dict)
        assert data["status"] in ("extracted", "extracting")

    async def test_extract_unknown_type(self, client):
        """Extraction with 'other' type returns basic fields."""
        doc_id = await self._upload_and_ocr(client, b"Some document with info: date 2024-03-01.", "other.txt")

        response = await client.post(
            f"/api/v1/extract/process/{doc_id}",
            json={"doc_type": "other"},
        )
        assert response.status_code == 200
        data = response.json()
        assert "extracted_data" in data
        assert isinstance(data["extracted_data"], dict)

    async def test_extraction_updates_status(self, client):
        """After extraction, document status changes to extracted."""
        doc_id = await self._upload_and_ocr(client, b"Contract between Party A and Party B.", "contract.txt")

        await client.post(
            f"/api/v1/extract/process/{doc_id}",
            json={"doc_type": "contract"},
        )

        doc_resp = await client.get(f"/api/v1/documents/{doc_id}")
        doc = doc_resp.json()
        assert doc["status"] in ("extracted", "extracting", "indexing", "indexed")

    async def test_extraction_result_endpoint(self, client):
        """GET /extract/{doc_id}/result returns extracted data."""
        doc_id = await self._upload_and_ocr(client, b"CV: John Doe, Software Engineer.", "cv.txt")

        await client.post(
            f"/api/v1/extract/process/{doc_id}",
            json={"doc_type": "cv"},
        )

        response = await client.get(f"/api/v1/extract/{doc_id}/result")
        assert response.status_code == 200
        data = response.json()
        assert "extracted_data" in data
        assert "doc_id" in data

    async def test_extraction_nonexistent_doc(self, client):
        """Extraction on unknown doc → 404."""
        response = await client.post(
            "/api/v1/extract/process/00000000-0000-0000-0000-000000000099",
            json={"doc_type": "invoice"},
        )
        assert response.status_code == 404
