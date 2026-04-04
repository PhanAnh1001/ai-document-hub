"""
FastAPI entry point for AI Document Hub backend.
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.routers import documents, ocr, extract, query
from app.routers import eval as eval_router

app = FastAPI(
    title="AI Document Hub API",
    version="0.1.0",
    description="Intelligent document processing: OCR, extraction, RAG Q&A",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(documents.router, prefix="/api/v1/documents", tags=["documents"])
app.include_router(ocr.router, prefix="/api/v1/ocr", tags=["ocr"])
app.include_router(extract.router, prefix="/api/v1/extract", tags=["extraction"])
app.include_router(query.router, prefix="/api/v1/query", tags=["rag"])
app.include_router(eval_router.router, prefix="/api/v1/eval", tags=["evaluation"])


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "ok", "service": "ai-document-hub"}
