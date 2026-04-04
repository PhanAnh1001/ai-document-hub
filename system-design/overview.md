# AI Document Hub — Project Overview

## Vision

**AI Document Hub** is an intelligent document processing platform for Vietnamese businesses. Upload any document (invoices, contracts, reports, CVs) and the system will:

1. **Extract text** via OCR (PaddleOCR local + Google Vision fallback)
2. **Parse structured data** via LLM (Groq/Llama 3.3 70B)
3. **Index for search** via RAG (BGE-M3 embeddings + pgvector)
4. **Answer questions** in natural language ("Tổng chi phí tháng 3 là bao nhiêu?")

**Target users:** Vietnamese SMEs, accountants, legal teams, HR departments.

---

## Tech Stack

### Architecture: Next.js (Frontend) → FastAPI (Backend + AI) → PostgreSQL/pgvector

| Layer | Technology | Why |
|-------|-----------|-----|
| **Frontend** | Next.js 15 + TypeScript + Tailwind CSS | Modern React, SSR, fast DX |
| **Backend/API** | Python FastAPI | Best ecosystem for AI/ML, async, fast |
| **OCR (primary)** | PaddleOCR (local) | Free, Vietnamese support, runs on CPU |
| **OCR (fallback)** | Google Vision API | High accuracy, free 1k units/month |
| **LLM** | Groq (Llama 3.3 70B) | Free 14.4k req/day, 500+ tokens/s |
| **LLM (backup)** | Gemini Flash | Free 1M tokens/day |
| **Embedding** | BGE-M3 (local, sentence-transformers) | Multilingual, excellent Vietnamese |
| **Vector DB** | PostgreSQL 16 + pgvector | ACID, SQL + vector in one DB |
| **Queue** | Redis | Async document processing |
| **Fine-tuning** | QLoRA (PEFT + HuggingFace) on Google Colab | Free T4 GPU |
| **Evaluation** | RAGAS + custom field-level metrics | Industry standard for RAG |
| **Local LLM (optional)** | Ollama + Qwen2.5-7B Q4 | Offline capability |

### Resource Requirements

| Environment | RAM | GPU | Cost/month |
|-------------|-----|-----|------------|
| Dev (local) | ~4.2GB (no Ollama) / ~10GB (with Ollama) | Not required | **$0** |
| Fine-tuning | Google Colab Free T4 | T4 16GB (free) | **$0** |
| Production | Lightsail 2GB + Groq free | Not required | **$10** |

---

## Cost Analysis — $0 Strategy

| Service | Free Tier | Usage |
|---------|-----------|-------|
| **PaddleOCR** | Open source, local | Primary OCR — unlimited |
| **Google Vision** | 1,000 units/month | Fallback OCR for complex documents |
| **Groq** | 30 RPM, 14,400 req/day | Primary LLM — extraction + RAG generation |
| **Gemini Flash** | 15 RPM, 1M tokens/day | Backup LLM |
| **BGE-M3** | Open source, local | Embedding — unlimited |
| **PostgreSQL + pgvector** | Self-hosted (Docker) | Database + vector store |
| **Google Colab** | T4 GPU, ~12h/day | Fine-tuning QLoRA |

**Total dev + demo cost: $0/month**

---

## Repo Structure

