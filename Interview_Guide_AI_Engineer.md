# Hướng dẫn trả lời phỏng vấn — Vị trí AI Engineer

---

## Chiến lược tổng thể

**Positioning của anh:** Anh không phải ML Researcher hay Data Scientist. Anh là **Applied AI Engineer** với nền tảng mạnh về system architecture + deployment. Đây là điểm bán:

> "Nhiều AI Engineer biết train model nhưng không biết deploy production. Tôi biết cả hai — và tôi đã deploy hệ thống xử lý 3 tỷ records."

**3 trụ cột khi trả lời:**
1. Tôi đã build AI system thực tế (ShareMe OCR, RAG, Meeting Summarizer)
2. Tôi biết deploy và scale (12 năm kinh nghiệm infrastructure)
3. Tôi tự học nhanh và có side project chứng minh

---

## PHẦN 1: Câu hỏi về bản thân & kinh nghiệm

### Q: "Giới thiệu về bản thân?"

> Tôi có 12 năm kinh nghiệm software engineering, trong đó 9 năm làm leader/architect. 2 năm gần đây tôi chuyển trọng tâm sang AI Engineering — bắt đầu từ project ShareMe khi tôi tự nghiên cứu và build hệ thống OCR + LLM cho scan business card. Từ đó tôi đã mở rộng sang RAG pipeline, fine-tuning model, và build thêm 3 side project AI cá nhân. Background system architecture giúp tôi không chỉ prototype mà còn deploy AI vào production thực tế — Docker, Kubernetes, CI/CD, monitoring — toàn bộ lifecycle.

### Q: "Tại sao muốn chuyển sang AI?"

> Trong quá trình làm Solution Architect, tôi thấy rằng phần lớn business logic truyền thống đang dần được AI thay thế hoặc augment. Khi làm project ShareMe, tôi phải build OCR + LLM pipeline từ đầu — lúc đó tôi nhận ra đây là hướng đi tiếp theo. Tôi không phải chuyển nghề — tôi mở rộng skill set. Một AI Engineer mà không biết deploy, scale, monitor thì model chỉ nằm trên notebook. Tôi cover được phần đó.

### Q: "Điểm yếu của anh trong AI/ML?"

**Trả lời thành thật nhưng có chiến lược:**

> Tôi chưa có kinh nghiệm train model from scratch ở quy mô lớn — ví dụ pre-train một LLM. Nhưng thực tế trong công việc AI Engineer, rất hiếm khi cần làm điều đó. Tôi focus vào fine-tuning, RAG, và prompt engineering — là những thứ tạo giá trị business nhanh nhất. Phần train model from scratch, tôi đang tự học thêm qua PyTorch tutorials và Hugging Face course.

---

## PHẦN 2: Câu hỏi kỹ thuật AI/ML

### Q: "Giải thích RAG là gì? Anh đã implement như thế nào?"

> RAG — Retrieval-Augmented Generation — là pattern kết hợp search với LLM generation. Thay vì để LLM trả lời từ training data (dễ hallucinate), mình retrieve context liên quan từ knowledge base trước, rồi đưa context đó vào prompt cho LLM generate câu trả lời.
>
> **Pipeline tôi build cho VietDoc-QA:**
> 1. **Ingestion:** PDF → pdfplumber extract text (hoặc PaddleOCR nếu scan) → recursive text splitter chia chunk 512 tokens, overlap 50 tokens.
> 2. **Embedding:** Dùng BGE-M3 — model multilingual support tiếng Việt tốt. Mỗi chunk được embed thành vector 1024 chiều.
> 3. **Storage:** ChromaDB — đủ nhẹ cho use case vài nghìn documents. Nếu scale lớn hơn thì chuyển sang pgvector hoặc Qdrant.
> 4. **Retrieval:** Query embedding → cosine similarity search → top-k chunks (k=5). Có thêm metadata filter (theo document, theo ngày).
> 5. **Generation:** Chunks + query → prompt template → Qwen2.5-7B qua Ollama. Prompt có instruction rõ: "Chỉ trả lời dựa trên context, nếu không có thông tin thì nói không biết."
> 6. **Evaluation:** RAGAS framework đo 4 metrics: faithfulness, answer relevancy, context precision, context recall.

**Câu follow-up thường gặp:**

