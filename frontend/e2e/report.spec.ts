import { test, expect } from "@playwright/test"
import fs from "fs"

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

    await expect(page.getByTestId("snapshot-report-pending-button")).toBeVisible()

    const statusAlert = page.getByTestId("snapshot-report-status-alert")
    await expect(statusAlert).toBeVisible()
    await expect(statusAlert).toContainText(/is (pending|running)/i)

    const downloadButton = page.getByTestId("snapshot-download-report-button")
    await expect(downloadButton).toBeVisible({ timeout: 60_000 })
    await expect(page.getByTestId("snapshot-view-report-link")).toBeVisible()
    await expect(page.getByTestId("snapshot-report-pending-button")).toHaveCount(0)

    const [download] = await Promise.all([
      page.waitForEvent("download"),
      downloadButton.click(),
    ])

    const path = await download.path()
    expect(path).toBeTruthy()

    const buffer = fs.readFileSync(path!)
    expect(buffer.toString("ascii", 0, 4)).toBe("%PDF")
  })
})