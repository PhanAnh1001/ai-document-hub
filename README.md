# AI Document Hub

> Intelligent document processing pipeline for Vietnamese businesses — upload invoices, contracts, CVs → OCR → LLM extraction → RAG Q&A.

[![CI](https://github.com/PhanAnh1001/ai-document-hub/actions/workflows/ci.yml/badge.svg)](https://github.com/PhanAnh1001/ai-document-hub/actions)
[![Backend Tests](https://img.shields.io/badge/pytest-45%20passed-brightgreen)](./backend/tests)
[![Frontend Tests](https://img.shields.io/badge/vitest-199%20passed-brightgreen)](./app)
[![Python](https://img.shields.io/badge/python-3.11-blue)](./backend)
[![Next.js](https://img.shields.io/badge/Next.js-15-black)](./app)

[Vietnamese README → README_VI.md](./README_VI.md)

---

## Problem

Vietnamese SMEs process hundreds of paper documents monthly (invoices, contracts, HR files). Manual data entry is slow, error-prone, and expensive. This project automates the entire pipeline: scan → understand → answer questions.

---

## AI Pipeline

```
Upload (PDF / Image / TXT)
        │
        ▼
   ┌─────────────────────────────────────────┐
   │  Step 1: OCR                            │
   │  PaddleOCR (local, free, Vietnamese)    │
   │    └── fallback: Google Vision API      │
   │    └── provider + confidence tracked    │
   └─────────────────┬───────────────────────┘
                     │ raw text
                     ▼
   ┌─────────────────────────────────────────┐
   │  Step 2: LLM Extraction                 │
   │  Groq / Llama 3.3 70B (500 tok/s free) │
   │  Type-specific prompts:                 │
   │    invoice  → vendor, total, tax, date  │
   │    contract → parties, value, terms     │
   │    CV       → candidate, skills, exp.   │
   └─────────────────┬───────────────────────┘
                     │ structured JSON
                     ▼
   ┌─────────────────────────────────────────┐
   │  Step 3: RAG Indexing                   │
   │  BGE-M3 embeddings (1024-dim, local)    │
   │  Chunk: 500 chars / 50 overlap          │
   │  pgvector IVFFLAT → cosine fallback     │
   └─────────────────┬───────────────────────┘
                     │ vector chunks in PostgreSQL
                     ▼
            Chat / Q&A with citations
```

All three steps run as an **async background task** after upload. The API returns immediately with a `doc_id`; the frontend polls the status endpoint.

---

## Architecture

```
Browser
  │  Next.js 15 (port 3000)
  │   ├── Upload UI
  │   ├── Extraction viewer (structured JSON)
  │   ├── RAG chat interface + citations
  │   └── Evaluation dashboard (RAGAS scores)
  │
  ▼ REST API
FastAPI (port 8000)
  │   ├── POST /api/v1/documents/upload
  │   ├── POST /api/v1/ocr/
  │   ├── POST /api/v1/extract/
  │   ├── POST /api/v1/query/          ← RAG Q&A with citations
  │   └── GET  /api/v1/eval/stats
  │
  ├── PostgreSQL 16 + pgvector   ← documents + vector chunks
  └── Redis 7                    ← task queue
```

---

## Technical Highlights

### 1. Dual OCR with Graceful Fallback
- **PaddleOCR** runs locally — no API cost, works offline, strong Vietnamese support
- Falls back to **Google Vision** when PaddleOCR is disabled or confidence is low
- Response always includes `provider` (`paddle` | `google_vision` | `text` | `mock`) and `confidence` for traceability

### 2. Type-Aware LLM Extraction
- Document type (invoice/contract/CV) selects a dedicated prompt template in `models/prompts.py`
- **Groq + Llama 3.3 70B** — 14,400 free requests/day at ~500 tokens/second
- Robust JSON parsing: strips markdown code fences, handles partial LLM output gracefully

### 3. RAG with pgvector
- **BGE-M3** (BAAI): 1024-dim multilingual embeddings, runs on CPU via sentence-transformers
- Native pgvector extension with IVFFLAT index for ANN search; falls back to pure-SQL cosine similarity over JSON-stored vectors when pgvector is unavailable
- Answers include **citations** (chunk_id, document_id, text snippet, chunk_index)

### 4. RAGAS-style Evaluation
- **Faithfulness** — fraction of answer words present in retrieved context
- **Answer Relevancy** — keyword overlap between question and answer
- **Context Precision** — signal-to-noise ratio of retrieved chunks
- **Field-level Extraction Accuracy** — compared against ground truth per field
- Eval datasets: 100 invoice samples + 200 RAG Q&A pairs with Vietnamese ground truth

### 5. QLoRA Fine-tuning
- `notebooks/fine_tune_qlora.ipynb` — domain-adapted extraction model on Google Colab T4 (free)
- Uses PEFT + HuggingFace Transformers; targets invoice/contract extraction tasks
- Reduces hallucination on domain-specific Vietnamese financial vocabulary

### 6. $0 Development Cost
| Service | Free Tier |
|---------|-----------|
| PaddleOCR | Open source, local |
| Google Vision | 1,000 units/month |
| Groq (Llama 3.3 70B) | 14,400 req/day |
| BGE-M3 | Local via sentence-transformers |
| pgvector | Self-hosted PostgreSQL |
| Google Colab T4 | Fine-tuning GPU |

---

## Quick Start

```bash
git clone https://github.com/PhanAnh1001/ai-document-hub.git
cd ai-document-hub
cp .env.example .env
# Optional: set GROQ_API_KEY for real LLM (system works without it using mock)

docker compose up --build
# Frontend: http://localhost:3000
# API docs: http://localhost:8000/docs
```

### Demo Flow
1. Upload a Vietnamese invoice PDF → OCR + extraction run automatically in the background
2. View extracted JSON: `vendor`, `total`, `tax_amount`, `date`, `invoice_number`
3. Ask **"Tổng chi phí mua hàng tháng 1 bao nhiêu?"** → RAG answer with source citations
4. Open the Evaluation dashboard → RAGAS scores for all processed documents

---

## Development

**Backend (Python / FastAPI):**
```bash
cd backend
uv sync
uv run uvicorn app.main:app --reload --port 8000
uv run pytest                        # 45 tests, all passing
uv run pytest tests/ -v --tb=short   # verbose output
```

**Frontend (Next.js):**
```bash
cd app
npm install
npm run dev          # port 3000
npm test             # 199 tests (Vitest)
npm run test:e2e     # Playwright E2E
```

---

## Project Structure

```
├── app/                      # Next.js 15 frontend
│   └── src/
│       ├── app/              # App Router pages
│       ├── components/       # UI components
│       ├── hooks/            # Custom React hooks
│       └── lib/              # API client, utils
│
├── backend/                  # Python FastAPI + AI services
│   ├── app/
│   │   ├── routers/          # documents, ocr, extract, query, eval
│   │   ├── services/
│   │   │   ├── ocr_service.py         # PaddleOCR + Google Vision
│   │   │   ├── extraction_service.py  # LLM + type-specific prompts
│   │   │   ├── rag_service.py         # BGE-M3 + pgvector
│   │   │   ├── evaluation_service.py  # RAGAS-style metrics
│   │   │   └── llm_client.py          # Groq API client + mock
│   │   ├── models/
│   │   │   ├── db_models.py           # SQLAlchemy ORM
│   │   │   ├── schemas.py             # Pydantic request/response
│   │   │   └── prompts.py             # LLM prompt templates
│   │   └── worker.py                  # Async background pipeline
│   └── tests/                # pytest — 45 tests, TDD
│
├── notebooks/                # Jupyter
│   ├── fine_tune_qlora.ipynb          # QLoRA fine-tuning (Colab T4)
│   ├── evaluation_rag.ipynb           # RAGAS evaluation
│   └── evaluation_extraction.ipynb    # Field-level accuracy
│
├── evaluation/datasets/      # Ground truth
│   ├── invoice_test_100.json          # 100 invoices with expected fields
│   └── rag_test_200.json              # 200 RAG Q&A pairs
│
├── data/samples/             # Sample Vietnamese documents
├── system-design/            # Architecture diagrams (10 Mermaid charts)
└── docker-compose.yml        # pgvector:pg16 + Redis 7 + FastAPI + Next.js
```

---

## Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15 + TypeScript + Tailwind CSS |
| Backend | Python FastAPI + SQLAlchemy (async) |
| OCR | PaddleOCR (local) + Google Vision API |
| LLM | Groq / Llama 3.3 70B + mock fallback |
| Embedding | BGE-M3 (sentence-transformers, 1024-dim) |
| Vector DB | PostgreSQL 16 + pgvector (IVFFLAT) |
| Task Queue | Redis 7 + FastAPI BackgroundTasks |
| Fine-tuning | QLoRA (PEFT + HuggingFace) on Colab T4 |
| Evaluation | RAGAS-style heuristics + field accuracy |
| Testing | pytest 45 / Vitest 199 / Playwright E2E |
| CI/CD | GitHub Actions + Docker + Amazon Lightsail |

---

## Known Limitations & Roadmap

| Area | Current | Next Step |
|------|---------|-----------|
| Auth | No auth (MVP) | JWT + multi-tenant |
| Doc type | Set by client | Auto-classify with LLM/classifier |
| LLM responses | Batch | SSE streaming for chat UX |
| Eval metrics | Heuristic (no LLM judge) | Real RAGAS with LLM-as-judge |
| Fine-tuning | Notebook only | MLflow experiment tracking |

---

## System Design

Detailed architecture diagrams (10 Mermaid charts): [`system-design/`](./system-design/)
