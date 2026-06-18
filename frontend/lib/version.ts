import packageJson from "../package.json"

/** Application version — sourced from package.json, overridable at Docker/CI build time. */
export const APP_VERSION =
  process.env.NEXT_PUBLIC_APP_VERSION ?? packageJson.version