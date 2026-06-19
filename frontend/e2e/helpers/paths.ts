import path from "path"
import { fileURLToPath } from "url"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

export const e2eRoot = path.resolve(__dirname, "..")
export const authDir = path.join(e2eRoot, ".auth")
export const bootstrapAuthFile = path.join(authDir, "bootstrap.json")
export const storageStateFile = path.join(authDir, "user.json")