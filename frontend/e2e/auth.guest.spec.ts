import { test, expect } from "@playwright/test"

test("redirects unauthenticated users to login", async ({ page }) => {
  await page.goto("/")
  await expect(page).toHaveURL("/login")
  await expect(page.getByText("Sign in to EchoState")).toBeVisible()
})