**"Chunk size bao nhiêu là tốt?"**
> Không có con số magic. Tôi thường bắt đầu 512 tokens, overlap 50-100. Chunk quá nhỏ thì mất context, quá lớn thì retrieval kém chính xác. Phải thử nghiệm và đo bằng evaluation metrics.

**"Làm sao xử lý khi context window LLM không đủ?"**
> Có vài cách: (1) Giảm k — lấy ít chunk hơn nhưng relevant hơn, dùng re-ranking. (2) Map-reduce — summarize từng chunk rồi combine. (3) Dùng model context window lớn hơn (Qwen2.5 có 128K context).

**"So sánh ChromaDB vs FAISS vs pgvector?"**
> - ChromaDB: Dễ dùng, có metadata filtering, tốt cho prototype và project nhỏ-trung.
> - FAISS: Facebook build, nhanh nhất cho pure vector search, nhưng không có metadata filtering built-in. Tốt khi cần search hàng triệu vectors.
> - pgvector: Tích hợp vào PostgreSQL — tốt khi đã có Postgres stack, cần ACID, cần SQL query kết hợp vector search.

---

### Q: "Fine-tuning là gì? Anh đã fine-tune model nào?"

> Fine-tuning là tiếp tục train một pre-trained model trên dataset domain-specific để nó perform tốt hơn cho task cụ thể. Khác với prompt engineering (không thay đổi model weights), fine-tuning thực sự update weights.
>
> **Tôi đã fine-tune trong 2 case:**
>
> **Case 1 — InvoiceAI:** Fine-tune Qwen2.5-7B bằng QLoRA (quantized LoRA, 4-bit) trên 500 mẫu hóa đơn Việt Nam. Input là raw OCR text, output là structured JSON. Dùng Hugging Face Transformers + PEFT library. Training trên Google Colab T4 GPU, khoảng 3 epoch, mất ~2 giờ.
>
> **Case 2 — MeetingMind:** Fine-tune LoRA adapter trên 200 meeting transcript samples cho Vietnamese summarization. Mục đích là model tóm tắt đúng format (bullet points, action items) thay vì paragraph dài.

**Câu follow-up thường gặp:**

**"LoRA vs full fine-tuning?"**
> LoRA (Low-Rank Adaptation) chỉ train thêm một tập matrix nhỏ gắn vào model gốc, không update toàn bộ weights. Ưu điểm: cần ít GPU (7B model fine-tune được trên 1 GPU 16GB), nhanh, model gốc không bị thay đổi. Nhược điểm: không mạnh bằng full fine-tuning nếu task quá khác so với pre-training data. Thực tế, LoRA đủ tốt cho 90% use case.

**"QLoRA khác gì LoRA?"**
> QLoRA = LoRA + quantization. Model gốc được quantize xuống 4-bit (giảm ~4x memory), rồi train LoRA adapters ở FP16. Kết quả gần bằng LoRA full precision nhưng chạy được trên GPU nhỏ hơn nhiều — ví dụ fine-tune 7B model trên T4 16GB.

**"Evaluation fine-tuned model như thế nào?"**
> Chia dataset: 80% train, 10% validation, 10% test. Metrics tùy task:
> - Extraction (InvoiceAI): field-level accuracy — đo từng field (tên, mã số thuế, tổng tiền) có đúng không.
> - Summarization (MeetingMind): ROUGE score + human evaluation (tôi tự đánh giá 50 samples).
> - Quan trọng: phải có test set KHÔNG overlap với training data.

---

### Q: "Anh biết gì về Computer Vision?"

> Kinh nghiệm chính của tôi ở CV là **OCR pipeline** — cụ thể:
>
> - **PaddleOCR:** Dùng cho tiếng Việt, tiếng Nhật. Pipeline: image preprocessing (deskew, denoise, contrast bằng OpenCV) → text detection → text recognition. PaddleOCR cho accuracy tốt với tiếng Việt hơn Tesseract.
> - **Google Vision OCR:** Dùng khi cần accuracy cao nhất và chấp nhận chi phí API. Handwriting recognition tốt hơn PaddleOCR.
> - **Document Layout Analysis:** Xác định vùng text, table, image trong document trước khi OCR — dùng Docling (IBM) hoặc LayoutLM.

**Nếu bị hỏi sâu hơn về CV (object detection, segmentation):**

