# Project Plan: AI Document Hub
> Last updated: 2026-04-04

## Trạng thái hiện tại
- Phase: 1+2 — Completed (FastAPI backend + Next.js AI Document Hub pages)
- Branch: `claude/complete-implementation-tdd-TLcKo`

## TODO
- [ ] Cập nhật docker-compose.yml: đổi sang FastAPI (port 8000) + pgvector + Redis
- [ ] Phase 3: RAG pipeline — BGE-M3 embedding, document indexing, query pipeline
- [ ] Phase 4: Evaluation — RAGAS pipeline, field-level accuracy, metrics dashboard
- [ ] Phase 5: Fine-tuning — QLoRA notebook trên Colab, dataset prep, before/after comparison
- [ ] Phase 6: Polish — sample data, README, demo prep, CI/CD

## Notes
- Concept: AI Document Hub — upload document (hóa đơn, hợp đồng, CV, báo cáo) → OCR → extraction → RAG Q&A
- Stack: Python FastAPI (port 8000) + Next.js (port 3000) + PostgreSQL/pgvector + Redis
- Backend: `backend/` — Python FastAPI, uv package manager, pytest (28/28 pass)
- Frontend: `app/` — Next.js 15, Vitest (199/199 pass), Playwright E2E
- OCR: PaddleOCR local (mock in tests) + Google Vision fallback
- LLM: Groq/Llama 3.3 70B — mock khi không có API key (`GROQ_API_KEY`)
- Embedding: BGE-M3 local (mock in tests, 1024 dim)
- Test strategy: SQLite (aiosqlite) cho pytest, jsdom cho Vitest, Playwright E2E mocked
- Run backend: `cd backend && uv run uvicorn app.main:app --reload`
- Run tests: backend `uv run pytest`, frontend `cd app && npm test`
- CI/CD secrets cần set: `GROQ_API_KEY`, `GOOGLE_VISION_API_KEY`, `LIGHTSAIL_HOST`, `LIGHTSAIL_SSH_KEY`

## Đã hoàn thành
- [x] Tạo system design: `system-design/overview.md` + `system-design/system-design.md` (10 Mermaid diagrams)
- [x] Phase 1+2: Python FastAPI backend (28 pytest) + Next.js AI Document Hub UI (199 Vitest), TDD
