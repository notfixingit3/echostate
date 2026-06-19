/** Canonical public site origin for sitemap/robots (no trailing slash). */
export function siteBaseUrl(): string {
  const configured = process.env.NEXT_PUBLIC_SITE_URL?.trim()
  if (configured) {
    return configured.replace(/\/$/, "")
  }

  const vercel = process.env.VERCEL_URL?.trim()
  if (vercel) {
    return `https://${vercel.replace(/\/$/, "")}`
  }

  return "http://localhost:3000"
}