> Tôi có kiến thức cơ bản về YOLO, CNN architecture, nhưng chưa deploy production system cho object detection. Focus hiện tại của tôi là document AI — OCR, layout analysis, document understanding. Nếu cần, tôi có thể học thêm vì framework như Ultralytics YOLO đã abstract hóa rất nhiều.

---

### Q: "Prompt Engineering — anh dùng kỹ thuật gì?"

> Tôi dùng nhiều kỹ thuật tùy task:
>
> 1. **Structured output prompting:** Cho LLM ví dụ JSON output mong muốn, kèm schema. Dùng cho extraction tasks (hóa đơn, business card).
> 2. **Few-shot prompting:** Cho 3-5 examples input-output trong prompt. Hiệu quả cho classification, NER.
> 3. **Chain-of-thought:** Yêu cầu model giải thích reasoning step-by-step trước khi cho answer. Dùng cho complex reasoning.
> 4. **System prompt design:** Define role, constraints, output format rõ ràng. Ví dụ: "You are a document extraction expert. Extract ONLY information present in the text. If a field is not found, return null."
> 5. **Prompt chaining:** Chia task phức tạp thành multiple LLM calls nối tiếp. Ví dụ: step 1 extract raw fields → step 2 validate & normalize → step 3 format output.

**Ví dụ thực tế từ ShareMe:**

> Business card OCR text thường bị lẫn lộn (tên, chức vụ, email, phone xen kẽ). Tôi dùng structured prompt với GPT-4o mini:
> - System prompt define exact JSON schema
> - Few-shot: 3 examples card text → JSON
> - Constraint: "If multiple phone numbers exist, return as array"
> - Accuracy đạt ~92% trên 200 mẫu test thực tế

---

### Q: "Docker/Kubernetes cho AI deployment — khác gì so với deploy app thường?"

**Đây là điểm mạnh — tận dụng tối đa:**

> Khác biệt chính:
>
> 1. **Image size:** AI model Docker image thường rất nặng (5-20GB vs 200MB cho web app). Cần multi-stage build, chỉ copy model weights và inference code vào final image.
> 2. **Resource allocation:** AI service cần GPU scheduling trong Kubernetes (nvidia device plugin, resource limits). CPU inference cần nhiều RAM hơn (7B model cần ~6GB RAM ở 4-bit quantization).
> 3. **Scaling pattern:** AI inference thường latency cao (1-10s per request), nên cần async processing + queue (Redis/SQS) thay vì scale horizontally như web app.
> 4. **Health checks:** Liveness probe phải account cho model loading time (có thể mất 30-60s load model vào memory).
> 5. **Cost optimization:** Dùng spot instances/preemptible VMs cho batch processing. Inference service dùng auto-scaling dựa trên queue depth thay vì CPU usage.

---

### Q: "LangChain vs LlamaIndex — khi nào dùng cái nào?"

> - **LangChain:** Framework tổng quát cho LLM applications. Mạnh ở chaining, agents, tool use. Dùng khi cần build complex workflows — ví dụ agent tự quyết định search web, query database, hay gọi API.
> - **LlamaIndex:** Chuyên sâu về data ingestion + RAG. Mạnh ở phần index, retrieve, query over documents. Dùng khi focus là document Q&A, knowledge base.
> - **Thực tế:** Tôi dùng LlamaIndex cho RAG pipeline (VietDoc-QA) vì nó handle chunking, embedding, retrieval tốt hơn. Dùng LangChain cho MeetingMind vì cần chain nhiều steps (transcribe → summarize → extract action items).
> - **Xu hướng:** Cả hai đang converge — LangChain thêm RAG features, LlamaIndex thêm agent features. Nhiều project mới dùng trực tiếp API + custom code thay vì framework.

---

## PHẦN 3: Câu hỏi system design

### Q: "Design một hệ thống Document Q&A cho công ty 500 người"

**Cách trả lời:** Vẽ architecture trên giấy/whiteboard, giải thích từng component.

