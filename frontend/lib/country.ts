/** Convert an ISO 3166-1 alpha-2 code to a regional-indicator emoji flag. */
export function countryCodeToFlag(code?: string | null): string | null {
  if (!code) return null

  const normalized = code.trim().toUpperCase()
  if (!/^[A-Z]{2}$/.test(normalized)) return null

  const codePoints = [...normalized].map(
    (char) => 0x1f1e6 + char.charCodeAt(0) - 65
  )

  return String.fromCodePoint(...codePoints)
}

export function normalizeCountryCode(code?: string | null): string | null {
  if (!code) return null
  const normalized = code.trim().toUpperCase()
  return /^[A-Z]{2}$/.test(normalized) ? normalized : null
}