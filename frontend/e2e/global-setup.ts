import { execSync, spawn } from "child_process"
import path from "path"
import { fileURLToPath } from "url"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const projectRoot = path.resolve(__dirname, "../..")

async function waitForHealth(maxSeconds = 180): Promise<void> {
  const deadline = Date.now() + maxSeconds * 1000

  while (Date.now() < deadline) {
    try {
      execSync(
        "docker compose exec -T api wget --spider http://localhost:8080/health",
        {
          cwd: projectRoot,
          stdio: "ignore",
          timeout: 10000,
        }
      )
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
      execSync(
        "docker compose exec -T frontend wget --spider http://127.0.0.1/",
        {
          cwd: projectRoot,
          stdio: "ignore",
          timeout: 10000,
        }
      )
      return
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 3000))
    }
  }

  throw new Error(`Frontend health check did not pass within ${maxSeconds}s`)
}

function isStackHealthy(): boolean {
  try {
    execSync(
      "docker compose exec -T api wget --spider http://localhost:8080/health",
      { cwd: projectRoot, stdio: "ignore", timeout: 10000 }
    )
    execSync(
      "docker compose exec -T frontend wget --spider http://127.0.0.1/",
      { cwd: projectRoot, stdio: "ignore", timeout: 10000 }
    )
    return true
  } catch {
    return false
  }
}

async function globalSetup() {
  if (isStackHealthy()) {
    console.log("[global-setup] Existing stack is healthy; skipping rebuild.")
    return
  }

  console.log("[global-setup] Bringing down any existing stack...")
  try {
    execSync("docker compose down -v", {
      cwd: projectRoot,
      stdio: "inherit",
      timeout: 120000,
    })
  } catch {
    // Ignore errors when there is no existing stack.
  }

  console.log("[global-setup] Building and starting Docker Compose stack...")
  const proc = spawn("docker", ["compose", "up", "--build", "-d"], {
    cwd: projectRoot,
    stdio: "inherit",
    shell: false,
  })

  await new Promise<void>((resolve, reject) => {
    proc.on("error", reject)
    proc.on("exit", (code) => {
      if (code === 0) {
        resolve()
      } else {
        reject(new Error(`docker compose up exited with code ${code}`))
      }
    })
  })

  console.log("[global-setup] Waiting for API health...")
  await waitForHealth()

  console.log("[global-setup] Waiting for frontend health...")
  await waitForFrontendHealth()

  console.log("[global-setup] Stack is ready.")
}

export default globalSetup
