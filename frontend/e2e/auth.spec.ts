import { test, expect } from "@playwright/test"

test.describe("Authenticated admin", () => {
  test("shows Profile and Settings in navigation", async ({ page }) => {
    await page.goto("/")
    await expect(page.getByTestId("nav-link-profile")).toBeVisible()
    await expect(page.getByTestId("nav-link-settings")).toBeVisible()
    await expect(page.getByTestId("nav-link-admin")).toBeVisible()
  })
})