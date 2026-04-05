# AI Document Hub — Tài liệu tiếng Việt

> Nền tảng xử lý tài liệu thông minh cho doanh nghiệp Việt Nam — upload hóa đơn, hợp đồng, CV → OCR → trích xuất dữ liệu → hỏi đáp RAG.

[English README → README.md](./README.md)

---

## Vấn đề giải quyết

Doanh nghiệp vừa và nhỏ tại Việt Nam xử lý hàng trăm tài liệu giấy mỗi tháng: hóa đơn VAT, hợp đồng, hồ sơ nhân sự. Nhập liệu thủ công chậm, dễ sai, và tốn người. Project này tự động hóa toàn bộ pipeline:

**Scan / Upload → OCR → Trích xuất dữ liệu có cấu trúc → Hỏi đáp bằng tiếng Việt**

---

## Pipeline AI

### Bước 1: OCR
- **PaddleOCR** chạy local — miễn phí, offline, hỗ trợ tiếng Việt tốt
- Fallback sang **Google Vision API** khi PaddleOCR không khả dụng
- Kết quả trả về kèm `provider` và `confidence` để traceable

### Bước 2: Trích xuất LLM
- **Groq / Llama 3.3 70B** — 14,400 request/ngày miễn phí, ~500 token/giây
- Prompt riêng theo loại tài liệu:
  - **Hóa đơn**: nhà cung cấp, tổng tiền, thuế VAT, ngày, số hóa đơn
  - **Hợp đồng**: các bên, giá trị, điều khoản, thời hạn
  - **CV**: ứng viên, kỹ năng, kinh nghiệm, học vấn
- Parse JSON robust: xử lý markdown code fence, lỗi định dạng từ LLM

### Bước 3: RAG Indexing & Q&A
- **BGE-M3** (BAAI): embedding 1024 chiều, đa ngôn ngữ, chạy CPU
- Lưu vào **PostgreSQL + pgvector** với IVFFLAT index
- Fallback cosine similarity thuần SQL khi pgvector chưa cài
- Câu trả lời kèm **citation** (đoạn văn gốc, document ID, vị trí chunk)

---

## Cách chạy nhanh

```bash
git clone https://github.com/PhanAnh1001/ai-document-hub.git
cd ai-document-hub
cp .env.example .env
# Tùy chọn: thêm GROQ_API_KEY để dùng LLM thật (không có thì dùng mock)

docker compose up --build
# Frontend: http://localhost:3000
# API docs: http://localhost:8000/docs
```

### Demo thực tế
1. Upload hóa đơn VAT tiếng Việt (PDF/ảnh) → OCR + trích xuất chạy tự động
2. Xem kết quả JSON: `vendor`, `total`, `tax_amount`, `date`, `invoice_number`
3. Hỏi: **"Tổng chi phí mua hàng tháng 1 bao nhiêu?"** → RAG trả lời kèm nguồn trích dẫn
4. Mở dashboard Evaluation → điểm RAGAS của các tài liệu đã xử lý

---

## Môi trường development

**Backend (Python / FastAPI):**
```bash
cd backend
uv sync
uv run uvicorn app.main:app --reload --port 8000
uv run pytest          # 45 tests
```

**Frontend (Next.js):**
```bash
cd app
npm install
npm run dev            # port 3000
npm test               # 199 tests (Vitest)
```

---

## Cấu trúc project

