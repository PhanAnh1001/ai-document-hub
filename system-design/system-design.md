# AI Document Hub — System Design

> Production-ready architecture for intelligent document processing with OCR, LLM extraction, RAG Q&A, fine-tuning, and evaluation.

---

## 1. Architecture Overview

```mermaid
graph TB
    subgraph Client
        UI[Next.js Frontend<br/>Port 3000]
    end

    subgraph Backend["Python FastAPI — Port 8000"]
        API[FastAPI Router]
        Worker[Async Worker]
        OCR[OCR Service]
        Extract[Extraction Service]
        RAG[RAG Service]
        Eval[Evaluation Service]
        LLM[LLM Client<br/>Groq / Gemini / Ollama]
    end

    subgraph Data
        PG[(PostgreSQL 16<br/>+ pgvector)]
        Redis[(Redis<br/>Job Queue)]
    end

    subgraph External["External APIs (Free Tier)"]
        Groq[Groq API<br/>Llama 3.3 70B<br/>Free 14.4k req/day]
        GVision[Google Vision OCR<br/>Free 1k/month]
        Gemini[Gemini Flash<br/>Free 1M tokens/day]
    end

    subgraph Local["Local AI Models"]
        Paddle[PaddleOCR<br/>Vietnamese]
        BGE[BGE-M3<br/>Embedding]
        Ollama[Ollama + Qwen2.5-7B<br/>Optional]
    end

    UI -->|HTTP/REST| API
    API -->|Enqueue jobs| Redis
    Redis -->|Consume jobs| Worker
    Worker --> OCR
    Worker --> Extract
    API --> RAG
    API --> Eval

    OCR --> Paddle
    OCR -->|Fallback| GVision
    Extract --> LLM
    RAG --> BGE
    RAG --> LLM
    LLM --> Groq
    LLM -->|Backup| Gemini
    LLM -->|Optional| Ollama

    OCR --> PG
    Extract --> PG
    RAG --> PG
    Eval --> PG
```

---

## 2. Document Processing Flow (End-to-End)

```mermaid
sequenceDiagram
    actor User
    participant UI as Next.js
    participant API as FastAPI
    participant Redis
    participant Worker
    participant OCR as OCR Service
    participant LLM as LLM Client
    participant DB as PostgreSQL
    participant VectorDB as pgvector

    User->>UI: Upload document (image/PDF)
    UI->>API: POST /api/v1/documents/upload
    API->>DB: Save document record (status: uploaded)
    API->>Redis: Enqueue processing job
    API-->>UI: 202 Accepted (job_id)

    Redis->>Worker: Pick up job
    Worker->>OCR: Extract text
    Note over OCR: PaddleOCR (primary)<br/>Google Vision (fallback)
    OCR-->>Worker: Raw OCR text + confidence
    Worker->>DB: Update (status: ocr_done, raw_text)

    Worker->>LLM: Extract structured data
    Note over LLM: Prompt template + few-shot<br/>→ Groq/Llama 3.3 70B
    LLM-->>Worker: Structured JSON
    Worker->>DB: Save extraction (status: extracted)

    Worker->>VectorDB: Index for RAG
    Note over VectorDB: Chunk text → BGE-M3 embed<br/>→ pgvector INSERT
    Worker->>DB: Update (status: indexed)

    UI->>API: GET /api/v1/documents/{id}
    API->>DB: Fetch document + extraction
    API-->>UI: Document details + structured data

    User->>UI: Ask question in chat
    UI->>API: POST /api/v1/query
    API->>VectorDB: Similarity search (top-k)
    API->>LLM: Generate answer with context
    LLM-->>API: Grounded answer + citations
    API-->>UI: Answer with source references
```

---

## 3. OCR Pipeline

```mermaid
flowchart LR
    subgraph Input
        IMG[Image/PDF Upload]
    end

    subgraph Preprocessing["Preprocessing (OpenCV)"]
        GRAY[Grayscale]
        DESKEW[Deskew Rotation]
        DENOISE[Denoise]
        CONTRAST[Adaptive Contrast]
        CROP[ROI Crop]
    end

    subgraph OCR_Primary["Primary: PaddleOCR (Local)"]
        DETECT[Text Detection]
        RECOG[Text Recognition]
        CONF[Confidence Score]
    end

    subgraph OCR_Fallback["Fallback: Google Vision API"]
        GOCR[Google Vision OCR]
    end

    subgraph Output
        TEXT[Raw OCR Text<br/>+ Bounding Boxes<br/>+ Confidence Scores]
    end

    IMG --> GRAY --> DESKEW --> DENOISE --> CONTRAST --> CROP
    CROP --> DETECT --> RECOG --> CONF

    CONF -->|confidence >= 0.8| TEXT
    CONF -->|confidence < 0.8| GOCR
    GOCR --> TEXT
```

