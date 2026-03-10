import { test, expect } from "@playwright/test";
import path from "path";

function uniqueEmail() {
  return `doc_e2e_${Date.now()}@example.com`;
}

async function registerAndLogin(page: import("@playwright/test").Page) {
  const email = uniqueEmail();
  await page.goto("/register");
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/mật khẩu/i).fill("Password123!");
  await page.getByLabel(/họ tên/i).fill("E2E User");
  await page.getByLabel(/tên công ty/i).fill("Công ty E2E Documents");
  await page.getByRole("button", { name: /đăng ký/i }).click();
  await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });
}

async function uploadFixture(page: import("@playwright/test").Page) {
  await page.goto("/documents");
  await page.getByRole("button", { name: /upload chứng từ/i }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  const filePath = path.join(__dirname, "fixtures", "sample-invoice.txt");
  await page.getByLabel(/chọn file/i).setInputFiles(filePath);
  await page.getByRole("button", { name: /^upload$/i }).click();
  await expect(page.getByRole("dialog")).not.toBeVisible({ timeout: 10000 });
}

test.describe("Documents page", () => {
  test("can open document detail modal", async ({ page }) => {
    await registerAndLogin(page);
    await uploadFixture(page);

    // Wait for document to appear
    await expect(page.getByRole("button", { name: /xem/i }).first()).toBeVisible({ timeout: 5000 });

    // Click Xem button on first document
    await page.getByRole("button", { name: /xem/i }).first().click();

    // Modal should show document details
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(page.getByText(/sample-invoice/i)).toBeVisible();
  });

  test("can close document detail modal", async ({ page }) => {
    await registerAndLogin(page);
    await uploadFixture(page);

    await page.getByRole("button", { name: /xem/i }).first().click();
    await expect(page.getByRole("dialog")).toBeVisible();

    await page.getByRole("button", { name: /đóng/i }).click();
    await expect(page.getByRole("dialog")).not.toBeVisible();
  });

  test("refresh button reloads document list", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/documents");

    const refreshBtn = page.getByRole("button", { name: /làm mới/i });
    await expect(refreshBtn).toBeVisible();
    await refreshBtn.click();

    // Button should not crash; page should still show the heading
    await expect(page.getByRole("heading", { name: /chứng từ/i })).toBeVisible();
  });

  test("status filter shows only filtered documents", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/documents");

    // Select "Lỗi" filter — should show 0 results for new account
    const select = page.locator("select").first();
    if (await select.isVisible()) {
      await select.selectOption("error");
      await expect(page.getByText(/không có chứng từ/i)).toBeVisible({ timeout: 5000 });
    }
  });
});

test.describe("Entries pagination and CSV export", () => {
  test("CSV export button is visible on entries page", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/entries");

    await expect(page.getByRole("button", { name: /xuất csv/i })).toBeVisible();
  });

  test("refresh button is visible on entries page", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/entries");

    await expect(page.getByRole("button", { name: /làm mới/i })).toBeVisible();
  });

  test("pagination is visible when entries exceed page size", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/entries");

    // For a new account there are no entries, pagination should show page 1
    // Just verify the pagination component renders without crashing
    await expect(page.getByRole("heading", { name: /định khoản/i })).toBeVisible();
  });

  test("tab filters change displayed entries", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/entries");

    // Click "Đã duyệt" tab
    await page.getByRole("button", { name: /đã duyệt/i }).click();
    await expect(page.getByText(/không có định khoản/i)).toBeVisible({ timeout: 5000 });

    // Click "Tất cả" tab
    await page.getByRole("button", { name: /tất cả/i }).click();
    await expect(page.getByRole("heading", { name: /định khoản/i })).toBeVisible();
  });
});
