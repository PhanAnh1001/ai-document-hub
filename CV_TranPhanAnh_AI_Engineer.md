# TRAN PHAN ANH — AI Engineer

---

## PERSONAL DETAIL

- **Full name:** Tran Phan Anh
- **Gender:** Male
- **Year of birth:** 1989

---

## SUMMARY

- 12+ years of experience in software engineering and system architecture, with the last 2+ years focusing on applied AI/ML solutions — including NLP, LLM integration, Computer Vision (OCR), and AI-powered automation pipelines.
- Hands-on experience building production AI systems: RAG pipelines, LLM-based document extraction, OCR with PaddleOCR/Google Vision, prompt engineering with GPT-4o and Gemini, and fine-tuning language models for domain-specific tasks.
- Strong background in deploying and scaling AI services using Docker, Kubernetes (GKE, ECS), and serverless architectures on GCP and AWS.
- Experience with the full ML lifecycle: data preprocessing, model training/fine-tuning, evaluation, deployment, and monitoring in production environments.
- Extensive expertise in distributed systems, event-driven architecture, and high-availability systems handling billions of records.
- Proven leadership in cross-functional teams, working closely with Product, Business, and Engineering in Agile environments.

---

## EDUCATION

- **University:** Hanoi University of Science and Technology (HUST)
- **Degree:** Bachelor of Science in Information Technology

## CERTIFICATION

- Top score in entrance exam: HUS High School for Gifted Students (Major: Information Technology)

## LANGUAGE

- Vietnamese: Native
- English: Professional working proficiency

---

## TECHNICAL SKILLS

| Category | Skills |
|----------|--------|
| **AI / ML** | LLM Integration (GPT-4o, Gemini, Llama 3, Qwen), Prompt Engineering, RAG (Retrieval-Augmented Generation), Fine-tuning (LoRA, QLoRA), LangChain, LlamaIndex, Ollama, Hugging Face Transformers |
| **NLP** | Text Classification, Named Entity Recognition, Document Q&A, Summarization, Embedding Models (BGE, Sentence-BERT) |
| **Computer Vision** | PaddleOCR, Google Vision OCR, OpenCV, Document Layout Analysis |
| **ML Frameworks** | PyTorch, TensorFlow (inference & fine-tuning), scikit-learn |
| **Data & Vector DB** | ChromaDB, FAISS, PostgreSQL (pgvector), Redis, MongoDB, MySQL, Firestore |
| **Programming** | Python, Java (Spring Boot, WebFlux), JavaScript/TypeScript, PHP, Node.js, Golang |
| **Cloud & Infra** | GCP (GKE, Cloud Run, Pub/Sub, Cloud SQL, Firestore), AWS (ECS, Lambda, SQS, S3, SageMaker, Bedrock), Terraform, Docker, Kubernetes |
| **MLOps** | Model versioning, A/B testing, evaluation pipelines (RAGAS), CI/CD (GitHub Actions, GitLab CI), monitoring & logging |
| **Architecture** | Microservices, Event-Driven Architecture, Domain-Driven Design, Distributed Systems |

---

## WORKING HISTORY

| Period | Company | Job Title |
|--------|---------|-----------|
| 2019 – present | Hybrid Technologies Vietnam | Solution Architect / AI Lead |
| 2017 – 2019 | Powergate Software | Technical Leader |
| 2013 – 2017 | Mytour.vn | Software Developer |

---

## KEY PROJECTS

### 1. ShareMe — AI-Powered Digital Business Card Platform (2024 – present)

**Company:** Hybrid Technologies Vietnam | **Role:** Solution Architect & AI Lead

**Description:**
Platform for creating and exchanging NFC electronic business cards. Led the AI module development including OCR-based business card scanning and AI-powered conversation summarization.

**AI Responsibilities:**
- Designed and implemented the AI business card scanning pipeline: PaddleOCR for text extraction → LLM (Gemini / GPT-4o mini) with structured prompt engineering for JSON formatting.
- Built a RAG-based Q&A system for users to query their collected business card database using natural language.
- Developed an AI meeting summarizer using Whisper for speech-to-text and GPT-4o for structured summarization with action items extraction.
- Implemented evaluation pipeline using RAGAS metrics to measure RAG retrieval quality and answer accuracy.
- Set up vector storage using ChromaDB for business card embeddings with sentence-transformer models.

**Technologies:**
NestJS, NextJS, Python, MySQL, PaddleOCR, Google Vision OCR, GPT-4o, Gemini, ChromaDB, Whisper, LangChain, AWS ECS Fargate, Lambda, S3, CloudFront.

---

### 2. Gnex Purchase — AI-Enhanced Used Car Deal Management (2024 – present)

**Company:** Hybrid Technologies Vietnam | **Role:** Solution Architect

**Description:**
System managing the deal process between customers and used car shops. Added AI capabilities for document processing and customer intent analysis.

**AI Responsibilities:**
- Built an OCR + LLM pipeline for auto-extracting structured data from car inspection reports (PDF/images) → JSON, reducing manual data entry by 70%.
- Implemented NLP-based customer inquiry classification using fine-tuned BERT model to auto-route deal requests to appropriate departments.
- Developed a document similarity search system using embedding models for finding comparable vehicle deals.