### OCR Provider Comparison

| Feature | PaddleOCR (Local) | Google Vision (API) |
|---------|-------------------|---------------------|
| Cost | Free | Free 1k/month |
| Vietnamese | Good | Excellent |
| Speed | ~1-3s/page | ~1-2s/page |
| Handwriting | Limited | Good |
| Offline | Yes | No |
| Strategy | **Primary** — always try first | **Fallback** — complex/low-confidence docs |

---

## 4. LLM Extraction Pipeline

```mermaid
flowchart TB
    subgraph Input
        RAW[Raw OCR Text]
        DOCTYPE[Document Type<br/>invoice / contract / cv / report]
    end

    subgraph PromptEngine["Prompt Engineering"]
        SELECT[Select Prompt Template]
        FEWSHOT[Inject Few-shot Examples<br/>Vietnamese samples]
        SCHEMA[Define Output JSON Schema]
        BUILD[Build Final Prompt]
    end

    subgraph LLM_Inference["LLM Inference"]
        GROQ[Groq API<br/>Llama 3.3 70B<br/>Primary — Free]
        GEMINI[Gemini Flash<br/>Backup — Free]
        OLLAMA[Ollama / Qwen2.5-7B<br/>Optional — Local]
    end

    subgraph Validation
        PARSE[Parse JSON Response]
        VALIDATE[Validate against Schema<br/>Pydantic]
        RETRY[Retry with Error Feedback<br/>Max 2 retries]
    end

    subgraph Output
        JSON[Structured JSON<br/>vendor, tax_id, amounts,<br/>dates, line_items]
    end

    RAW --> SELECT
    DOCTYPE --> SELECT
    SELECT --> FEWSHOT --> SCHEMA --> BUILD

    BUILD --> GROQ
    GROQ -->|Timeout/Error| GEMINI
    GEMINI -->|Timeout/Error| OLLAMA

    GROQ --> PARSE
    GEMINI --> PARSE
    OLLAMA --> PARSE

    PARSE --> VALIDATE
    VALIDATE -->|Valid| JSON
    VALIDATE -->|Invalid| RETRY
    RETRY --> BUILD
```

### Prompt Template Strategy

```
System: You are a Vietnamese document extraction expert.
Extract ONLY information present in the text.
If a field is not found, return null.
Output format: JSON matching the schema below.

Schema: {json_schema per document type}

Few-shot examples:
- Example 1: [Vietnamese invoice OCR text] → [Expected JSON]
- Example 2: [Vietnamese invoice OCR text] → [Expected JSON]
- Example 3: [Vietnamese invoice OCR text] → [Expected JSON]

User: Extract structured data from this document:
{raw_ocr_text}
```

### Extraction Schema per Document Type

| Document Type | Key Fields |
|---------------|-----------|
| **Invoice** (hóa đơn) | vendor_name, tax_id, invoice_number, date, total_amount, vat_amount, line_items[] |
| **Contract** (hợp đồng) | parties[], effective_date, expiry_date, contract_value, key_terms[] |
| **CV/Resume** | full_name, email, phone, education[], experience[], skills[] |
| **Report** (báo cáo) | title, date, author, summary, key_metrics[] |

---

## 5. RAG Pipeline

```mermaid
flowchart TB
    subgraph Ingestion["Document Ingestion (Async)"]
        DOC[Extracted Document Text]
        CHUNK[Recursive Text Splitter<br/>512 tokens, overlap 50]
        EMBED_I[BGE-M3 Embedding<br/>Local, 1024 dimensions]
        STORE[pgvector INSERT<br/>+ metadata: doc_id, doc_type, date]
    end

    subgraph Query["Query Pipeline (Sync)"]
        Q[User Question]
        EMBED_Q[BGE-M3 Embedding<br/>Query vector]
        SEARCH[pgvector Cosine Similarity<br/>top-20 candidates]
        FILTER[Metadata Filter<br/>doc_type, date_range, user_org]
        RERANK[LLM Re-ranker<br/>top-20 → top-5]
        CONTEXT[Build Context<br/>top-5 chunks + metadata]
    end

    subgraph Generation
        PROMPT[Grounded Prompt<br/>Context + Question + Instructions]
        LLM_GEN[Groq / Llama 3.3 70B]
        CITE[Extract Citations<br/>Source doc + page]
        ANSWER[Answer + Citations]
    end

    DOC --> CHUNK --> EMBED_I --> STORE
    Q --> EMBED_Q --> SEARCH --> FILTER --> RERANK --> CONTEXT
    CONTEXT --> PROMPT --> LLM_GEN --> CITE --> ANSWER
```

