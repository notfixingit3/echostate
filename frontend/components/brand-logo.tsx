import { cn } from "@/lib/utils"

type BrandLogoProps = {
  variant?: "banner" | "icon"
  className?: string
  title?: string
}

function RadarMark({ className }: { className?: string }) {
  return (
    <g className={cn("text-primary", className)}>
      <circle cx="24" cy="24" r="3.25" className="fill-primary" />
      <circle
        cx="24"
        cy="24"
        r="7.5"
        className="fill-none stroke-primary"
        strokeWidth="2"
        opacity="0.95"
      />
      <circle
        cx="24"
        cy="24"
        r="12.5"
        className="fill-none stroke-primary"
        strokeWidth="1.75"
        opacity="0.65"
      />
      <circle
        cx="24"
        cy="24"
        r="17.5"
        className="fill-none stroke-primary"
        strokeWidth="1.25"
        opacity="0.4"
      />
      <path
        d="M24 6.5 A17.5 17.5 0 0 1 41.5 24"
        className="fill-none stroke-primary"
        strokeWidth="2.25"
        strokeLinecap="round"
        opacity="0.9"
      />
      <circle cx="39" cy="21" r="1.75" className="fill-primary" opacity="0.85" />
    </g>
  )
}

export function BrandLogo({
  variant = "banner",
  className,
  title = "EchoState",
}: BrandLogoProps) {
  if (variant === "icon") {
    return (
      <svg
        viewBox="0 0 48 48"
        role="img"
        aria-label={title}
        className={cn("h-auto w-auto shrink-0", className)}
      >
        <title>{title}</title>
        <RadarMark />
      </svg>
    )
  }

  return (
    <svg
      viewBox="0 0 280 64"
      role="img"
      aria-label={title}
      className={cn("h-auto w-auto", className)}
    >
      <title>{title}</title>
      <g transform="translate(8 8)">
        <RadarMark />
      </g>
      <text
        x="72"
        y="41"
        className="fill-foreground font-heading"
        fontSize="28"
        fontWeight="600"
        letterSpacing="-0.02em"
      >
        EchoState
      </text>
    </svg>
  )
}