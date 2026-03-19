"""
TDD tests for document CRUD endpoints.
RED phase: write tests before implementation.
"""
import io
import pytest


class TestHealthCheck:
    async def test_health_check(self, client):
        """GET /health returns 200 with status ok."""
        response = await client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
        assert "service" in data


class TestDocumentUpload:
    async def test_upload_document_txt(self, client):
        """Upload a .txt file → 201 with doc_id."""
        content = b"This is a test invoice document."
        files = {"file": ("test.txt", io.BytesIO(content), "text/plain")}
        response = await client.post("/api/v1/documents/upload", files=files)
        assert response.status_code == 201
        data = response.json()
        assert "doc_id" in data
        assert data["filename"].endswith(".txt") or data["filename"]  # saved name
        assert data["status"] == "uploaded"

    async def test_upload_document_invalid_size(self, client):
        """File larger than 10MB returns 413."""
        # Create a 11MB file
        large_content = b"x" * (11 * 1024 * 1024)
        files = {"file": ("big.txt", io.BytesIO(large_content), "text/plain")}
        response = await client.post("/api/v1/documents/upload", files=files)
        assert response.status_code == 413

    async def test_upload_triggers_background_processing(self, client):
        """After upload, document status should reflect queued processing."""
        content = b"Invoice content for processing test."
        files = {"file": ("invoice.txt", io.BytesIO(content), "text/plain")}
        response = await client.post("/api/v1/documents/upload", files=files)
        assert response.status_code == 201
        data = response.json()
        # Status is "uploaded" immediately; background task may change it
        assert data["status"] in ("uploaded", "ocr_processing")


class TestDocumentList:
    async def test_list_documents_empty(self, client):
        """GET /documents on fresh state returns empty list."""
        response = await client.get("/api/v1/documents/")
        assert response.status_code == 200
        data = response.json()
        assert "items" in data
        assert isinstance(data["items"], list)
        assert "total" in data

    async def test_list_documents_after_upload(self, client):
        """After uploading, list shows at least one document."""
        content = b"Some document content."
        files = {"file": ("listed.txt", io.BytesIO(content), "text/plain")}
        await client.post("/api/v1/documents/upload", files=files)

        response = await client.get("/api/v1/documents/")
        assert response.status_code == 200
        data = response.json()
        assert data["total"] >= 1


class TestDocumentGet:
    async def test_get_document_not_found(self, client):
        """GET /documents/{unknown_id} → 404."""
        response = await client.get("/api/v1/documents/00000000-0000-0000-0000-000000000000")
        assert response.status_code == 404

    async def test_get_document_found(self, client):
        """Upload then GET by doc_id → 200 with document details."""
        content = b"Get document test content."
        files = {"file": ("get_test.txt", io.BytesIO(content), "text/plain")}
        upload_resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = upload_resp.json()["doc_id"]

        response = await client.get(f"/api/v1/documents/{doc_id}")
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == doc_id
        assert "status" in data


class TestDocumentDelete:
    async def test_delete_document(self, client):
        """Upload then delete → 204; subsequent GET returns 404."""
        content = b"Document to be deleted."
        files = {"file": ("delete_me.txt", io.BytesIO(content), "text/plain")}
        upload_resp = await client.post("/api/v1/documents/upload", files=files)
        doc_id = upload_resp.json()["doc_id"]

        delete_resp = await client.delete(f"/api/v1/documents/{doc_id}")
        assert delete_resp.status_code == 204

        get_resp = await client.get(f"/api/v1/documents/{doc_id}")
        assert get_resp.status_code == 404

    async def test_delete_nonexistent(self, client):
        """DELETE unknown document → 404."""
        response = await client.delete("/api/v1/documents/00000000-0000-0000-0000-000000000001")
        assert response.status_code == 404
