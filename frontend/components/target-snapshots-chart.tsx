"use client"

import * as React from "react"
import { format } from "date-fns"
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { fetchApi } from "@/lib/api"
import type { PaginatedResponse, SnapshotSummary } from "@/lib/types"

export function TargetSnapshotsChart({ targetId }: { targetId: string }) {
  const [data, setData] = React.useState<{ date: string; changes: number; rawDate: Date }[]>([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const result = await fetchApi<PaginatedResponse<SnapshotSummary>>(
          `/api/targets/${targetId}/snapshots?page=1&limit=100`
        )
        if (!cancelled) {
          const chartData = [...result.data].reverse().map(s => ({
            date: format(new Date(s.scanned_at), "MMM d, HH:mm"),
            rawDate: new Date(s.scanned_at),
            changes: s.changes?.length || 0,
          }))
          setData(chartData)
        }
      } catch (err) {
        // Ignore error for chart
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    load()

    return () => {
      cancelled = true
    }
  }, [targetId])

  if (loading || data.length < 2) {
    return null
  }

  return (
    <Card className="overflow-hidden border-border/60 bg-card/60 backdrop-blur-sm">
      <CardHeader className="pb-4">
        <CardTitle className="text-base">Snapshot Activity</CardTitle>
        <CardDescription>Number of changes detected over time</CardDescription>
      </CardHeader>
      <CardContent className="h-64 px-4 sm:px-6">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="colorChanges" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.3} />
                <stop offset="95%" stopColor="var(--primary)" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border)" opacity={0.5} />
            <XAxis
              dataKey="date"
              axisLine={false}
              tickLine={false}
              tick={{ fontSize: 12, fill: "var(--muted-foreground)" }}
              dy={10}
              minTickGap={30}
            />
            <YAxis
              axisLine={false}
              tickLine={false}
              tick={{ fontSize: 12, fill: "var(--muted-foreground)" }}
              width={36}
              allowDecimals={false}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "var(--card)",
                borderColor: "var(--border)",
                borderRadius: "var(--radius)",
                fontSize: "12px",
              }}
              itemStyle={{ color: "var(--foreground)" }}
            />
            <Area
              type="monotone"
              dataKey="changes"
              name="Changes"
              stroke="var(--primary)"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#colorChanges)"
              activeDot={{ r: 4, strokeWidth: 0, fill: "var(--primary)" }}
            />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}
