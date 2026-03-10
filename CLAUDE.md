# Project: {project name}

## Context
- Solo founder (AnhTP), bootstrap
- SaaS: ...
- Stack: Next.js + Go + PostgreSQL + Cloudflare + Amazon Lightsail + Python
- Tích hợp: Google OCR, ...

## Glossary
...

## Quy tắc cơ bản
- Luôn review lại trước khi trả kết quả cuối
- Trả lời tiếng Việt, comment code bằng tiếng Anh
- Thẳng thắn, không nói chung chung

## Skills & Evals (BẮT BUỘC)

Skill-creator **không dùng eval mặc định** — phải define thủ công.

Khi tạo skill:
1. **Design trước** — lưu vào `design/skills/<agent>/<NN>-<skill>.md` TRƯỚC khi implement
2. **Evals bắt buộc** — mỗi design file phải có ≥3 happy path + ≥2 edge case + ≥1 API error case
3. **Format eval**: mỗi case có `id`, `description`, `input`, `expected`, `pass_criteria`
4. **Không ship skill chưa có evals**

Cấu trúc:
```
design/skills/<agent>/<NN>-<skill>.md
design/skills/<agent>/README.md
.claude/skills/<skill>/SKILL.md
.claude/skills/<skill>/evals/evals.json
```

## Testing (BẮT BUỘC)

**TDD flow**: RED (viết test fail) → GREEN (code tối thiểu) → REFACTOR

- Frontend unit: Vitest — `cd app && npm test`
- Frontend E2E: Playwright — `cd app && npm run test:e2e`
- Backend: `cd backend && go test -race -cover ./...`
- Test file đặt cạnh source: `foo.ts` → `foo.test.ts`
- **Test PHẢI pass trước khi commit**

## Session Workflow (BẮT BUỘC)

### Bắt đầu session
1. Đọc `plans/plan.md` — load context
2. Checkout đúng branch (không tạo mới nếu đã có)

### Kết thúc session
1. Commit + push tất cả thay đổi
2. Tạo PR vào `master` nếu branch chưa có PR
3. Cập nhật `plans/plan.md` (chỉ khi có thay đổi thực):
   - TODO: thêm/xóa tasks
   - Notes: cập nhật thông tin kỹ thuật quan trọng
   - Lịch sử: gộp session mới vào milestone tương ứng (1 dòng/milestone, KHÔNG liệt kê sub-tasks)
4. Commit + push `plans/plan.md`

### Format plan.md (tối ưu token)
```
## Trạng thái hiện tại   ← phase + branch hiện tại
## TODO                  ← flat list, ưu tiên cao lên trên
## Notes                 ← key technical refs, env vars, paths
## Lịch sử (milestone)  ← 1 dòng/milestone, KHÔNG sub-tasks
```
> File DUY NHẤT — không tạo thêm. Lịch sử chi tiết ở `git log`.
