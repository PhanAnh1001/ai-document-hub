# AI Kế Toán

AI agent kế toán tự động cho SME Việt Nam — OCR hóa đơn, định khoản VAS, tích hợp MISA AMIS.

## Tech Stack

- **Frontend**: Next.js 15, TypeScript, Tailwind CSS
- **Backend**: Go 1.24, chi router, pgx
- **Database**: PostgreSQL 16
- **Testing**: Vitest (unit), Playwright (E2E), Go test

---

## Chạy local

### Yêu cầu

| Công cụ | Phiên bản tối thiểu |
|---------|---------------------|
| Node.js | 20+ |
| Go      | 1.24+ |
| PostgreSQL | 16+ |
| Docker (tùy chọn) | 24+ |

---

### Cách 1 — Docker Compose (nhanh nhất)

Chạy toàn bộ stack (frontend + backend + PostgreSQL) chỉ với 1 lệnh:

```bash
docker compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

Dừng và xóa containers:

```bash
docker compose down
```

Xóa cả data volume (reset DB):

```bash
docker compose down -v
```

---

### Cách 2 — Chạy thủ công từng service

#### 1. Tạo database PostgreSQL

```bash
psql -U postgres -c "CREATE DATABASE ai_ke_toan;"
```

#### 2. Setup Backend

```bash
cd backend

# Sao chép file env
cp .env.example .env
```

Chỉnh sửa `backend/.env` — xem phần **Biến môi trường** bên dưới.

Chạy migrations:

```bash
# Dùng psql trực tiếp
psql "$DATABASE_URL" -f migrations/00001_initial_schema.up.sql
psql "$DATABASE_URL" -f migrations/00002_org_integrations.up.sql
psql "$DATABASE_URL" -f migrations/00003_fix_entry_status_constraint.up.sql
psql "$DATABASE_URL" -f migrations/00004_entries_updated_at.up.sql
psql "$DATABASE_URL" -f migrations/00005_entries_reject_reason.up.sql
```

Khởi động backend:

```bash
go run ./cmd/server
```

Backend chạy tại http://localhost:8080.

#### 3. Setup Frontend

```bash
cd app

# Sao chép file env
cp .env.example .env.local
```

Chỉnh sửa `app/.env.local` — xem phần **Biến môi trường** bên dưới.

Cài dependencies và chạy:

```bash
npm install
npm run dev
```

Frontend chạy tại http://localhost:3000.

---

## Biến môi trường

### Backend (`backend/.env`)

| Biến | Bắt buộc | Mặc định | Mô tả |
|------|----------|----------|-------|
| `PORT` | | `8080` | Port backend lắng nghe |
| `DATABASE_URL` | ✅ | | PostgreSQL connection string |
| `JWT_SECRET` | ✅ | | Secret key JWT, tối thiểu 32 ký tự |
| `CORS_ORIGINS` | | `http://localhost:3000` | Origin được phép gọi API |
| `OCR_PROVIDER` | | `mock` | `mock` (dev), `google_vision` (free), hoặc `fpt` (production VN) |
| `OCR_API_KEY` | | | API key tương ứng với `OCR_PROVIDER` |
| `ANTHROPIC_API_KEY` | | | Claude API key |
| `MISA_API_URL` | | | MISA AMIS API endpoint |
| `MISA_API_KEY` | | | MISA AMIS API key |
| `RESEND_API_KEY` | | | Resend email API key |
| `SENTRY_DSN` | | | Sentry DSN (error monitoring) |

**Cấu hình tối thiểu để chạy local:**

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/ai_ke_toan?sslmode=disable
JWT_SECRET=dev-secret-key-change-in-production-32chars
CORS_ORIGINS=http://localhost:3000
OCR_PROVIDER=mock
```

> **VAS Rule Engine không cần API key.** Toàn bộ logic định khoản (10 VAS rules) chạy 100% local — pure Go, không có network call. Với `OCR_PROVIDER=mock`, pipeline đầy đủ hoạt động mà không cần bất kỳ external service nào: Upload → Mock OCR → VAS Rules → Định khoản pending.

### Frontend (`app/.env.local`)

| Biến | Bắt buộc | Mặc định | Mô tả |
|------|----------|----------|-------|
| `NEXT_PUBLIC_API_URL` | ✅ | `http://localhost:8080` | URL backend API |
| `NEXT_PUBLIC_SENTRY_DSN` | | | Sentry DSN (frontend) |
| `SENTRY_AUTH_TOKEN` | | | Sentry auth token (chỉ cần khi build) |

**Cấu hình tối thiểu:**

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## OCR — Google Vision (free tier)

Google Cloud Vision API có **1,000 units miễn phí/tháng** (mỗi trang = 1 unit). Phù hợp để test thực tế mà không tốn tiền.

### Giới hạn free tier