**Technologies:**
PHP-Laravel, Python, MySQL, PaddleOCR, GPT-4o mini, Hugging Face Transformers, FAISS, AWS ECS Fargate, Lambda, SQS, EventBridge.

---

### 3. NOVA — AI-Assisted Language Learning Platform (2024 – present)

**Company:** Hybrid Technologies Vietnam | **Role:** Solution Architect

**Description:**
Online/offline foreign language learning platform. Contributed AI features for personalized learning experience.

**AI Responsibilities:**
- Integrated LLM-powered AI tutor for grammar correction and conversation practice using RAG over the lesson content database.
- Implemented pronunciation evaluation using Whisper speech-to-text + text similarity scoring against reference transcripts.
- Built an automated lesson content tagging system using zero-shot classification with multilingual models.

**Technologies:**
Java, NextJS, Python, Whisper, GPT-4o, LangChain, ChromaDB, AWS EC2, S3, EventBridge.

---

### 4. Mamatenna & Conobie — Content Platform with AI Recommendation (2019 – 2024)

**Company:** Hybrid Technologies Vietnam | **Role:** Solution Architect

**Description:**
Content platform for mothers and children. Led infrastructure migration from AWS to GCP and added AI-powered content features.

**AI Responsibilities:**
- Developed NLP-based article auto-tagging and categorization system using Japanese BERT model, improving content discoverability.
- Built a content recommendation engine using embedding-based similarity search, increasing average session duration by 25%.
- Implemented AI-powered content summarization for article previews using LLM APIs.
- Led full infrastructure migration from AWS to GCP, including Kubernetes to Cloud Run serverless migration.

**Technologies:**
PHP-Laravel, Python, MySQL, Hugging Face Transformers, sentence-transformers, GCP Cloud Run, GCS, Cloud SQL, Kubernetes, Docker.

---

### 5. PPC-ENTOURAGE — Amazon Advertising Analytics with ML (2017 – 2019)

**Company:** Powergate Software | **Role:** Technical Leader

**Description:**
System for collecting and analyzing Amazon advertising campaign data at scale — processing 20M+ records/day with 3B+ total records.

**AI Responsibilities:**
- Developed anomaly detection system using statistical models and scikit-learn to alert on unusual campaign performance changes.
- Built keyword recommendation engine using TF-IDF and embedding similarity for optimizing ad targeting.
- Designed data pipeline architecture handling 20M daily record inserts/updates using AWS Redshift and batch processing.

**Technologies:**
NodeJS, ReactJS, Python, PostgreSQL, scikit-learn, AWS Redshift, SQS, Lambda, S3.

---

## SIDE PROJECTS (Personal / Open Source)

### 1. VietDoc-QA — Vietnamese Document Q&A System

**GitHub:** github.com/phananh/vietdoc-qa

RAG pipeline for answering questions over Vietnamese PDF documents (legal, financial, internal company docs).

**Technical Details:**
- Document ingestion: pdfplumber + PaddleOCR for scanned PDFs → chunking with recursive text splitter.
- Embedding: BGE-M3 multilingual model for Vietnamese text embeddings.
- Vector store: ChromaDB with metadata filtering.
- LLM: Ollama running Qwen2.5-7B locally, with fallback to GPT-4o API.
- Evaluation: RAGAS pipeline measuring faithfulness, answer relevancy, context precision — achieved 0.82 avg score on 200-sample test set.
- Deployment: Docker Compose with Nginx reverse proxy, FastAPI backend.

---

### 2. InvoiceAI — Automated Invoice Extraction Pipeline

**GitHub:** github.com/phananh/invoice-ai

End-to-end pipeline for extracting structured data from Vietnamese invoices (hóa đơn VAT).

**Technical Details:**
- OCR: PaddleOCR with custom preprocessing (deskew, denoise, contrast enhancement using OpenCV).
- Extraction: Fine-tuned Qwen2.5-7B using QLoRA (4-bit) on 500 labeled Vietnamese invoice samples for structured JSON extraction.
- Accuracy: 94% field-level accuracy on test set of 100 invoices (measured per-field: seller name, tax ID, total amount, line items).
- API: FastAPI service with async processing, Redis queue for batch jobs.
- Deployment: Docker, tested on both local GPU (RTX 3060) and Google Colab.

---

### 3. MeetingMind — AI Meeting Summarizer & Action Tracker

**GitHub:** github.com/phananh/meetingmind

Tool for recording, transcribing, and summarizing business meetings with automatic action item extraction.

**Technical Details:**
- Speech-to-text: Whisper (medium model) for Vietnamese + English mixed-language meetings.
- Summarization: LangChain with map-reduce pattern for long meetings (1h+), using GPT-4o.
- Action items: Custom prompt chain for extracting assignee, deadline, and task description.
- Storage: PostgreSQL with pgvector for semantic search over past meeting transcripts.
- Frontend: Simple React dashboard for browsing meetings and tracking action items.
- Fine-tuned a LoRA adapter on 200 meeting transcript samples for better Vietnamese summarization quality.

---

*Updated: April 2026*
