import { execSync } from "child_process"
import path from "path"
import { fileURLToPath } from "url"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const projectRoot = path.resolve(__dirname, "../..")
const composeFiles = [
  "-f",
  "docker-compose.yml",
  "-f",
  "docker-compose.dev.yml",
]

const composeProject =
  process.env.E2E_COMPOSE_PROJECT?.trim() ||
  (process.env.CI ? "echostate-e2e" : "")

function composeArgs(): string[] {
  const args = ["docker", "compose"]
  if (composeProject) {
    args.push("-p", composeProject)
  }
  args.push(...composeFiles)
  return args
}

async function globalTeardown() {
  // Never pass -v here. Local dev uses the default compose project; CI uses
  // E2E_COMPOSE_PROJECT=echostate-e2e so volumes stay isolated from dev data.
  console.log("[global-teardown] Stopping Docker Compose stack (volumes preserved)...")
  try {
    execSync(composeArgs().concat(["down"]).join(" "), {
      cwd: projectRoot,
      stdio: "inherit",
      timeout: 120000,
    })
  } catch {
    // Best-effort cleanup.
  }
  console.log("[global-teardown] Stack stopped.")
}

export default globalTeardown