```
├── app/                      # Next.js 15 frontend
│   └── src/
│       ├── app/              # App Router pages (upload, chat, eval)
│       ├── components/       # UI components
│       └── lib/              # API client
│
├── backend/                  # Python FastAPI + dịch vụ AI
│   ├── app/
│   │   ├── routers/          # documents, ocr, extract, query, eval
│   │   ├── services/
│   │   │   ├── ocr_service.py         # PaddleOCR + Google Vision
│   │   │   ├── extraction_service.py  # LLM + prompt templates
│   │   │   ├── rag_service.py         # BGE-M3 + pgvector
│   │   │   └── evaluation_service.py  # RAGAS-style metrics
│   │   ├── models/
│   │   │   └── prompts.py             # Prompt templates theo loại tài liệu
│   │   └── worker.py                  # Async background pipeline
│   └── tests/                # pytest 45 tests (TDD)
│
├── notebooks/
│   ├── fine_tune_qlora.ipynb          # Fine-tuning QLoRA trên Colab T4
│   ├── evaluation_rag.ipynb           # Đánh giá RAG với RAGAS
│   └── evaluation_extraction.ipynb    # Độ chính xác trích xuất từng field
│
├── evaluation/datasets/
│   ├── invoice_test_100.json          # 100 hóa đơn có ground truth
│   └── rag_test_200.json              # 200 cặp hỏi-đáp tiếng Việt
│
├── data/samples/             # 5 tài liệu mẫu tiếng Việt
└── system-design/            # 10 sơ đồ kiến trúc (Mermaid)
```

---

## Stack

| Tầng | Công nghệ |
|------|-----------|
| Frontend | Next.js 15 + TypeScript + Tailwind CSS |
| Backend | Python FastAPI + SQLAlchemy async |
| OCR | PaddleOCR (local) + Google Vision API |
| LLM | Groq / Llama 3.3 70B + mock fallback |
| Embedding | BGE-M3 (sentence-transformers, 1024-dim) |
| Vector DB | PostgreSQL 16 + pgvector |
| Task Queue | Redis 7 + FastAPI BackgroundTasks |
| Fine-tuning | QLoRA (PEFT + HuggingFace) trên Colab T4 |
| Evaluation | RAGAS-style + field-level accuracy |
| Testing | pytest 45 / Vitest 199 / Playwright E2E |
| CI/CD | GitHub Actions + Docker + Amazon Lightsail |

---

## Chi phí: $0 khi development

| Dịch vụ | Free tier |
|---------|-----------|
| PaddleOCR | Open source, chạy local |
| Google Vision | 1,000 unit/tháng |
| Groq (Llama 3.3 70B) | 14,400 request/ngày |
| BGE-M3 | Local qua sentence-transformers |
| pgvector | Self-hosted PostgreSQL |
| Google Colab T4 | GPU miễn phí cho fine-tuning |
| Production (Lightsail 2GB) | ~$10/tháng |

---

## Evaluation & Metrics

### Các chỉ số RAG (RAGAS-style)
- **Faithfulness**: câu trả lời có dựa trên context không? (word overlap)
- **Answer Relevancy**: câu trả lời có liên quan đến câu hỏi không?
- **Context Precision**: context retrieved có signal-to-noise tốt không?

### Chỉ số trích xuất
- **Field-level accuracy**: so sánh từng trường (vendor, total, date...) với ground truth
- Dataset: 100 hóa đơn VAT + 200 cặp hỏi đáp tiếng Việt

---

## Fine-tuning QLoRA

`notebooks/fine_tune_qlora.ipynb` — train model domain-specific trên Google Colab T4 (miễn phí):
- PEFT + HuggingFace Transformers
- Mục tiêu: giảm hallucination cho thuật ngữ tài chính/pháp lý tiếng Việt
- Cần >500 samples để fine-tune hiệu quả (hiện có ~100 samples, đang thu thập thêm)

---

## Hạn chế hiện tại

| Vấn đề | Trạng thái | Hướng cải thiện |
|--------|-----------|----------------|
| Auth | Chưa có (MVP) | JWT + multi-tenant |
| Auto-classify | Client tự chọn loại tài liệu | LLM/classifier tự phát hiện |
| LLM streaming | Batch response | SSE streaming |
| Eval metrics | Heuristic | RAGAS với LLM-as-judge |
| Fine-tuning data | ~100 samples | Cần >500 cho QLoRA hiệu quả |

---

## Kiến trúc chi tiết

10 sơ đồ Mermaid: [`system-design/`](./system-design/)
