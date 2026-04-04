import { test, expect } from "@playwright/test";

test.describe("Home page", () => {
  test("shows landing page with CTA buttons", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByRole("heading", { name: /AI Kế Toán/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /Dùng thử miễn phí/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /Đăng nhập/i })).toBeVisible();
  });

  test("navigates to login page", async ({ page }) => {
    await page.goto("/");
    await page.getByRole("link", { name: /Đăng nhập/i }).click();
    await expect(page).toHaveURL("/login");
  });

  test("navigates to register page", async ({ page }) => {
    await page.goto("/");
    await page.getByRole("link", { name: /Dùng thử miễn phí/i }).click();
    await expect(page).toHaveURL("/register");
  });
});