> **Requirements clarification:**
> - Bao nhiêu documents? → Giả sử 10,000 documents, tổng 50,000 pages.
> - Loại documents? → PDF, Word, scan.
> - Concurrent users? → 50-100 cùng lúc.
> - Latency requirement? → Dưới 10 giây cho mỗi câu hỏi.
> - Security? → Documents nội bộ, cần access control.
>
> **Architecture:**
>
> ```
> [User] → [Web UI / API Gateway]
>            ↓
> [FastAPI Backend]
>   ├── /upload → [Document Processor]
>   │                ├── PDF → pdfplumber
>   │                ├── Scan → PaddleOCR
>   │                ├── Word → python-docx
>   │                ↓
>   │            [Chunker] → [Embedder (BGE-M3)] → [Vector DB (pgvector)]
>   │
>   └── /query → [Query Handler]
>                  ├── Embed query → Vector search (top-k)
>                  ├── Re-ranker (cross-encoder)
>                  ├── Access control filter
>                  ↓
>                [LLM (GPT-4o / Qwen)] → [Response + Citations]
> ```
>
> **Key decisions:**
> - **pgvector thay vì ChromaDB:** Vì 500 users cần multi-tenancy, access control, ACID compliance → Postgres phù hợp hơn.
> - **Re-ranker:** Sau retrieval, dùng cross-encoder model re-rank top 20 → top 5 chunks. Tăng precision đáng kể.
> - **Async processing:** Upload document → queue (Redis) → worker process → notify user khi done. Không block API.
> - **Caching:** Cache frequent queries + embeddings trong Redis. LLM response cache cho identical queries.
> - **Monitoring:** Track retrieval quality metrics, latency, token usage, user satisfaction ratings.

---

## PHẦN 4: Câu hỏi về evaluation

### Q: "Làm sao đánh giá chất lượng AI system?"

> Tùy loại task:
>
> **RAG System:**
> - **Faithfulness:** Câu trả lời có đúng với context retrieved không? (không hallucinate)
> - **Answer Relevancy:** Câu trả lời có liên quan đến câu hỏi không?
> - **Context Precision:** Chunks retrieved có chứa thông tin cần thiết không?
> - **Context Recall:** Có bỏ sót thông tin quan trọng không?
> - Tool: RAGAS framework tự động đo 4 metrics này.
>
> **Extraction (OCR + LLM):**
> - **Field-level accuracy:** Mỗi field extracted (tên, số, ngày) so sánh với ground truth.
> - **Exact match rate:** Bao nhiêu % documents extract hoàn toàn đúng tất cả fields.
> - Tôi dùng 100-200 labeled samples, chia 80/20 train/test.
>
> **Classification:**
> - Precision, Recall, F1-score per class.
> - Confusion matrix để thấy model nhầm classes nào.
>
> **Production monitoring:**
> - User feedback (thumbs up/down).
> - Latency P50, P95, P99.
> - Token usage và cost per query.
> - Drift detection: accuracy giảm theo thời gian khi data distribution thay đổi.

---

## PHẦN 5: Câu hỏi behavioral

### Q: "Kể về một thử thách kỹ thuật khó nhất?"

> Project ShareMe — khi build OCR cho business card tiếng Nhật. Thử thách: business card Nhật có layout phức tạp (vertical text, multiple fonts, decorative elements). PaddleOCR default accuracy chỉ ~70%.
>
> **Cách giải quyết:**
> 1. Thêm image preprocessing: deskew, adaptive thresholding, crop ROI.
> 2. So sánh PaddleOCR vs Google Vision OCR — Google Vision tốt hơn cho tiếng Nhật nhưng tốn tiền.
> 3. Quyết định dùng hybrid: PaddleOCR trước, nếu confidence score < threshold thì fallback sang Google Vision.
> 4. LLM prompt engineering để "sửa" OCR errors — ví dụ nếu phone number bị miss digit, LLM có thể infer từ pattern.
> 5. Kết quả: accuracy tăng từ 70% lên 92%.

### Q: "Làm sao anh keep up với AI thay đổi nhanh?"

> - Đọc hàng ngày: Hugging Face blog, Anthropic blog, arXiv papers (qua Twitter/X summaries).
> - Thử ngay: Mỗi model mới release, tôi pull về Ollama test trong 30 phút. So sánh với model cũ trên cùng benchmark prompts.
> - Side projects: Mỗi kỹ thuật mới tôi học, tôi apply vào side project. VietDoc-QA ban đầu dùng LangChain, sau chuyển sang LlamaIndex, gần đây thử dùng vanilla code không framework.
> - Community: Tham gia Discord/Slack groups về AI Engineering, đọc discussions về production issues.

---

