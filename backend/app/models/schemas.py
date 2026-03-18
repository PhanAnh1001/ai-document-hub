from datetime import datetime
from typing import Any, Optional
from pydantic import BaseModel, EmailStr


# ── Document schemas ──────────────────────────────────────────────────────────

class DocumentUploadResponse(BaseModel):
    doc_id: str
    filename: str
    status: str
    message: str = "Document uploaded successfully"


class DocumentResponse(BaseModel):
    id: str
    filename: str
    original_filename: str
    file_size: Optional[float] = None
    mime_type: Optional[str] = None
    doc_type: str
    status: str
    ocr_text: Optional[str] = None
    ocr_confidence: Optional[float] = None
    extracted_data: Optional[dict] = None
    error_message: Optional[str] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class DocumentListResponse(BaseModel):
    items: list[DocumentResponse]
    total: int
    page: int
    page_size: int


# ── OCR schemas ───────────────────────────────────────────────────────────────

class OCRResponse(BaseModel):
    doc_id: str
    text: str
    confidence: float
    provider: str  # "paddle" | "google_vision" | "mock"
    status: str


# ── Extraction schemas ────────────────────────────────────────────────────────

class LineItem(BaseModel):
    description: Optional[str] = None
    quantity: Optional[float] = None
    unit_price: Optional[float] = None
    amount: Optional[float] = None


class ExtractionRequest(BaseModel):
    doc_type: Optional[str] = "other"


class ExtractionResponse(BaseModel):
    doc_id: str
    doc_type: str
    status: str
    extracted_data: dict[str, Any]


# ── RAG / Query schemas ───────────────────────────────────────────────────────

class QueryRequest(BaseModel):
    question: str
    doc_ids: Optional[list[str]] = None  # filter to specific docs


class Citation(BaseModel):
    chunk_id: str
    document_id: str
    text_snippet: str
    chunk_index: int


class QueryResponse(BaseModel):
    answer: str
    sources: list[Citation]
    question: str


class ChatHistoryItem(BaseModel):
    id: str
    question: str
    answer: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


# ── Evaluation schemas ────────────────────────────────────────────────────────

class EvalResponse(BaseModel):
    faithfulness: float
    answer_relevancy: float
    context_precision: float
    extraction_accuracy: Optional[float] = None
    doc_id: Optional[str] = None
    evaluated_at: Optional[datetime] = None


class EvalStatsResponse(BaseModel):
    total_evaluations: int
    avg_faithfulness: float
    avg_answer_relevancy: float
    avg_context_precision: float
    avg_extraction_accuracy: Optional[float] = None


class EvalRunRequest(BaseModel):
    doc_id: str
    question: Optional[str] = None
    expected_answer: Optional[str] = None


# ── Auth schemas ──────────────────────────────────────────────────────────────

class UserCreate(BaseModel):
    email: str
    password: str
    full_name: Optional[str] = None


class UserResponse(BaseModel):
    id: str
    email: str
    full_name: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class Token(BaseModel):
    access_token: str
    token_type: str = "bearer"


class LoginRequest(BaseModel):
    email: str
    password: str
