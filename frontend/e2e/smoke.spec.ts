import { test, expect } from "@playwright/test"

const UUID_REGEX =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

test.describe("Smoke", () => {
  test("home page renders logo and navigation", async ({ page }) => {
    await page.goto("/")

    await expect(page.getByTestId("nav-logo")).toBeVisible()
    await expect(page.getByTestId("theme-toggle")).toBeVisible()

    const manifestHref = await page
      .locator('link[rel="manifest"]')
      .getAttribute("href")
    expect(manifestHref).toMatch(/manifest/)

    for (const label of ["Home", "Targets", "Snapshots", "Graph", "Reports"]) {
      await expect(
        page.getByTestId(`nav-link-${label.toLowerCase()}`)
      ).toBeVisible()
    }

    await expect(page.getByTestId("footer")).toBeVisible()
    await expect(page.getByTestId("footer-version")).toHaveText(
      /v0\.0\.\d+(-(beta|dev)\.\d+)?/
    )
    await expect(page.getByTestId("scan-submit-button")).toBeVisible()
    await expect(page.getByTestId("scan-and-report-button")).toBeVisible()
  })

  test("scan submission shows result card with snapshot link", async ({
    page,
  }) => {
    await page.goto("/")

    const hostInput = page.getByTestId("scan-host-input")
    await hostInput.fill("example.com")

    await page.getByTestId("scan-submit-button").click()

    const resultCard = page.getByTestId("scan-result-card")
    await expect(resultCard).toBeVisible({ timeout: 120_000 })

    const snapshotId = page.getByTestId("scan-result-snapshot-id")
    await expect(snapshotId).toHaveText(UUID_REGEX)

    const viewLink = page.getByTestId("scan-result-view-link")
    const href = await viewLink.getAttribute("href")
    const id = await snapshotId.textContent()
    expect(href).toBe(`/snapshot?id=${id}`)
  })

  test("clicking a target opens the target detail page", async ({ page }) => {
    await page.goto("/targets")

    const firstRow = page.locator('[data-testid^="target-row-"]').first()
    await expect(firstRow).toBeVisible()

    const testId = await firstRow.getAttribute("data-testid")
    const targetId = testId?.replace("target-row-", "")
    expect(targetId).toBeTruthy()

    await firstRow.click()
    await expect(page).toHaveURL(`/target?id=${targetId}`)
    await expect(
      page.getByRole("heading", { name: "Target detail" })
    ).toBeVisible()
    await expect(page.getByTestId("target-rescan-button")).toBeVisible()
  })

  test("targets page renders table", async ({ page }) => {
    await page.goto("/targets")

    const table = page.getByTestId("targets-table-container")
    await expect(table).toBeVisible()

    await expect(table.locator("table tbody tr")).not.toHaveCount(0)
  })

  test("snapshots page renders table with Client IP column", async ({
    page,
  }) => {
    await page.goto("/snapshots")

    const table = page.getByTestId("snapshots-table-container")
    await expect(table).toBeVisible()

    await expect(table.getByRole("columnheader", { name: "Client IP" })).toBeVisible()
    await expect(table.locator("table tbody tr")).not.toHaveCount(0)
  })

  test("reports page renders table with status badges", async ({ page }) => {
    await page.goto("/reports")

    const table = page.getByTestId("reports-table-container")
    await expect(table).toBeVisible()

    const statusHeader = table.getByRole("columnheader", { name: "Status" })
    await expect(statusHeader).toBeVisible()

    await expect(table.locator("table tbody tr")).not.toHaveCount(0)
  })
})