## PHẦN 6: Những thứ CẦN BIẾT nhưng không cần giỏi

Đây là những chủ đề có thể bị hỏi surface-level. Biết đủ để trả lời, không cần deep dive:

### Transformer Architecture
> Self-attention mechanism cho phép model "nhìn" tất cả tokens trong input cùng lúc (khác RNN xử lý tuần tự). Multi-head attention = nhiều "perspectives" khác nhau. Positional encoding thêm thông tin vị trí vì attention không có khái niệm thứ tự.

### Embedding là gì?
> Biểu diễn text dưới dạng dense vector trong không gian nhiều chiều. Texts có nghĩa tương tự sẽ có vectors gần nhau (cosine similarity cao). Model embedding phổ biến: sentence-transformers, BGE, OpenAI text-embedding-3.

### Tokenization
> Chia text thành tokens (subwords). BPE (Byte Pair Encoding) là phương pháp phổ biến nhất. Tiếng Việt tokenize phức tạp hơn tiếng Anh vì dấu và từ ghép.

### Hallucination và cách giảm
> LLM "bịa" thông tin không có trong training data. Giảm bằng: RAG (ground vào documents thực), temperature thấp (0.0-0.3), prompt instruction rõ ràng ("chỉ trả lời dựa trên context"), citation/source tracking.

### RLHF
> Reinforcement Learning from Human Feedback — phương pháp align model với human preferences. Gồm 3 bước: supervised fine-tuning → train reward model từ human rankings → optimize policy bằng PPO. Tôi chưa tự làm RLHF, nhưng hiểu concept.

---

## PHẦN 7: Red flags — Câu hỏi nguy hiểm và cách xử lý

### "Anh đã train model from scratch chưa?"
> **Thành thật:** "Chưa pre-train from scratch. Tôi focus vào fine-tuning và applied AI. Pre-training cần hàng nghìn GPU hours và dataset hàng TB — đó là việc của research lab. Công việc của AI Engineer là tận dụng pre-trained models hiệu quả nhất."

### "Giải thích backpropagation?"
> Thuật toán tính gradient của loss function theo weights, lan truyền ngược từ output layer về input layer bằng chain rule. Dùng gradient này để update weights qua optimizer (SGD, Adam). Biết concept nhưng hàng ngày tôi dùng PyTorch autograd — framework handle phần này.

### "Viết code PyTorch training loop ngay bây giờ?"
> **Chuẩn bị trước — học thuộc basic training loop:**
> ```python
> model = AutoModelForCausalLM.from_pretrained("model_name")
> optimizer = torch.optim.AdamW(model.parameters(), lr=2e-5)
>
> for epoch in range(num_epochs):
>     for batch in dataloader:
>         outputs = model(**batch)
>         loss = outputs.loss
>         loss.backward()
>         optimizer.step()
>         optimizer.zero_grad()
> ```
> Thực tế fine-tuning, tôi dùng Hugging Face Trainer hoặc PEFT library — abstract hóa training loop.

### "Kinh nghiệm TensorFlow?"
> "Tôi chủ yếu dùng PyTorch ecosystem vì community AI Engineer đang lean về PyTorch. TensorFlow tôi dùng cho inference — load SavedModel, TFLite cho mobile deployment. Nếu project yêu cầu TensorFlow, tôi có thể chuyển đổi vì concepts giống nhau."

---

## PHẦN 8: Checklist trước phỏng vấn

- [ ] Chạy lại 3 side projects, đảm bảo chạy được, có README rõ ràng
- [ ] Push code lên GitHub, có commit history tự nhiên (không push 1 lần)
- [ ] Chuẩn bị demo VietDoc-QA — impressive nhất khi show live
- [ ] Ôn lại RAGAS metrics — biết giải thích từng metric
- [ ] Chạy thử fine-tuning Qwen2.5-7B bằng QLoRA trên Colab ít nhất 1 lần
- [ ] Đọc OWASP Top 10 for LLM — hay bị hỏi về prompt injection
- [ ] Chuẩn bị 2-3 câu hỏi ngược cho interviewer (thể hiện chủ động):
  - "Team đang dùng model nào? Self-hosted hay API?"
  - "Evaluation pipeline hiện tại như thế nào?"
  - "Biggest challenge khi deploy AI vào production ở công ty?"

---

*Cập nhật: Tháng 4/2026*
