import { execSync } from "child_process"
import fs from "fs"
import path from "path"
import { fileURLToPath } from "url"

import { authDir, bootstrapAuthFile } from "./helpers/paths"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const projectRoot = path.resolve(__dirname, "../..")
const composeFiles = [
  "-f",
  "docker-compose.yml",
  "-f",
  "docker-compose.dev.yml",
]
const composeCmd = ["docker", "compose", ...composeFiles].join(" ")
const reuseStack = process.env.E2E_REUSE_STACK === "1"

async function waitForHealth(maxSeconds = 180): Promise<void> {
  const deadline = Date.now() + maxSeconds * 1000

  while (Date.now() < deadline) {
    try {
      execSync(`${composeCmd} exec -T api wget --spider http://localhost:8080/health`, {
        cwd: projectRoot,
        stdio: "ignore",
        timeout: 10000,
      })
      return
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 3000))
    }
  }

  throw new Error(`API health check did not pass within ${maxSeconds}s`)
}

async function waitForFrontendHealth(maxSeconds = 180): Promise<void> {
  const deadline = Date.now() + maxSeconds * 1000

  while (Date.now() < deadline) {
    try {
      execSync(`${composeCmd} exec -T frontend wget --spider http://127.0.0.1/`, {
        cwd: projectRoot,
        stdio: "ignore",
        timeout: 10000,
      })
      return
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 3000))
    }
  }

  throw new Error(`Frontend health check did not pass within ${maxSeconds}s`)
}

function isStackHealthy(): boolean {
  try {
    execSync(`${composeCmd} exec -T api wget --spider http://localhost:8080/health`, {
      cwd: projectRoot,
      stdio: "ignore",
      timeout: 10000,
    })
    execSync(`${composeCmd} exec -T frontend wget --spider http://127.0.0.1/`, {
      cwd: projectRoot,
      stdio: "ignore",
      timeout: 10000,
    })
    return true
  } catch {
    return false
  }
}

async function isAuthRequired(): Promise<boolean> {
  try {
    const response = await fetch("http://localhost:8080/api/auth/config")
    if (!response.ok) {
      return false
    }
    const data = (await response.json()) as { auth_required?: boolean }
    return data.auth_required === true
  } catch {
    return false
  }
}

function writeBootstrapMetadata(payload: Record<string, string>): void {
  fs.mkdirSync(authDir, { recursive: true })
  fs.writeFileSync(bootstrapAuthFile, JSON.stringify(payload, null, 2) + "\n")
}

function bootstrapAdmin(): void {
  console.log("[global-setup] Bootstrapping first admin user...")
  const output = execSync(
    `${composeCmd} run --rm api auth bootstrap-admin --name "E2E Admin"`,
    {
      cwd: projectRoot,
      encoding: "utf8",
      timeout: 180000,
      stdio: ["ignore", "pipe", "pipe"],
    }
  )

  const codeMatch = output.match(/Enrollment code \(single-use, expires [^)]+\):\s*(\S+)/)
  const userMatch = output.match(/Admin user created:\s*(?:.+?)\s*\(([0-9a-f-]{36})\)/i)

  if (!codeMatch?.[1] || !userMatch?.[1]) {
    throw new Error(`Failed to parse bootstrap-admin output:\n${output}`)
  }

  writeBootstrapMetadata({
    enrollmentCode: codeMatch[1],
    userId: userMatch[1],
    displayName: "E2E Admin",
    codeType: "enrollment",
  })
}

function issueRecoveryCode(): void {
  console.log("[global-setup] Issuing admin recovery code for Playwright setup...")
  const output = execSync(`${composeCmd} run --rm api auth issue-admin-code`, {
    cwd: projectRoot,
    encoding: "utf8",
    timeout: 180000,
    stdio: ["ignore", "pipe", "pipe"],
  })

  const codeMatch = output.match(/Recovery code for .+ \(expires [^)]+\):\s*(\S+)/)
  if (!codeMatch?.[1]) {
    throw new Error(
      `Failed to parse recovery code output. Set ECHOSTATE_BREAK_GLASS_SECRET in .env for E2E reuse:\n${output}`
    )
  }

  writeBootstrapMetadata({
    enrollmentCode: codeMatch[1],
    userId: "",
    displayName: "E2E Admin",
    codeType: "recovery",
  })
}

async function ensureBootstrapMetadata(): Promise<void> {
  if (await isAuthRequired()) {
    issueRecoveryCode()
    return
  }
  bootstrapAdmin()
}

function startStack(): void {
  console.log("[global-setup] Building and starting Docker Compose stack...")
  execSync(`${composeCmd} up --build -d`, {
    cwd: projectRoot,
    stdio: "inherit",
    timeout: 900000,
  })
}

function stopStack(removeVolumes: boolean): void {
  const args = removeVolumes ? "down -v" : "down"
  try {
    execSync(`${composeCmd} ${args}`, {
      cwd: projectRoot,
      stdio: "inherit",
      timeout: 180000,
    })
  } catch {
    // Ignore when no stack exists.
  }
}

async function globalSetup() {
  if (reuseStack && isStackHealthy()) {
    console.log("[global-setup] E2E_REUSE_STACK=1 — reusing healthy stack.")
    await ensureBootstrapMetadata()
    return
  }

  console.log("[global-setup] Recreating Docker Compose stack with fresh volumes...")
  stopStack(true)
  startStack()

  console.log("[global-setup] Waiting for API health...")
  await waitForHealth()
  console.log("[global-setup] Waiting for frontend health...")
  await waitForFrontendHealth()

  bootstrapAdmin()
  console.log("[global-setup] Stack is ready.")
}

export default globalSetup