| | Google Vision | Viettel AI | FPT.AI |
|---|---|---|---|
| Free quota | 1,000 units/tháng | Không có (trả phí) | Không có (trả phí) |
| Loại OCR | General purpose | Chuyên biệt hóa đơn VN | Chuyên biệt hóa đơn VN |
| Ngôn ngữ | 50+ ngôn ngữ | Tiếng Việt tối ưu | Tiếng Việt tối ưu |
| Định dạng | JPG, PNG, PDF, GIF, BMP, WEBP | JPG, PNG, PDF | JPG, PNG, PDF |
| Kết quả | Raw text → parser tự động | Structured JSON (VN invoice fields) | Structured JSON (seller, amount, date) |
| Phù hợp | Dev/test, đa ngôn ngữ | Production VN, tích hợp hệ sinh thái Viettel | Production VN, hóa đơn điện tử |
| `OCR_PROVIDER` | `google_vision` | `viettel` | `fpt` |

### Bước 1 — Tạo Google Cloud project và bật Vision API

1. Truy cập [Google Cloud Console](https://console.cloud.google.com/)
2. Tạo project mới (hoặc chọn project có sẵn)
3. Vào **APIs & Services → Library**, tìm **"Cloud Vision API"**, bấm **Enable**
4. Vào **APIs & Services → Credentials**
5. Bấm **Create Credentials → API key**
6. Copy API key vừa tạo

> **Bảo mật:** Nên restrict API key chỉ cho Cloud Vision API:
> Credentials → Edit key → API restrictions → Restrict key → chọn **Cloud Vision API**

### Bước 2 — Kiểm tra billing (bắt buộc, dù dùng free)

Google yêu cầu bật billing account mới dùng được Vision API, kể cả trong free tier.

1. Vào **Billing** trong Cloud Console
2. Liên kết billing account với project (có thể dùng thẻ Visa/Mastercard — sẽ không bị charge nếu dưới 1,000 units/tháng)

### Bước 3 — Cấu hình backend

Chỉnh `backend/.env`:

```env
OCR_PROVIDER=google_vision
OCR_API_KEY=AIzaSy...your-key-here
```

Khởi động lại backend:

```bash
go run ./cmd/server
```

### Bước 4 — Test thử

Upload một ảnh hóa đơn qua trang `/documents`. Backend sẽ:

1. Gửi ảnh (base64) lên `https://vision.googleapis.com/v1/images:annotate`
2. Nhận full text từ `DOCUMENT_TEXT_DETECTION`
3. Tự động parse: tên công ty, tổng tiền, thuế GTGT, ngày hóa đơn
4. Áp VAS rule engine → tạo định khoản kế toán

### Theo dõi quota

Kiểm tra số lượng đã dùng:
**Google Cloud Console → APIs & Services → Cloud Vision API → Metrics**

---

## OCR — Viettel AI

Viettel AI OCR chuyên biệt cho hóa đơn tiếng Việt, độ chính xác ~99% cho văn bản in.

### Bước 1 — Đăng ký tài khoản

1. Truy cập [viettelgroup.ai](https://viettelgroup.ai)
2. Đăng ký tài khoản → đăng nhập
3. Vào menu **Token** → tạo token mới
4. Copy token vừa tạo

Hoặc liên hệ trực tiếp để được cấp enterprise account:
- Email: viettelai@viettel.com.vn
- Hotline: +84 98 1900 911

### Bước 2 — Cấu hình backend

Chỉnh `backend/.env`:

```env
OCR_PROVIDER=viettel
OCR_API_KEY=your-viettel-token-here
```

Khởi động lại backend:

```bash
go run ./cmd/server
```

### Bước 3 — Test thử

Upload một ảnh hóa đơn qua trang `/documents`. Backend sẽ:

1. Gửi file (multipart/form-data) lên `https://viettelgroup.ai/cv/api/v1/ocr`
2. Auth bằng header `token: <api_key>`
3. Nhận full text từ response
4. Tự động parse: tên công ty, tổng tiền, thuế GTGT, ngày hóa đơn
5. Áp VAS rule engine → tạo định khoản kế toán

### Định dạng file hỗ trợ

`BMP`, `PNG`, `JPG`, `JPEG`, `TIF`, `TIFF`, `PDF`

---

## Chạy tests

```bash
# Frontend unit tests (Vitest)
cd app && npm test

# Frontend test coverage
cd app && npm run coverage

# E2E tests (Playwright) — cần frontend + backend đang chạy
cd app && npm run test:e2e

# Backend unit tests
cd backend && go test ./...

# Backend với race detection + coverage
cd backend && go test -race -cover ./...
```

---

## Cấu trúc thư mục

```
ai-accounting/
├── app/                    # Next.js frontend
│   ├── src/
│   │   ├── app/           # App router pages
│   │   ├── components/    # React components
│   │   └── lib/           # API client, utilities
│   └── e2e/               # Playwright E2E tests
├── backend/                # Go backend
│   ├── cmd/server/        # Entry point
│   ├── internal/          # Business logic, handlers, services
│   └── migrations/        # PostgreSQL migrations (SQL)
└── docker-compose.yml
```
