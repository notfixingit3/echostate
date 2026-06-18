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

async function globalTeardown() {
  // Stop containers only — do not pass -v. Wiping postgres_data destroys the
  // user's targets, snapshots, and auth users on the same compose project.
  console.log("[global-teardown] Stopping Docker Compose stack (volumes preserved)...")
  try {
    execSync(["docker", "compose", ...composeFiles, "down"].join(" "), {
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
