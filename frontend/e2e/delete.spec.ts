import { test, expect } from "@playwright/test"

test.describe("List delete", () => {
  test.beforeEach(async ({ page }) => {
    page.on("dialog", (dialog) => dialog.accept())
  })

  test("admin deletes a snapshot from the list", async ({ page }) => {
    await page.goto("/")
    await page.getByTestId("scan-host-input").fill("delete-single.example.com")
    await page.getByTestId("scan-submit-button").click()
    await expect(page.getByTestId("scan-result-card")).toBeVisible({ timeout: 120_000 })

    await page.goto("/snapshots")
    const row = page.locator('[data-testid^="snapshot-row-"]').filter({
      hasText: "delete-single.example.com",
    })
    await expect(row).toBeVisible({ timeout: 30_000 })

    const testId = await row.getAttribute("data-testid")
    const snapshotId = testId?.replace("snapshot-row-", "")
    expect(snapshotId).toBeTruthy()

    await row.getByRole("button", { name: /delete snapshot/i }).click()
    await expect(page.locator(`[data-testid="snapshot-row-${snapshotId}"]`)).toHaveCount(0)
  })

  test("admin bulk-deletes selected snapshots", async ({ page }) => {
    for (const host of ["bulk-a.example.com", "bulk-b.example.com"]) {
      await page.goto("/")
      await page.getByTestId("scan-host-input").fill(host)
      await page.getByTestId("scan-submit-button").click()
      await expect(page.getByTestId("scan-result-card")).toBeVisible({ timeout: 120_000 })
    }

    await page.goto("/snapshots")
    const rowA = page.locator('[data-testid^="snapshot-row-"]').filter({
      hasText: "bulk-a.example.com",
    })
    const rowB = page.locator('[data-testid^="snapshot-row-"]').filter({
      hasText: "bulk-b.example.com",
    })
    await expect(rowA).toBeVisible({ timeout: 30_000 })
    await expect(rowB).toBeVisible({ timeout: 30_000 })

    const idA = (await rowA.getAttribute("data-testid"))?.replace("snapshot-row-", "")
    const idB = (await rowB.getAttribute("data-testid"))?.replace("snapshot-row-", "")

    await rowA.getByTestId(`snapshot-select-${idA}`).click()
    await rowB.getByTestId(`snapshot-select-${idB}`).click()

    await expect(page.getByTestId("bulk-action-bar")).toBeVisible()
    await page.getByTestId("bulk-delete-button").click()

    await expect(page.locator(`[data-testid="snapshot-row-${idA}"]`)).toHaveCount(0)
    await expect(page.locator(`[data-testid="snapshot-row-${idB}"]`)).toHaveCount(0)
  })

  test("scanner deletes a report from the list", async ({ page }) => {
    await page.goto("/")
    await page.getByTestId("scan-host-input").fill("report-delete.example.com")
    await page.getByTestId("scan-submit-button").click()
    await expect(page.getByTestId("scan-result-card")).toBeVisible({ timeout: 120_000 })

    const snapshotId = (await page.getByTestId("scan-result-snapshot-id").textContent())?.trim()
    expect(snapshotId).toBeTruthy()

    await page.goto(`/snapshot?id=${snapshotId}`)
    await page.getByTestId("snapshot-create-report-button").click()
    await expect(page.getByTestId("snapshot-report-status-alert")).toBeVisible({
      timeout: 30_000,
    })

    await page.goto(`/reports?snapshot_id=${snapshotId}`)
    const row = page.locator('[data-testid^="report-row-"]').first()
    await expect(row).toBeVisible({ timeout: 30_000 })

    const reportId = (await row.getAttribute("data-testid"))?.replace("report-row-", "")
    expect(reportId).toBeTruthy()

    await row.getByRole("button", { name: /delete report/i }).click()
    await expect(page.locator(`[data-testid="report-row-${reportId}"]`)).toHaveCount(0)
  })
})