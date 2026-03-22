import { test, expect } from "@playwright/test";

async function registerAndLogin(page: import("@playwright/test").Page) {
  const email = `eval_e2e_${Date.now()}@example.com`;
  await page.goto("/register");
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/mật khẩu/i).fill("Password123!");
  await page.getByLabel(/họ tên/i).fill("Eval E2E User");
  await page.getByLabel(/tên công ty/i).fill("Eval E2E Co");
  await page.getByRole("button", { name: /đăng ký/i }).click();
  await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });
}

test.describe("Evaluation page", () => {
  test("evaluation page loads", async ({ page }) => {
    await registerAndLogin(page);

    // Mock eval stats API
    await page.route("**/api/v1/eval/stats**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          faithfulness: 0.87,
          answer_relevancy: 0.76,
          context_precision: 0.91,
          total_documents: 15,
        }),
      });
    });

    await page.goto("/evaluation");

    await expect(
      page.getByRole("heading", { name: /đánh giá/i })
    ).toBeVisible();
  });

  test("evaluation page shows stat cards", async ({ page }) => {
    await registerAndLogin(page);

    await page.route("**/api/v1/eval/stats**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          faithfulness: 0.87,
          answer_relevancy: 0.76,
          context_precision: 0.91,
          total_documents: 15,
        }),
      });
    });

    await page.goto("/evaluation");

    // Total documents card
    await expect(page.getByText("15")).toBeVisible({ timeout: 8000 });

    // Metric labels
    await expect(page.getByText("Faithfulness").first()).toBeVisible();
    await expect(page.getByText("Answer Relevancy").first()).toBeVisible();
    await expect(page.getByText("Context Precision").first()).toBeVisible();
  });

  test("run evaluation button is visible and clickable", async ({ page }) => {
    await registerAndLogin(page);

    await page.route("**/api/v1/eval/stats**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          faithfulness: 0,
          answer_relevancy: 0,
          context_precision: 0,
          total_documents: 0,
        }),
      });
    });

    await page.route("**/api/v1/eval/run**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ message: "Đánh giá đã được khởi động" }),
      });
    });

    await page.goto("/evaluation");

    const runBtn = page.getByRole("button", { name: /chạy đánh giá/i });
    await expect(runBtn).toBeVisible({ timeout: 8000 });
    await runBtn.click();

    // Should show success message
    await expect(
      page.getByText(/đánh giá đã được khởi động/i)
    ).toBeVisible({ timeout: 5000 });
  });

  test("handles API error gracefully", async ({ page }) => {
    await registerAndLogin(page);

    await page.route("**/api/v1/eval/stats**", async (route) => {
      await route.fulfill({ status: 500, body: "Internal Server Error" });
    });

    await page.goto("/evaluation");

    await expect(
      page.getByText(/không thể tải số liệu/i)
    ).toBeVisible({ timeout: 8000 });
  });
});
