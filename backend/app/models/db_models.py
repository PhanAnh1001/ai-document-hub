import uuid
from datetime import datetime
from sqlalchemy import Column, String, DateTime, Text, Float, JSON, Enum, ForeignKey
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from sqlalchemy.orm import DeclarativeBase
import enum

# Try to import pgvector; gracefully skip if not installed or DB doesn't support it
try:
    from pgvector.sqlalchemy import Vector
    PGVECTOR_AVAILABLE = True
except ImportError:
    PGVECTOR_AVAILABLE = False
    Vector = None


class DocumentStatus(str, enum.Enum):
    UPLOADED = "uploaded"
    OCR_PROCESSING = "ocr_processing"
    OCR_DONE = "ocr_done"
    EXTRACTING = "extracting"
    EXTRACTED = "extracted"
    INDEXING = "indexing"
    INDEXED = "indexed"
    FAILED = "failed"


class DocumentType(str, enum.Enum):
    INVOICE = "invoice"
    CONTRACT = "contract"
    CV = "cv"
    REPORT = "report"
    OTHER = "other"


class Base(DeclarativeBase):
    pass


def _uuid_default():
    return uuid.uuid4()


# Use String for UUID to support both SQLite (tests) and PostgreSQL (prod)
class User(Base):
    __tablename__ = "users"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    email = Column(String(255), unique=True, nullable=False, index=True)
    hashed_password = Column(String(255), nullable=False)
    full_name = Column(String(255))
    created_at = Column(DateTime, default=datetime.utcnow)


class Document(Base):
    __tablename__ = "documents"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String(36), ForeignKey("users.id"), nullable=False)
    filename = Column(String(500), nullable=False)
    original_filename = Column(String(500), nullable=False)
    file_path = Column(String(1000))
    file_size = Column(Float)
    mime_type = Column(String(100))
    doc_type = Column(String(50), default=DocumentType.OTHER.value)
    status = Column(String(50), default=DocumentStatus.UPLOADED.value)
    ocr_text = Column(Text)
    ocr_confidence = Column(Float)
    extracted_data = Column(JSON)
    error_message = Column(Text)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class DocumentChunk(Base):
    __tablename__ = "document_chunks"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    document_id = Column(String(36), ForeignKey("documents.id"), nullable=False)
    chunk_index = Column(Float, nullable=False)
    chunk_text = Column(Text, nullable=False)
    # Embedding stored as JSON array (fallback when pgvector not available)
    embedding_json = Column(JSON)
    created_at = Column(DateTime, default=datetime.utcnow)


class ChatHistory(Base):
    __tablename__ = "chat_history"

    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String(36), ForeignKey("users.id"), nullable=False)
    question = Column(Text, nullable=False)
    answer = Column(Text)
    context_chunks = Column(JSON)
    created_at = Column(DateTime, default=datetime.utcnow)
