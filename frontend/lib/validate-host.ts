const HOSTNAME_LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/

function normalizeHost(host: string): string {
  const trimmed = host.trim()
  if (!trimmed) return ""

  let input = trimmed
  if (!input.includes("://")) {
    input = `https://${input}`
  }

  try {
    const url = new URL(input)
    let hostname = url.hostname || url.pathname
    hostname = hostname.replace(/^www\./i, "")
    return hostname.toLowerCase()
  } catch {
    return trimmed
      .replace(/^https?:\/\//i, "")
      .replace(/^www\./i, "")
      .toLowerCase()
  }
}

function isIPv4(value: string): boolean {
  const parts = value.split(".")
  if (parts.length !== 4) return false
  return parts.every((part) => {
    if (!/^\d{1,3}$/.test(part)) return false
    const n = Number(part)
    return n >= 0 && n <= 255
  })
}

function isIPv6(value: string): boolean {
  return value.includes(":")
}

export function validateScanTarget(host: string): string {
  const normalized = normalizeHost(host)
  if (!normalized) {
    throw new Error("Please enter a host, IP address, or URL.")
  }

  if (isIPv4(normalized) || isIPv6(normalized)) {
    return normalized
  }

  if (!normalized.includes(".")) {
    throw new Error(
      `Invalid host "${normalized}": hostnames must include a domain suffix (e.g. example.com).`
    )
  }

  if (normalized.length > 253) {
    throw new Error("Host is too long.")
  }

  const labels = normalized.split(".")
  if (labels.length < 2) {
    throw new Error(`Invalid host "${normalized}".`)
  }

  for (const label of labels) {
    if (!label || label.length > 63 || !HOSTNAME_LABEL.test(label)) {
      throw new Error(`Invalid host "${normalized}".`)
    }
  }

  const tld = labels[labels.length - 1]
  if (tld.length < 2) {
    throw new Error(
      `Invalid host "${normalized}": domain suffix is too short.`
    )
  }

  return normalized
}