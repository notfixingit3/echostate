import { expect, type Page } from "@playwright/test"

export async function acceptConfirmDialog(page: Page) {
  const confirmButton = page.getByTestId("confirm-dialog-confirm")
  await expect(confirmButton).toBeVisible({ timeout: 5_000 })
  await confirmButton.click()
  await expect(page.getByTestId("confirm-dialog")).toHaveCount(0)
}