import { test, expect } from "@playwright/test";

function uniqueEmail() {
  return `settings_test_${Date.now()}@example.com`;
}

async function registerAndLogin(page: import("@playwright/test").Page) {
  const email = uniqueEmail();
  await page.goto("/register");
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/mật khẩu/i).fill("Password123!");
  await page.getByLabel(/họ tên/i).fill("Test User");
  await page.getByLabel(/tên công ty/i).fill("Công ty Settings Test");
  await page.getByRole("button", { name: /đăng ký/i }).click();
  await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });
}

test.describe("Settings page", () => {
  test("can update organization name", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/settings");

    await expect(page.getByRole("heading", { name: /cài đặt/i })).toBeVisible();

    // Clear and fill org name
    const nameInput = page.getByLabel(/tên doanh nghiệp/i);
    await nameInput.clear();
    await nameInput.fill("Công ty Mới ABC");

    await page.getByRole("button", { name: /^lưu$/i }).click();

    await expect(page.getByText(/đã lưu thành công/i)).toBeVisible({ timeout: 5000 });
  });

  test("shows OCR provider select with fpt and mock options", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/settings");

    const ocrSelect = page.getByLabel(/ocr provider/i);
    await expect(ocrSelect).toBeVisible();

    await expect(ocrSelect.getByRole("option", { name: /FPT.AI/i })).toBeAttached();
    await expect(ocrSelect.getByRole("option", { name: /Mock/i })).toBeAttached();
  });

  test("can save connection settings", async ({ page }) => {
    await registerAndLogin(page);
    await page.goto("/settings");

    // Select FPT.AI provider
    await page.getByLabel(/ocr provider/i).selectOption("fpt");

    // Fill MISA URL
    await page.getByLabel(/misa api url/i).fill("https://api.misa.vn/test");

    await page.getByRole("button", { name: /lưu kết nối/i }).click();

    await expect(page.getByText(/đã lưu thành công/i)).toBeVisible({ timeout: 5000 });
  });
});
