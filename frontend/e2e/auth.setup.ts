import { test as setup } from "@playwright/test"
import fs from "fs"

import { enableWebAuthn, loginWithEnrollmentCode, type BootstrapAuth } from "./helpers/auth"
import { authDir, bootstrapAuthFile, storageStateFile } from "./helpers/paths"

setup("authenticate admin with passkey", async ({ page }) => {
  if (!fs.existsSync(bootstrapAuthFile)) {
    throw new Error(
      `Missing ${bootstrapAuthFile}. Global setup should bootstrap admin before Playwright runs.`
    )
  }

  const bootstrap = JSON.parse(
    fs.readFileSync(bootstrapAuthFile, "utf8")
  ) as BootstrapAuth

  fs.mkdirSync(authDir, { recursive: true })
  if (fs.existsSync(storageStateFile)) {
    fs.unlinkSync(storageStateFile)
  }

  await enableWebAuthn(page)
  await loginWithEnrollmentCode(page, bootstrap)
  await page.getByTestId("nav-account-menu").waitFor({ state: "visible", timeout: 30_000 })
  await page.context().storageState({ path: storageStateFile })
})