### RAG Configuration

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Chunk size | 512 tokens | Balance between context and retrieval precision |
| Chunk overlap | 50 tokens | Prevent information loss at boundaries |
| Embedding model | BGE-M3 (1024 dim) | Best multilingual model for Vietnamese |
| Top-k retrieval | 20 | Cast wide net for re-ranking |
| Re-rank to | 5 | Focus on most relevant chunks |
| Distance metric | Cosine similarity | Standard for normalized embeddings |
| Temperature | 0.1 | Low creativity, high accuracy for factual Q&A |

### RAG Generation Prompt

```
System: You are a helpful assistant that answers questions based ONLY on the
provided context. If the information is not in the context, say "Tôi không
tìm thấy thông tin này trong tài liệu."

Always cite your sources using [Source: document_name, page X] format.

Context:
{retrieved_chunks with metadata}

User: {user_question}
```

---

## 6. Fine-tuning Pipeline

```mermaid
flowchart LR
    subgraph DataPrep["Data Preparation"]
        RAW_DATA[500 Vietnamese Invoice Samples<br/>OCR text + ground truth JSON]
        SPLIT[Train/Val/Test Split<br/>400 / 50 / 50]
        FORMAT[Format for Training<br/>instruction + input + output]
    end

    subgraph Training["Training (Google Colab Free — T4 GPU)"]
        BASE[Base Model<br/>Qwen2.5-7B]
        QUANT[Quantize 4-bit<br/>BitsAndBytes NF4]
        LORA[LoRA Config<br/>rank=16, alpha=32<br/>target: q_proj, v_proj]
        TRAIN[Train 3 Epochs<br/>lr=2e-4, batch=4<br/>~2 hours on T4]
    end

    subgraph Evaluation
        TEST[Test Set<br/>50 samples]
        FIELD[Field-level Accuracy<br/>per field comparison]
        COMPARE[Before vs After<br/>Base model vs Fine-tuned]
    end

    subgraph Deploy
        ADAPTER[Export LoRA Adapter<br/>~50MB]
        MERGE[Optional: Merge with Base]
        SERVE[Load in Production<br/>via Ollama or vLLM]
    end

    RAW_DATA --> SPLIT --> FORMAT
    FORMAT --> BASE --> QUANT --> LORA --> TRAIN
    TRAIN --> TEST --> FIELD --> COMPARE
    TRAIN --> ADAPTER --> MERGE --> SERVE
```

### Fine-tuning Configuration

| Parameter | Value |
|-----------|-------|
| Base model | Qwen2.5-7B |
| Method | QLoRA (4-bit NF4) |
| LoRA rank | 16 |
| LoRA alpha | 32 |
| Target modules | q_proj, v_proj |
| Learning rate | 2e-4 |
| Batch size | 4 |
| Epochs | 3 |
| Training time | ~2h on T4 GPU |
| Dataset | 500 labeled Vietnamese invoices |
| Adapter size | ~50MB |

### Expected Results

| Metric | Before (Base) | After (Fine-tuned) |
|--------|--------------|-------------------|
| Field accuracy (vendor_name) | ~85% | ~95% |
| Field accuracy (tax_id) | ~80% | ~96% |
| Field accuracy (total_amount) | ~88% | ~94% |
| Field accuracy (line_items) | ~70% | ~88% |
| **Average** | **~81%** | **~93%** |

---

## 7. Evaluation Pipeline

