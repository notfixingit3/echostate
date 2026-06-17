import { execSync } from "child_process"
import path from "path"
import { fileURLToPath } from "url"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const projectRoot = path.resolve(__dirname, "../..")

async function globalTeardown() {
  console.log("[global-teardown] Stopping Docker Compose stack...")
  try {
    execSync("docker compose down -v", {
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
