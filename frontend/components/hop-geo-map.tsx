"use client"

import type { GraphGeoPoint } from "@/lib/types"

type HopGeoMapProps = {
  points: GraphGeoPoint[]
  className?: string
}

const MAP_WIDTH = 720
const MAP_HEIGHT = 360

function project(lat: number, lon: number) {
  const x = ((lon + 180) / 360) * MAP_WIDTH
  const y = ((90 - lat) / 180) * MAP_HEIGHT
  return { x, y }
}

export function HopGeoMap({ points, className }: HopGeoMapProps) {
  if (points.length === 0) {
    return (
      <div
        className={
          className ??
          "flex h-48 items-center justify-center rounded-lg border border-border/50 bg-muted/10 text-sm text-muted-foreground"
        }
      >
        No geolocated hops yet. Re-scan targets to enrich traceroute hops.
      </div>
    )
  }

  return (
    <div
      className={className ?? "overflow-hidden rounded-lg border border-border/50 bg-slate-950/80"}
      data-testid="hop-geo-map"
    >
      <svg
        viewBox={`0 0 ${MAP_WIDTH} ${MAP_HEIGHT}`}
        className="h-auto w-full"
        role="img"
        aria-label="Traceroute hop world map"
      >
        <rect width={MAP_WIDTH} height={MAP_HEIGHT} fill="#0f172a" />
        {Array.from({ length: 13 }).map((_, index) => {
          const y = (index / 12) * MAP_HEIGHT
          return (
            <line
              key={`lat-${index}`}
              x1={0}
              y1={y}
              x2={MAP_WIDTH}
              y2={y}
              stroke="rgba(148, 163, 184, 0.12)"
              strokeWidth={1}
            />
          )
        })}
        {Array.from({ length: 25 }).map((_, index) => {
          const x = (index / 24) * MAP_WIDTH
          return (
            <line
              key={`lon-${index}`}
              x1={x}
              y1={0}
              x2={x}
              y2={MAP_HEIGHT}
              stroke="rgba(148, 163, 184, 0.12)"
              strokeWidth={1}
            />
          )
        })}

        {points.map((point, index) => {
          const { x, y } = project(point.lat, point.lon)
          const color =
            index === points.length - 1
              ? "#f59e0b"
              : index === 0
                ? "#22d3ee"
                : "#38bdf8"
          return (
            <g key={`${point.target_id ?? "all"}-${point.hop}-${point.ip}`}>
              <circle cx={x} cy={y} r={5} fill={color} opacity={0.9} />
              <circle cx={x} cy={y} r={9} fill={color} opacity={0.18} />
              <title>
                {point.target_label ? `${point.target_label} — ` : ""}
                hop {point.hop}: {point.ip} ({point.label})
              </title>
            </g>
          )
        })}
      </svg>
      <div className="flex flex-wrap gap-3 border-t border-border/40 px-3 py-2 text-[11px] text-muted-foreground">
        <span className="inline-flex items-center gap-1.5">
          <span className="size-2 rounded-full bg-cyan-400" />
          First geolocated hop
        </span>
        <span className="inline-flex items-center gap-1.5">
          <span className="size-2 rounded-full bg-sky-400" />
          Transit hops
        </span>
        <span className="inline-flex items-center gap-1.5">
          <span className="size-2 rounded-full bg-amber-400" />
          Last plotted hop
        </span>
      </div>
    </div>
  )
}