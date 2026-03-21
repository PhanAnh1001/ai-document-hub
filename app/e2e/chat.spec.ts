import { test, expect } from "@playwright/test";

// Helper: register a new account and log in so we land on /dashboard
async function registerAndLogin(page: import("@playwright/test").Page) {
  const email = `chat_e2e_${Date.now()}@example.com`;
  await page.goto("/register");
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/mật khẩu/i).fill("Password123!");
  await page.getByLabel(/họ tên/i).fill("Chat E2E User");
  await page.getByLabel(/tên công ty/i).fill("Chat E2E Co");
  await page.getByRole("button", { name: /đăng ký/i }).click();
  await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });
}

test.describe("Chat page", () => {
  test("chat page loads and shows input", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/chat");

    await expect(page.getByRole("heading", { name: /chat ai/i })).toBeVisible();
    await expect(page.getByRole("textbox", { name: /nhập câu hỏi/i })).toBeVisible();
  });

  test("send button is disabled when input is empty", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/chat");

    await expect(page.getByRole("textbox", { name: /nhập câu hỏi/i })).toBeVisible();
    const sendBtn = page.getByRole("button", { name: /gửi tin nhắn/i });
    await expect(sendBtn).toBeDisabled();
  });

  test("shows example prompts in empty state", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/chat");

    await expect(
      page.getByText(/bắt đầu đặt câu hỏi về tài liệu/i)
    ).toBeVisible();
  });

  test("clicking example prompt fills the input", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/chat");

    // Wait for empty state with example prompts
    await expect(
      page.getByText(/hợp đồng này có những điều khoản/i)
    ).toBeVisible({ timeout: 5000 });

    await page.getByText(/hợp đồng này có những điều khoản/i).click();

    const textarea = page.getByRole("textbox", { name: /nhập câu hỏi/i });
    await expect(textarea).not.toBeEmpty();
  });

  test("send message shows user message in chat", async ({ page }) => {
    await registerAndLogin(page);

    // Mock the RAG API at the route level using Playwright route interception
    await page.route("**/api/v1/query/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          answer: "Đây là câu trả lời thử nghiệm từ AI.",
          sources: [
            {
              doc_id: "test-doc-001",
              chunk_text: "Đoạn văn bản mẫu từ tài liệu.",
              score: 0.88,
            },
          ],
        }),
      });
    });

    await page.goto("/chat");
    await expect(page.getByRole("textbox", { name: /nhập câu hỏi/i })).toBeVisible();

    await page.getByRole("textbox", { name: /nhập câu hỏi/i }).fill(
      "Hợp đồng này có giá trị bao nhiêu?"
    );
    await page.getByRole("button", { name: /gửi tin nhắn/i }).click();

    // User message should appear
    await expect(
      page.getByText("Hợp đồng này có giá trị bao nhiêu?")
    ).toBeVisible({ timeout: 5000 });
  });

  test("send message shows AI answer with sources", async ({ page }) => {
    await registerAndLogin(page);

    await page.route("**/api/v1/query/**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          answer: "Giá trị hợp đồng là 500 triệu đồng.",
          sources: [
            {
              doc_id: "doc-abc123",
              chunk_text: "Điều 5: Giá trị hợp đồng là 500.000.000 VNĐ",
              score: 0.95,
            },
          ],
        }),
      });
    });

    await page.goto("/chat");
    await expect(page.getByRole("textbox", { name: /nhập câu hỏi/i })).toBeVisible();

    await page.getByRole("textbox", { name: /nhập câu hỏi/i }).fill(
      "Giá trị hợp đồng là bao nhiêu?"
    );
    await page.getByRole("button", { name: /gửi tin nhắn/i }).click();

    // Answer should appear
    await expect(
      page.getByText("Giá trị hợp đồng là 500 triệu đồng.")
    ).toBeVisible({ timeout: 8000 });

    // Sources section
    await expect(page.getByText(/nguồn tham khảo/i)).toBeVisible();
  });
});
