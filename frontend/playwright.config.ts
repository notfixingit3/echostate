import { defineConfig, devices } from "@playwright/test"
import path from "path"
import { fileURLToPath } from "url"

import { storageStateFile } from "./e2e/helpers/paths"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

export default defineConfig({
  testDir: "./e2e",
  timeout: 180_000,
  fullyParallel: false,
  expect: { timeout: 30_000 },
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  outputDir: "test-results",
  reporter: process.env.CI
    ? [["github"], ["list"]]
    : [["html", { outputFolder: "playwright-report", open: "never" }], ["list"]],
  use: {
    baseURL: "http://localhost:3001",
    trace: process.env.CI ? "retain-on-failure" : "on-first-retry",
    screenshot: "only-on-failure",
    video: process.env.CI ? "retain-on-failure" : "off",
    actionTimeout: 30_000,
    navigationTimeout: 60_000,
  },
  globalSetup: path.resolve(__dirname, "e2e/global-setup.ts"),
  globalTeardown: path.resolve(__dirname, "e2e/global-teardown.ts"),
  projects: [
    {
      name: "setup",
      testMatch: /auth\.setup\.ts/,
    },
    {
      name: "guest",
      testMatch: /auth\.guest\.spec\.ts/,
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "chromium",
      dependencies: ["setup"],
      testIgnore: [/auth\.setup\.ts/, /auth\.guest\.spec\.ts/],
      use: {
        ...devices["Desktop Chrome"],
        storageState: storageStateFile,
      },
    },
  ],
})