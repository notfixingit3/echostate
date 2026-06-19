import { test, expect } from "@playwright/test"

test.describe("Authenticated admin", () => {
  test("shows Server settings link and account menu", async ({ page }) => {
    await page.goto("/")
    await expect(page.getByTestId("nav-link-server-settings")).toBeVisible()
    await expect(page.getByTestId("nav-link-admin")).not.toBeVisible()
    await expect(page.getByTestId("nav-account-menu")).toBeVisible()
    await page.getByTestId("nav-account-menu").click()
    await expect(page.getByTestId("nav-account-profile")).toBeVisible()
    await expect(page.getByTestId("nav-account-sign-out")).toBeVisible()
  })
})