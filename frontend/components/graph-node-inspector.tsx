"use client"

import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { CopyButton } from "@/components/copy-button"
import { formatFieldLabel } from "@/lib/intel"
import type { GraphEdge, GraphNode } from "@/lib/types"

const INSPECTOR_FIELDS: Record<string, string[]> = {
  Target: ["host", "id", "created_at"],
  Snapshot: ["id", "scanned_at", "data_hash"],
  ASN: ["number"],
  IP: ["address"],
  Prefix: ["cidr", "hijack_risk", "rpki_status"],
  Hop: ["hop", "ip", "rtt_ms", "country", "city", "vantage_label"],
  SharedHop: ["ip"],
  JARM: ["hash"],
  Favicon: ["mmh3"],
  CertIssuer: ["name"],
  DMARCPolicy: ["policy", "subdomain_policy"],
  SOAZone: ["zone", "mname", "serial", "rname", "refresh", "expire"],
  Subdomain: ["host"],
  DNSHost: ["host"],
  CertSAN: ["name"],
  IX: ["name", "country", "city"],
}

const CROSS_VIEW_LINKS: Record<string, string[]> = {
  ASN: ["bgp", "peering"],
  DNSHost: ["dns"],
  SOAZone: ["dns"],
  DMARCPolicy: ["dns"],
  CertSAN: ["cert"],
  CertIssuer: ["cert"],
  Subdomain: ["ct"],
  Hop: ["traceroute"],
  SharedHop: ["traceroute"],
  IX: ["peering"],
  JARM: ["infra"],
  Favicon: ["infra"],
  IP: ["infra"],
}

function formatPropValue(value: unknown): string {
  if (value === null || value === undefined || value === "") return "—"
  if (typeof value === "number") return String(value)
  return String(value)
}

export function GraphNodeInspector({
  selected,
  nodes,
  edges,
  onSelectNode,
  onFilterTarget,
  onJumpToView,
}: {
  selected: GraphNode | null
  nodes: GraphNode[]
  edges: GraphEdge[]
  onSelectNode: (node: GraphNode) => void
  onFilterTarget?: (targetId: string) => void
  onJumpToView?: (view: string) => void
}) {
  if (!selected) {
    return (
      <p className="text-muted-foreground">
        Click a node to inspect properties and jump to related targets.
      </p>
    )
  }

  const props = selected.props ?? {}
  const fields = INSPECTOR_FIELDS[selected.type] ?? Object.keys(props).slice(0, 8)
  const neighbors = edges
    .filter((edge) => edge.source === selected.id || edge.target === selected.id)
    .map((edge) => {
      const neighborID = edge.source === selected.id ? edge.target : edge.source
      return nodes.find((node) => node.id === neighborID)
    })
    .filter((node): node is GraphNode => !!node)
    .slice(0, 12)

  const targetID =
    selected.type === "Target"
      ? String(props.id ?? "")
      : String(props.target_id ?? "")

  const crossViews = CROSS_VIEW_LINKS[selected.type] ?? []

  return (
    <div className="flex flex-col gap-3 text-sm" data-testid="graph-node-inspector">
      <div>
        <div className="text-xs uppercase tracking-wide text-muted-foreground">
          {selected.type}
        </div>
        <div className="mt-1 font-medium break-all">{selected.label}</div>
      </div>

      {props.shared ? (
        <Badge variant="secondary" className="w-fit">
          Shared infrastructure
        </Badge>
      ) : null}
      {typeof props.hijack_risk === "string" && props.hijack_risk ? (
        <Badge variant="outline" className="w-fit font-mono text-xs">
          Hijack risk: {props.hijack_risk}
        </Badge>
      ) : null}

      {crossViews.length > 0 && onJumpToView ? (
        <div className="flex flex-wrap gap-1.5">
          {crossViews.map((view) => (
            <Button
              key={view}
              variant="outline"
              size="sm"
              className="h-7 text-xs capitalize"
              onClick={() => onJumpToView(view)}
            >
              Open {view} view
            </Button>
          ))}
        </div>
      ) : null}

      <dl className="flex flex-col gap-2">
        {fields.map((field) => (
          <div key={field} className="rounded-md border border-border/50 bg-muted/10 px-3 py-2">
            <dt className="text-[11px] uppercase tracking-wide text-muted-foreground">
              {formatFieldLabel(field)}
            </dt>
            <dd className="mt-1 flex items-start justify-between gap-2 break-all font-mono text-xs">
              <span>{formatPropValue(props[field])}</span>
              {typeof props[field] === "string" && props[field] ? (
                <CopyButton value={String(props[field])} />
              ) : null}
            </dd>
          </div>
        ))}
      </dl>

      {neighbors.length > 0 ? (
        <div className="flex flex-col gap-2">
          <div className="text-xs uppercase tracking-wide text-muted-foreground">
            Neighbors
          </div>
          <div className="flex flex-wrap gap-1.5">
            {neighbors.map((node) => (
              <Button
                key={node.id}
                variant="outline"
                size="sm"
                className="h-auto max-w-full py-1 font-mono text-[11px]"
                onClick={() => onSelectNode(node)}
                data-testid={`graph-neighbor-${node.type.toLowerCase()}`}
              >
                {node.type}: {node.label}
              </Button>
            ))}
          </div>
        </div>
      ) : null}

      {selected.type === "Target" && props.id ? (
        <Button
          variant="secondary"
          size="sm"
          className="w-full"
          render={
            <Link
              href={`/target?id=${String(props.id)}`}
              data-testid="graph-target-link"
            />
          }
        >
          Open target
        </Button>
      ) : null}

      {selected.type === "Snapshot" && props.id ? (
        <Button
          variant="secondary"
          size="sm"
          className="w-full"
          render={
            <Link href={`/snapshot?id=${String(props.id)}`} />
          }
        >
          Open snapshot
        </Button>
      ) : null}

      {targetID && selected.type !== "Target" && onFilterTarget ? (
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          onClick={() => onFilterTarget(targetID)}
        >
          Filter graph to target
        </Button>
      ) : null}
    </div>
  )
}