import { test, expect } from "@playwright/test";
import path from "path";

// Generate unique email to avoid conflicts between test runs
function uniqueEmail() {
  return `test_${Date.now()}@example.com`;
}

test.describe("Entry review flow", () => {
  test("register → upload invoice → review entries → approve", async ({ page }) => {
    const email = uniqueEmail();
    const password = "Password123!";

    // 1. Register a new account
    await page.goto("/register");
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/mật khẩu/i).fill(password);
    await page.getByLabel(/họ tên/i).fill("Test User");
    await page.getByLabel(/tên công ty/i).fill("Công ty Test E2E");
    await page.getByRole("button", { name: /đăng ký/i }).click();

    // After register, should redirect to dashboard
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });

    // 2. Navigate to documents page and upload fixture
    await page.goto("/documents");
    await page.getByRole("button", { name: /upload chứng từ/i }).click();

    await expect(page.getByRole("dialog")).toBeVisible();

    // Select document type
    await page.getByLabel(/loại chứng từ/i).selectOption("invoice");

    // Upload the fixture file
    const fixturePath = path.join(__dirname, "fixtures", "sample-invoice.txt");
    await page.getByLabel(/chọn file/i).setInputFiles(fixturePath);

    await page.getByRole("button", { name: /tải lên/i }).click();

    // Wait for dialog to close
    await expect(page.getByRole("dialog")).not.toBeVisible({ timeout: 5000 });

    // Wait for document to appear in the list
    await expect(page.getByText("sample-invoice.txt")).toBeVisible({ timeout: 10000 });

    // 3. Poll until document is processed (status != uploaded/processing)
    await expect(async () => {
      await page.reload();
      // Look for booked or extracted status badge
      const statusCells = page.locator("td").filter({ hasText: /booked|extracted/ });
      await expect(statusCells.first()).toBeVisible();
    }).toPass({ timeout: 20000, intervals: [2000] });

    // 4. Navigate to entries page — pending tab is default
    await page.goto("/entries");
    await expect(page.getByRole("button", { name: "Chờ duyệt" })).toBeVisible();

    // Wait for at least one entry row in pending tab
    await expect(async () => {
      const rows = page.locator("tbody tr");
      await expect(rows.first()).toBeVisible();
    }).toPass({ timeout: 10000, intervals: [1000] });

    // 5. Approve the first pending entry
    const firstApproveBtn = page.getByRole("button", { name: "Duyệt" }).first();
    await expect(firstApproveBtn).toBeVisible();
    await firstApproveBtn.click();

    // After approval, the status badge for that entry should show "Đã duyệt"
    await expect(page.getByText("Đã duyệt").first()).toBeVisible({ timeout: 5000 });
  });

  test("shows success toast after approving an entry", async ({ page }) => {
    const email = uniqueEmail();
    const password = "Password123!";

    // Register and upload
    await page.goto("/register");
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/mật khẩu/i).fill(password);
    await page.getByLabel(/họ tên/i).fill("Toast Test User");
    await page.getByLabel(/tên công ty/i).fill("Công ty Toast E2E");
    await page.getByRole("button", { name: /đăng ký/i }).click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });

    await page.goto("/documents");
    await page.getByRole("button", { name: /upload chứng từ/i }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
    await page.getByLabel(/loại chứng từ/i).selectOption("invoice");
    const fixturePath = require("path").join(__dirname, "fixtures", "sample-invoice.txt");
    await page.getByLabel(/chọn file/i).setInputFiles(fixturePath);
    await page.getByRole("button", { name: /tải lên/i }).click();
    await expect(page.getByRole("dialog")).not.toBeVisible({ timeout: 5000 });

    // Wait for processing
    await expect(async () => {
      await page.reload();
      await expect(page.locator("td").filter({ hasText: /booked|extracted/ }).first()).toBeVisible();
    }).toPass({ timeout: 20000, intervals: [2000] });

    // Go to entries and approve
    await page.goto("/entries");
    await expect(async () => {
      await expect(page.locator("tbody tr").first()).toBeVisible();
    }).toPass({ timeout: 10000, intervals: [1000] });

    await page.getByRole("button", { name: "Duyệt" }).first().click();

    // Toast should appear
    await expect(page.getByText(/đã duyệt định khoản/i)).toBeVisible({ timeout: 5000 });
  });

  test("shows success toast after rejecting an entry", async ({ page }) => {
    const email = uniqueEmail();
    const password = "Password123!";

    await page.goto("/register");
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/mật khẩu/i).fill(password);
    await page.getByLabel(/họ tên/i).fill("Reject Toast User");
    await page.getByLabel(/tên công ty/i).fill("Công ty Reject Toast");
    await page.getByRole("button", { name: /đăng ký/i }).click();
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });

    await page.goto("/documents");
    await page.getByRole("button", { name: /upload chứng từ/i }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
    await page.getByLabel(/loại chứng từ/i).selectOption("invoice");
    const fixturePath = require("path").join(__dirname, "fixtures", "sample-invoice.txt");
    await page.getByLabel(/chọn file/i).setInputFiles(fixturePath);
    await page.getByRole("button", { name: /tải lên/i }).click();
    await expect(page.getByRole("dialog")).not.toBeVisible({ timeout: 5000 });

    await expect(async () => {
      await page.reload();
      await expect(page.locator("td").filter({ hasText: /booked|extracted/ }).first()).toBeVisible();
    }).toPass({ timeout: 20000, intervals: [2000] });

    await page.goto("/entries");
    await expect(async () => {
      await expect(page.locator("tbody tr").first()).toBeVisible();
    }).toPass({ timeout: 10000, intervals: [1000] });

    await page.getByRole("button", { name: "Từ chối" }).first().click();

    await expect(page.getByText(/đã từ chối định khoản/i)).toBeVisible({ timeout: 5000 });
  });
});
