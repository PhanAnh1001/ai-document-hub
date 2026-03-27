# AI Document Hub

Intelligent document processing platform for Vietnamese businesses.
Upload invoices, contracts, CVs → OCR → LLM extraction → RAG Q&A.

## Features

- **OCR**: PaddleOCR (local, free) + Google Vision (fallback)
- **LLM Extraction**: Groq/Llama 3.3 70B — structured JSON from documents
- **RAG Q&A**: BGE-M3 embeddings + pgvector — chat with your documents
- **Evaluation**: RAGAS metrics + field-level accuracy dashboard
- **Cost**: $0/month in development

## Quick Start

```bash
# Clone and configure
git clone https://github.com/phananh1001/ai-classification.git
cd ai-classification
cp .env.example .env
# Optional: add GROQ_API_KEY for real LLM (works without it using mock)

# Start all services
docker compose up --build

# Access
# Frontend: http://localhost:3000
# API docs: http://localhost:8000/docs
```

## Development

**Backend (Python FastAPI):**
```bash
cd backend
uv sync
uv run uvicorn app.main:app --reload --port 8000
uv run pytest  # 28 tests
```

**Frontend (Next.js):**
```bash
cd app
npm install
npm run dev  # port 3000
npm test     # 199 tests
```

## Architecture

```
Next.js (3000) → FastAPI (8000) → PostgreSQL/pgvector + Redis
                       ↓
              PaddleOCR / Google Vision
              Groq API / BGE-M3
```

## Demo Flow

1. Upload Vietnamese invoice → auto OCR + extraction
2. Ask "Tổng chi phí mua hàng tháng này bao nhiêu?" → RAG answer with citations
3. View Evaluation dashboard → RAGAS scores

## Project Structure

```
├── app/          # Next.js frontend
├── backend/      # Python FastAPI + AI services
│   ├── app/      # FastAPI app
│   └── tests/    # pytest (28 tests)
├── notebooks/    # Jupyter: fine-tuning + evaluation
├── data/samples/ # Sample Vietnamese documents
└── system-design/# Architecture diagrams
```

## Stack

| Layer | Tech |
|-------|------|
| Frontend | Next.js 15 + TypeScript + Tailwind |
| Backend | Python FastAPI + SQLAlchemy async |
| OCR | PaddleOCR + Google Vision API |
| LLM | Groq (Llama 3.3 70B) + Gemini Flash backup |
| Embedding | BGE-M3 (local, 1024 dim) |
| Vector DB | PostgreSQL 16 + pgvector |
| Fine-tuning | QLoRA (PEFT) on Google Colab |