```mermaid
flowchart TB
    subgraph RAG_Eval["RAG Evaluation (RAGAS)"]
        QA_SET[200 Q&A Test Pairs<br/>question + ground_truth + context]
        RAGAS[RAGAS Framework]
        F[Faithfulness<br/>Answer grounded in context?]
        AR[Answer Relevancy<br/>Answer relevant to question?]
        CP[Context Precision<br/>Retrieved context relevant?]
        CR[Context Recall<br/>Missing important context?]
        RAGAS_SCORE[RAGAS Scores<br/>Target: avg >= 0.80]
    end

    subgraph Extract_Eval["Extraction Evaluation"]
        INV_SET[100 Labeled Invoices<br/>OCR text + ground truth JSON]
        EXTRACT[Run Extraction Pipeline]
        FIELD_CMP[Field-by-field Comparison]
        EXACT[Exact Match Rate<br/>All fields correct?]
        FIELD_ACC[Per-field Accuracy<br/>vendor, tax_id, total, date, items]
        EXTRACT_SCORE[Extraction Scores<br/>Target: avg >= 90%]
    end

    subgraph Dashboard["Metrics Dashboard"]
        API_METRICS[/api/v1/eval/summary]
        CHART_RAG[RAGAS Radar Chart]
        CHART_EXT[Extraction Accuracy Bar Chart]
        TREND[Accuracy Trend Over Time]
    end

    QA_SET --> RAGAS
    RAGAS --> F & AR & CP & CR
    F & AR & CP & CR --> RAGAS_SCORE

    INV_SET --> EXTRACT --> FIELD_CMP
    FIELD_CMP --> EXACT & FIELD_ACC
    EXACT & FIELD_ACC --> EXTRACT_SCORE

    RAGAS_SCORE --> API_METRICS
    EXTRACT_SCORE --> API_METRICS
    API_METRICS --> CHART_RAG & CHART_EXT & TREND
```

### Target Metrics

| Category | Metric | Target |
|----------|--------|--------|
| RAG | Faithfulness | >= 0.85 |
| RAG | Answer Relevancy | >= 0.80 |
| RAG | Context Precision | >= 0.75 |
| RAG | Context Recall | >= 0.80 |
| RAG | **Average RAGAS** | **>= 0.80** |
| Extraction | Vendor name accuracy | >= 95% |
| Extraction | Tax ID accuracy | >= 96% |
| Extraction | Total amount accuracy | >= 94% |
| Extraction | Line items accuracy | >= 88% |
| Extraction | **Average field accuracy** | **>= 93%** |

---

## 8. Database Schema

```mermaid
erDiagram
    users {
        uuid id PK
        string email UK
        string password_hash
        string name
        timestamp created_at
    }

    documents {
        uuid id PK
        uuid user_id FK
        string filename
        string file_path
        string file_type "image/pdf"
        string doc_type "invoice/contract/cv/report"
        string status "uploaded/ocr_done/extracted/indexed/error"
        text raw_ocr_text
        float ocr_confidence
        string ocr_provider "paddleocr/google_vision"
        timestamp created_at
        timestamp updated_at
    }

    extractions {
        uuid id PK
        uuid document_id FK
        jsonb structured_data "Extracted JSON"
        string llm_provider "groq/gemini/ollama"
        string llm_model "llama-3.3-70b"
        int prompt_tokens
        int completion_tokens
        float confidence
        timestamp created_at
    }

    chunks {
        uuid id PK
        uuid document_id FK
        int chunk_index
        text content
        vector embedding "1024 dimensions (BGE-M3)"
        jsonb metadata "page, position, doc_type"
    }

    conversations {
        uuid id PK
        uuid user_id FK
        string title
        timestamp created_at
    }

    messages {
        uuid id PK
        uuid conversation_id FK
        string role "user/assistant"
        text content
        jsonb citations "Source documents + pages"
        int prompt_tokens
        int completion_tokens
        timestamp created_at
    }

    eval_results {
        uuid id PK
        string eval_type "rag/extraction"
        jsonb metrics "RAGAS scores or field accuracy"
        int sample_count
        string model_version
        timestamp created_at
    }

    users ||--o{ documents : "uploads"
    users ||--o{ conversations : "has"
    documents ||--o| extractions : "has"
    documents ||--o{ chunks : "split into"
    conversations ||--o{ messages : "contains"
```

---

## 9. Deployment Architecture

```mermaid
graph TB
    subgraph Docker["Docker Compose (Development)"]
        NEXT[next-app<br/>Node.js 20<br/>Port 3000]
        FAST[fastapi-app<br/>Python 3.12<br/>Port 8000]
        PG[postgres<br/>PostgreSQL 16 + pgvector<br/>Port 5432]
        RD[redis<br/>Redis 7<br/>Port 6379]
        WORKER_C[worker<br/>Python 3.12<br/>Background processor]
    end

    subgraph Volumes
        PG_DATA[pg_data]
        REDIS_DATA[redis_data]
        UPLOADS[uploads/]
    end

    NEXT -->|HTTP| FAST
    FAST --> PG
    FAST --> RD
    WORKER_C --> PG
    WORKER_C --> RD

    PG --> PG_DATA
    RD --> REDIS_DATA
    FAST --> UPLOADS

    subgraph Production["Production (Lightsail $10/month)"]
        NGINX[Nginx Reverse Proxy<br/>SSL Termination]
        NEXT_P[Next.js<br/>Build + Serve]
        FAST_P[FastAPI<br/>Gunicorn + Uvicorn]
        PG_P[PostgreSQL<br/>+ pgvector]
        RD_P[Redis]
    end

    NGINX --> NEXT_P
    NGINX --> FAST_P
    FAST_P --> PG_P
    FAST_P --> RD_P
```

