import { expect, type CDPSession, type Page } from "@playwright/test"

export type BootstrapAuth = {
  enrollmentCode: string
  userId: string
  displayName: string
  codeType?: string
}

export async function enableWebAuthn(page: Page): Promise<CDPSession> {
  const client = await page.context().newCDPSession(page)
  await client.send("WebAuthn.enable")
  await client.send("WebAuthn.addVirtualAuthenticator", {
    options: {
      protocol: "ctap2",
      transport: "internal",
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
    },
  })
  return client
}

export async function loginWithEnrollmentCode(
  page: Page,
  bootstrap: BootstrapAuth
): Promise<void> {
  await page.goto("/login")
  await page.waitForLoadState("domcontentloaded")

  const codeInput = page.getByTestId("login-code-input")
  const useCodeButton = page.getByTestId("login-use-code-button")

  // Login defaults to step=passkey before hydration; wait for the stable step.
  await expect(codeInput.or(useCodeButton)).toBeVisible({ timeout: 15_000 })

  if (!(await codeInput.isVisible())) {
    await useCodeButton.click()
    await expect(codeInput).toBeVisible({ timeout: 15_000 })
  }

  await codeInput.fill(bootstrap.enrollmentCode)
  await page.getByTestId("login-verify-code-button").click()

  const registerOnDevice = page.getByTestId("login-register-passkey-button")
  const createPasskey = page.getByTestId("login-create-passkey-button")

  await expect(registerOnDevice.or(createPasskey)).toBeVisible({ timeout: 15_000 })

  if (await registerOnDevice.isVisible()) {
    await registerOnDevice.click()
  }

  await expect(createPasskey).toBeVisible({ timeout: 15_000 })
  await createPasskey.click()
  await page.waitForURL("/", { timeout: 30_000 })
}