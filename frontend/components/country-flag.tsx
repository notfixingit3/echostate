import { cn } from "@/lib/utils"
import { countryCodeToFlag, normalizeCountryCode } from "@/lib/country"

export function CountryFlag({
  code,
  showCode = true,
  className,
  fallback = "—",
}: {
  code?: string | null
  showCode?: boolean
  className?: string
  fallback?: string
}) {
  const normalized = normalizeCountryCode(code)
  const flag = countryCodeToFlag(normalized)

  if (!normalized) {
    return <span className={cn("text-muted-foreground", className)}>{fallback}</span>
  }

  return (
    <span
      className={cn("inline-flex items-center gap-1.5", className)}
      title={normalized}
    >
      {flag ? (
        <span className="text-base leading-none" role="img" aria-label={normalized}>
          {flag}
        </span>
      ) : null}
      {showCode ? (
        <span className="font-mono text-xs tracking-wide">{normalized}</span>
      ) : null}
    </span>
  )
}