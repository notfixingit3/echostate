import { defineConfig, devices } from "@playwright/test"
import path from "path"
import { fileURLToPath } from "url"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const projectRoot = path.resolve(__dirname, "..")

export default defineConfig({
  testDir: "./e2e",
  timeout: 120_000,
  fullyParallel: false,
  expect: { timeout: 30000 },
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  outputDir: path.resolve(projectRoot, ".omo/evidence/task-21-traces"),
  reporter: [
    ["html", { outputFolder: path.resolve(projectRoot, ".omo/evidence/task-21-playwright-report") }],
    ["list"],
  ],
  use: {
    baseURL: "http://localhost:3001",
    trace: "on",
    screenshot: "on",
    video: "on",
    actionTimeout: 30000,
    navigationTimeout: 60000,
  },
  globalSetup: path.resolve(__dirname, "e2e/global-setup.ts"),
  globalTeardown: path.resolve(__dirname, "e2e/global-teardown.ts"),
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
})
