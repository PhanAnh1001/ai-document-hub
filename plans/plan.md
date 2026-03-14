# Project Plan: AI Document Hub
> Last updated: 2026-04-04

## Trạng thái hiện tại
- Phase: 0 — System Design (review)
- Branch: `claude/setup-side-project-6zd9p`

## TODO
- [ ] Review system design files (`system-design/overview.md`, `system-design/system-design.md`)
- [ ] Phase 1: FastAPI skeleton, Docker setup, PostgreSQL + pgvector, PaddleOCR integration
- [ ] Phase 2: LLM extraction module, Groq integration, prompt templates, basic Next.js UI
- [ ] Phase 3: RAG pipeline — BGE-M3 embedding, document indexing, query, chat UI
- [ ] Phase 4: Evaluation — RAGAS pipeline, field-level accuracy, metrics dashboard
- [ ] Phase 5: Fine-tuning — QLoRA notebook trên Colab, dataset prep, before/after comparison
- [ ] Phase 6: Polish — sample data, README, demo prep, CI/CD

## Đã hoàn thành
- [x] Tạo system design: `system-design/overview.md` + `system-design/system-design.md` (10 Mermaid diagrams)

## Notes
- Concept: AI Document Hub — upload document (hóa đơn, hợp đồng, CV, báo cáo) → OCR → extraction → RAG Q&A
- Stack: Python FastAPI + Next.js + PostgreSQL/pgvector + Redis (bỏ Go backend)
- OCR: PaddleOCR local (primary) + Google Vision (fallback, free 1k/month)
- LLM: Groq/Llama 3.3 70B (free 14.4k req/day) → Gemini Flash (backup) → Ollama (optional)
- Embedding: BGE-M3 local (sentence-transformers, 1024 dim)
- Fine-tuning: QLoRA trên Google Colab Free (T4 GPU)
- RAM local: ~4.2GB (không Ollama), máy dev 16GB
- Chi phí: $0/tháng (dev + demo)
- CI/CD secrets cần set: `GROQ_API_KEY`, `GOOGLE_VISION_API_KEY`, `LIGHTSAIL_HOST`, `LIGHTSAIL_SSH_KEY`

## Lịch sử (milestone)
- 2026-04-04: Tạo system design cho AI Document Hub (overview + architecture + 10 Mermaid diagrams)
