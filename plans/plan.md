# Project Plan: AI Document Hub
> Last updated: 2026-04-04

## Trạng thái hiện tại
- Phase: ALL DONE — Full implementation complete
- Branch: `claude/complete-implementation-tdd-TLcKo`

## TODO
- [ ] Merge PR vào master sau khi review
- [ ] Set CI/CD secrets: `GROQ_API_KEY`, `GOOGLE_VISION_API_KEY`, `LIGHTSAIL_HOST`, `LIGHTSAIL_SSH_KEY`
- [ ] Smoke test với `docker compose up --build` trên máy thật
- [ ] Add thêm training data cho fine-tuning (>500 samples cho QLoRA)

## Notes
- Concept: AI Document Hub — upload document (hóa đơn, hợp đồng, CV, báo cáo) → OCR → extraction → RAG Q&A
- Stack: Python FastAPI (port 8000) + Next.js (port 3000) + PostgreSQL/pgvector + Redis
- Backend: `backend/` — Python FastAPI, uv package manager, pytest **45/45 pass**
- Frontend: `app/` — Next.js 15, Vitest **199/199 pass**, Playwright E2E
- OCR: PaddleOCR local (mock in tests) + Google Vision fallback
- LLM: Groq/Llama 3.3 70B — mock khi không có API key (`GROQ_API_KEY`)
- Embedding: BGE-M3 local (mock in tests, 1024 dim)
- pgvector: `PGVECTOR_AVAILABLE` flag — native vector search khi có pgvector, fallback JSON cosine
- docker-compose: `pgvector/pgvector:pg16` + Redis 7 + FastAPI port 8000
- Test strategy: SQLite (aiosqlite) cho pytest, jsdom cho Vitest, Playwright E2E mocked
- Run backend: `cd backend && uv run uvicorn app.main:app --reload`
- Run tests: backend `uv run pytest`, frontend `cd app && npm test`
- Quick start: `cp .env.example .env && docker compose up --build`
- Eval datasets: `backend/evaluation/datasets/` (invoice_test_100.json, rag_test_200.json)
- Notebooks: `notebooks/` (fine_tune_qlora, evaluation_rag, evaluation_extraction)
- Sample docs: `data/samples/` (5 Vietnamese sample files)

## Đã hoàn thành
- [x] Tạo system design: `system-design/overview.md` + `system-design/system-design.md` (10 Mermaid diagrams)
- [x] Phase 1+2: Python FastAPI backend (28→45 pytest) + Next.js AI Document Hub UI (199 Vitest), TDD
- [x] docker-compose: pgvector/pgvector:pg16 + Redis 7 + FastAPI port 8000
- [x] Phase 3+4: pgvector native vector search + evaluation datasets + batch eval + 17 new tests
- [x] Phase 5+6: QLoRA notebooks (Colab) + RAGAS eval notebook + sample data + CI/CD + README
