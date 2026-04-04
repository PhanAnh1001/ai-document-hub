# Project Plan: AI Document Hub
> Last updated: 2026-04-04

## Trạng thái hiện tại
- Phase: 1–2 — In Progress (Python FastAPI backend + Next.js UI đang build)
- Branch: `claude/complete-implementation-tdd-TLcKo`

## TODO
- [ ] Hoàn tất Python FastAPI backend (agents đang chạy): routers, services, TDD pytest
- [ ] Hoàn tất Next.js frontend: `/chat`, `/evaluation`, cập nhật `/documents`, Vitest + Playwright E2E
- [ ] Cập nhật docker-compose.yml: đổi sang FastAPI (port 8000) + pgvector + Redis
- [ ] Phase 3: RAG pipeline — BGE-M3 embedding, document indexing, query, chat UI
- [ ] Phase 4: Evaluation — RAGAS pipeline, field-level accuracy, metrics dashboard
- [ ] Phase 5: Fine-tuning — QLoRA notebook trên Colab, dataset prep, before/after comparison
- [ ] Phase 6: Polish — sample data, README, demo prep, CI/CD

## Notes
- Concept: AI Document Hub — upload document (hóa đơn, hợp đồng, CV, báo cáo) → OCR → extraction → RAG Q&A
- Stack: Python FastAPI (port 8000) + Next.js (port 3000) + PostgreSQL/pgvector + Redis
- Go backend (backend/) đang được thay thế bằng Python FastAPI
- OCR: PaddleOCR local (primary) + Google Vision (fallback, free 1k/month) — mock trong tests
- LLM: Groq/Llama 3.3 70B (free 14.4k req/day) → Gemini Flash (backup) — mock khi không có API key
- Embedding: BGE-M3 local (sentence-transformers, 1024 dim) — mock trong tests
- Fine-tuning: QLoRA trên Google Colab Free (T4 GPU)
- Test strategy: SQLite (aiosqlite) cho pytest, Vitest cho Next.js, Playwright E2E
- CI/CD secrets cần set: `GROQ_API_KEY`, `GOOGLE_VISION_API_KEY`, `LIGHTSAIL_HOST`, `LIGHTSAIL_SSH_KEY`

## Đã hoàn thành
- [x] Tạo system design: `system-design/overview.md` + `system-design/system-design.md` (10 Mermaid diagrams)
- [x] Review system design, xác nhận tech stack và kiến trúc
- [x] Khởi động Phase 1+2: scaffold FastAPI backend + Next.js AI Document Hub pages (TDD)