```
ai-classification/
├── README.md                         # Quick start guide
├── CLAUDE.md                         # AI agent configuration
├── docker-compose.yml                # Dev: FastAPI + Next.js + Postgres + Redis
├── docker-compose.prod.yml           # Production deployment
├── .env.example                      # Environment variables template
│
├── app/                              # Next.js Frontend
│   ├── src/
│   │   ├── app/
│   │   │   ├── (auth)/               # Login, register
│   │   │   ├── (dashboard)/
│   │   │   │   ├── documents/        # Upload & document management
│   │   │   │   ├── chat/             # RAG Q&A chat interface
│   │   │   │   ├── evaluation/       # AI metrics dashboard
│   │   │   │   └── dashboard/        # Overview & stats
│   │   │   └── layout.tsx
│   │   ├── components/               # Shared UI components
│   │   └── lib/                      # API client, utilities
│   ├── package.json
│   └── Dockerfile
│
├── backend/                          # Python FastAPI Backend
│   ├── app/
│   │   ├── main.py                   # FastAPI entry point
│   │   ├── config.py                 # Settings (env vars, model paths)
│   │   ├── routers/
│   │   │   ├── documents.py          # CRUD documents
│   │   │   ├── ocr.py                # OCR endpoints
│   │   │   ├── extract.py            # LLM extraction endpoints
│   │   │   ├── query.py              # RAG Q&A endpoints
│   │   │   └── eval.py               # Evaluation metrics endpoints
│   │   ├── services/
│   │   │   ├── ocr_service.py        # PaddleOCR + Google Vision
│   │   │   ├── extraction_service.py # LLM structured extraction
│   │   │   ├── rag_service.py        # Embedding, retrieval, generation
│   │   │   ├── evaluation_service.py # RAGAS + field-level metrics
│   │   │   └── llm_client.py         # Unified LLM interface (Groq/Gemini/Ollama)
│   │   ├── preprocessing/
│   │   │   ├── image_processor.py    # OpenCV: deskew, denoise, contrast
│   │   │   └── text_processor.py     # Vietnamese text normalization
│   │   ├── models/
│   │   │   ├── schemas.py            # Pydantic request/response models
│   │   │   ├── db_models.py          # SQLAlchemy models
│   │   │   └── prompts.py            # All LLM prompt templates
│   │   └── worker.py                 # Redis queue consumer (async processing)
│   ├── migrations/                   # Alembic database migrations
│   ├── tests/                        # Pytest tests
│   ├── pyproject.toml                # Python dependencies
│   └── Dockerfile
│
├── notebooks/                        # Jupyter Notebooks (Google Colab ready)
│   ├── fine_tune_qlora.ipynb         # QLoRA fine-tuning Qwen2.5-7B
│   ├── evaluation_rag.ipynb          # RAGAS evaluation pipeline
│   └── evaluation_extraction.ipynb   # Field-level accuracy measurement
│
├── evaluation/                       # Evaluation artifacts
│   ├── datasets/
│   │   ├── invoice_test_100.json     # 100 labeled Vietnamese invoices
│   │   └── rag_test_200.json         # 200 Q&A pairs for RAG eval
│   └── results/
│       ├── extraction_accuracy.json  # Field-level accuracy results
│       └── ragas_scores.json         # RAGAS metric results
│
├── data/                             # Sample data for demo
│   └── samples/                      # 5-10 sample Vietnamese documents
│
├── system-design/                    # Architecture & design docs
│   ├── overview.md                   # This file
│   └── system-design.md             # Detailed system design + Mermaid diagrams
│
├── plans/
│   └── plan.md                       # Project tracking
│
└── .github/
    └── workflows/                    # CI/CD
```

---

## Key Features & AI Modules

### Module 1: OCR Pipeline
- PaddleOCR with OpenCV preprocessing (deskew, denoise, contrast)
- Google Vision API as high-accuracy fallback
- Confidence-based routing: PaddleOCR → if confidence < 0.8 → Google Vision
- Vietnamese language optimized

### Module 2: LLM Structured Extraction
- Takes raw OCR text → structured JSON (vendor, tax_id, amounts, dates, line items)
- Prompt templates per document type (invoice, contract, CV, report)
- Groq (Llama 3.3 70B) as primary, Gemini Flash as backup
- Few-shot prompting with Vietnamese examples

### Module 3: RAG Document Q&A
- Document ingestion: chunk → BGE-M3 embed → pgvector storage
- Retrieval: cosine similarity + metadata filtering (by document type, date range)
- Re-ranking: LLM-based re-ranker (top-20 → top-5)
- Generation: grounded prompt with citations
- Chat interface with conversation history

### Module 4: Fine-tuning Pipeline (Colab)
- QLoRA (4-bit) fine-tuning of Qwen2.5-7B on Vietnamese invoice data
- 500 labeled samples, 3 epochs, ~2h on T4 GPU
- Before/after accuracy comparison
- Exportable LoRA adapter

### Module 5: Evaluation Pipeline
- **RAG**: RAGAS metrics — faithfulness, answer relevancy, context precision, context recall
- **Extraction**: field-level accuracy per field (vendor, tax_id, total, date, line items)
- Dashboard showing real-time metrics
- Reproducible via Jupyter notebooks

---

## Quick Start

```bash
# Clone
git clone https://github.com/phananh1001/ai-classification.git
cd ai-classification

# Configure
cp .env.example .env
# Set GROQ_API_KEY, GOOGLE_VISION_API_KEY (optional)

# Start all services
docker compose up --build

# Access
# Frontend: http://localhost:3000
# Backend API: http://localhost:8000/docs (FastAPI Swagger)
```

---

## Demo Script (3 minutes)

1. **Upload** a Vietnamese invoice image → watch real-time processing status
2. **View extraction** → structured JSON with vendor, tax_id, amounts, line items
3. **Ask a question** in chat → "Tổng chi phí mua hàng tháng này bao nhiêu?" → RAG-grounded answer with citations
4. **Show metrics** → Evaluation dashboard with RAGAS scores + extraction accuracy

---

## Roadmap

| Phase | Focus | Duration |
|-------|-------|----------|
| **1. Foundation** | FastAPI skeleton, Docker setup, PostgreSQL + pgvector, PaddleOCR integration | Week 1 |
| **2. Extraction** | LLM extraction module, Groq integration, prompt templates, basic Next.js UI | Week 2 |
| **3. RAG** | BGE-M3 embedding, document indexing, query pipeline, chat UI | Week 3 |
| **4. Evaluation** | RAGAS pipeline, field-level accuracy, metrics dashboard | Week 4 |
| **5. Fine-tuning** | QLoRA notebook, dataset preparation, before/after comparison | Week 5 |
| **6. Polish** | Sample data, README, demo prep, CI/CD | Week 6 |
