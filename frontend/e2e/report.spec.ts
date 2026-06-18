import { test, expect } from "@playwright/test"

const UUID_REGEX =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

test.describe("Report flow", () => {
  test("creates and downloads a PDF report from a snapshot", async ({ page }) => {
    await page.goto("/")

    await page.getByTestId("scan-host-input").fill("example.com")
    await page.getByTestId("scan-submit-button").click()

    const resultCard = page.getByTestId("scan-result-card")
    await expect(resultCard).toBeVisible({ timeout: 120_000 })

    const snapshotId = await page
      .getByTestId("scan-result-snapshot-id")
      .textContent()
    expect(snapshotId).toMatch(UUID_REGEX)

    await page.getByTestId("scan-result-view-link").click()
    await expect(page).toHaveURL(`/snapshot?id=${snapshotId}`)

    await page.getByTestId("snapshot-create-report-button").click()

    const statusAlert = page.getByTestId("snapshot-report-status-alert")
    await expect(statusAlert).toBeVisible()
    await expect(statusAlert).toContainText(/is pending/i)

    const alertText = await statusAlert.textContent()
    const reportIdMatch = alertText?.match(
      /Report\s+([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})/i
    )
    expect(reportIdMatch).toBeTruthy()
    const reportId = reportIdMatch![1]

    await page.goto(`/report?id=${reportId}`)

    const statusBadge = page.getByTestId("report-detail-status")
    await expect(statusBadge).toBeVisible()
    await expect(statusBadge).toHaveText("completed", { timeout: 60_000 })

    const [download] = await Promise.all([
      page.waitForEvent("download"),
      page.getByTestId("report-detail-download-button").click(),
    ])

    const path = await download.path()
    expect(path).toBeTruthy()

    const fs = await import("fs")
    const buffer = fs.readFileSync(path!)
    expect(buffer.toString("ascii", 0, 4)).toBe("%PDF")
  })
})