### Docker Compose Services

| Service | Image | RAM | Port |
|---------|-------|-----|------|
| next-app | node:20-alpine | ~300MB | 3000 |
| fastapi-app | python:3.12-slim | ~150MB | 8000 |
| worker | python:3.12-slim | ~1.5GB (PaddleOCR + BGE-M3) | — |
| postgres | pgvector/pgvector:pg16 | ~200MB | 5432 |
| redis | redis:7-alpine | ~50MB | 6379 |
| **Total** | | **~2.2GB** | |

> Worker loads PaddleOCR + BGE-M3 models. On 16GB RAM machine, still leaves ~13GB headroom.

---

## 10. API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/documents/upload` | Upload document, start async processing |
| GET | `/api/v1/documents` | List user's documents |
| GET | `/api/v1/documents/{id}` | Get document details + extraction |
| POST | `/api/v1/ocr/extract` | Sync OCR extraction (for testing) |
| POST | `/api/v1/extract` | Sync LLM extraction (for testing) |
| POST | `/api/v1/query` | RAG Q&A query |
| GET | `/api/v1/conversations` | List conversations |
| GET | `/api/v1/conversations/{id}` | Get conversation messages |
| GET | `/api/v1/eval/summary` | Latest evaluation metrics |
| POST | `/api/v1/eval/run` | Trigger evaluation pipeline |

---

## 11. LLM Provider Strategy

```mermaid
flowchart TB
    REQ[LLM Request]
    GROQ{Groq API<br/>Available?}
    GEMINI{Gemini API<br/>Available?}
    OLLAMA{Ollama<br/>Running?}
    ERROR[Return Error<br/>+ Queue for Retry]

    REQ --> GROQ
    GROQ -->|Yes| GROQ_CALL[Groq: Llama 3.3 70B<br/>500+ tok/s, free]
    GROQ -->|No / Rate Limit| GEMINI
    GEMINI -->|Yes| GEMINI_CALL[Gemini Flash<br/>Free backup]
    GEMINI -->|No / Rate Limit| OLLAMA
    OLLAMA -->|Yes| OLLAMA_CALL[Qwen2.5-7B Q4<br/>Local, slow but reliable]
    OLLAMA -->|No| ERROR

    GROQ_CALL --> RESULT[Response]
    GEMINI_CALL --> RESULT
    OLLAMA_CALL --> RESULT
```

### Provider Comparison

| Provider | Model | Speed | Cost | Availability |
|----------|-------|-------|------|-------------|
| **Groq** | Llama 3.3 70B | 500+ tok/s | Free (14.4k req/day) | API, needs internet |
| **Gemini** | Gemini Flash | ~200 tok/s | Free (1M tok/day) | API, needs internet |
| **Ollama** | Qwen2.5-7B Q4 | ~30 tok/s | Free (local) | Local, needs 6GB RAM |

**Strategy:** Groq first (fastest + free), Gemini backup (also free), Ollama offline fallback.

---

## 12. Security Considerations

| Risk | Mitigation |
|------|-----------|
| Prompt Injection | Input sanitization, system prompt isolation, output validation |
| File Upload Attack | File type validation, max size limit (20MB), virus scan |
| API Key Exposure | Environment variables, never commit `.env`, Docker secrets in prod |
| Data Privacy | Per-user document isolation, no cross-user RAG retrieval |
| Rate Limiting | FastAPI rate limiter, respect Groq/Google API limits |
| SQL Injection | SQLAlchemy ORM, parameterized queries |

---

## 13. Monitoring & Observability

| Metric | Tool | Purpose |
|--------|------|---------|
| API Latency (P50/P95/P99) | FastAPI middleware + logging | Performance tracking |
| OCR Processing Time | Custom metrics | Pipeline bottleneck detection |
| LLM Token Usage | Per-request logging | Cost tracking |
| RAG Retrieval Quality | RAGAS periodic eval | Accuracy monitoring |
| Extraction Accuracy | Field-level eval | Model drift detection |
| Error Rate | Sentry / structured logging | Reliability |
| Queue Depth | Redis monitoring | Processing backlog |
