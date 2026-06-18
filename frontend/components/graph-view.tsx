"use client"

import * as React from "react"
import dynamic from "next/dynamic"
import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import {
  GraphHistoryBar,
  type GraphCompareMode,
} from "@/components/graph-history-bar"
import { GraphSavedViews } from "@/components/graph-saved-views"
import { HelpTip } from "@/components/help-tip"
import { GraphNodeInspector } from "@/components/graph-node-inspector"
import { HopGeoMap } from "@/components/hop-geo-map"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { fetchApi } from "@/lib/api"
import { exportCanvasPNG, exportGraphSVG } from "@/lib/graph-export"
import type {
  GraphIntelEvent,
  GraphNode,
  GraphPath,
  GraphResponse,
  PaginatedResponse,
  SavedGraphView,
  SnapshotSummary,
  TargetSummary,
} from "@/lib/types"
import {
  AlertCircleIcon,
  ArrowRightIcon,
  DownloadIcon,
  ImageIcon,
  InfoIcon,
  Maximize2Icon,
  MinusIcon,
  NetworkIcon,
  PlusIcon,
  RefreshCwIcon,
  RouteIcon,
  ScrollTextIcon,
  GlobeIcon,
  ShieldIcon,
} from "lucide-react"

const ForceGraph2D = dynamic(() => import("react-force-graph-2d"), {
  ssr: false,
})

function formatVantageLabel(vantageId: string, label?: string): string {
  const id = vantageId || "local"
  const text = (label || id).toLowerCase()
  if (id === "hackertarget" || id === "external" || text.includes("hackertarget")) {
    return "External vantage"
  }
  if (text.includes("hacker")) {
    return "External vantage"
  }
  return label || id
}

function matchesVantageFilter(pathVantage: string | undefined, filter: string): boolean {
  const id = pathVantage || "local"
  if (filter === "all") return true
  if (filter === "external" || filter === "hackertarget") {
    return id === "external" || id === "hackertarget"
  }
  return id === filter
}

function normalizeVantageId(vantageId: string): string {
  return vantageId === "hackertarget" ? "external" : vantageId
}

type GraphViewMode =
  | "infra"
  | "bgp"
  | "traceroute"
  | "peering"
  | "ct"
  | "dns"
  | "cert"

const GRAPH_VIEW_MODES: GraphViewMode[] = [
  "infra",
  "bgp",
  "traceroute",
  "peering",
  "ct",
  "dns",
  "cert",
]

const VIEW_META: Record<
  GraphViewMode,
  { title: string; description: string; legend: Record<string, string> }
> = {
  infra: {
    title: "Infrastructure graph",
    description:
      "Drag nodes to explore shared ASN, IP, JARM, favicon, and certificate links.",
    legend: {
      Target: "#0d9488",
      ASN: "#7c3aed",
      IP: "#d97706",
      JARM: "#059669",
      Favicon: "#db2777",
      CertIssuer: "#2563eb",
      Snapshot: "#64748b",
    },
  },
  bgp: {
    title: "BGP route graph",
    description:
      "Origin ASN, announced prefix, RIPEstat visible origins, and AS paths — edge colors reflect RPKI and hijack risk.",
    legend: {
      Target: "#0d9488",
      ASN: "#7c3aed",
      Prefix: "#ea580c",
      "Low risk": "#22c55e",
      "Medium risk": "#eab308",
      "High risk": "#ef4444",
    },
  },
  traceroute: {
    title: "Traceroute graph",
    description:
      "Hop chains from local and external vantages, with shared transit hops converging on SharedHop nodes.",
    legend: {
      Target: "#0d9488",
      Hop: "#0891b2",
      SharedHop: "#f97316",
      IP: "#d97706",
    },
  },
  peering: {
    title: "Peering / IX graph",
    description:
      "Internet exchange presence for each target ASN from PeeringDB.",
    legend: {
      Target: "#0d9488",
      ASN: "#7c3aed",
      IX: "#0ea5e9",
    },
  },
  ct: {
    title: "CT subdomain graph",
    description:
      "Certificate transparency subdomains linked to each target, with scan links when already tracked.",
    legend: {
      Target: "#0d9488",
      Subdomain: "#be185d",
    },
  },
  dns: {
    title: "DNS dependency graph",
    description:
      "Nameserver, mail exchanger, CNAME, SOA zone, and DMARC policy dependencies for each target.",
    legend: {
      Target: "#0d9488",
      DNSHost: "#4f46e5",
      DMARCPolicy: "#0f766e",
      SOAZone: "#7c3aed",
    },
  },
  cert: {
    title: "Certificate SAN graph",
    description:
      "TLS subject alternative names shared across targets, with links to tracked hosts.",
    legend: {
      Target: "#0d9488",
      CertSAN: "#9333ea",
      CertIssuer: "#2563eb",
    },
  },
}

type ForceNode = GraphNode & { color?: string; x?: number; y?: number }
type ForceLink = {
  source: string
  target: string
  label: string
  color?: string
}

export function GraphView() {
  const containerRef = React.useRef<HTMLDivElement>(null)
  const graphApiRef = React.useRef<{
    zoomToFit?: (ms?: number, padding?: number) => void
    zoom?: (k?: number, durationMs?: number) => void
    d3ReheatSimulation?: () => void
    graphData?: () => { nodes: ForceNode[]; links: ForceLink[] }
    graph2ScreenCoords?: (x: number, y: number) => { x: number; y: number }
  } | null>(null)
  const [size, setSize] = React.useState({ width: 0, height: 0 })
  const [graph, setGraph] = React.useState<GraphResponse | null>(null)
  const [targets, setTargets] = React.useState<TargetSummary[]>([])
  const [targetFilter, setTargetFilter] = React.useState("all")
  const [viewMode, setViewMode] = React.useState<GraphViewMode>("infra")
  const [selected, setSelected] = React.useState<GraphNode | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [vantageFilter, setVantageFilter] = React.useState("all")
  const [snapshotId, setSnapshotId] = React.useState<string | undefined>()
  const [compareSnapshotId, setCompareSnapshotId] = React.useState<string | undefined>()
  const [compareMode, setCompareMode] = React.useState<GraphCompareMode>("previous")
  const [pinnedNodes, setPinnedNodes] = React.useState<Record<string, { x: number; y: number }>>({})
  const [snapshotCount, setSnapshotCount] = React.useState<number | null>(null)
  const selectedRef = React.useRef<GraphNode | null>(null)

  React.useEffect(() => {
    selectedRef.current = selected
  }, [selected])

  const loadGraph = React.useCallback(
    async (
      targetId?: string,
      view: GraphViewMode = "infra",
      preserveSelection = false,
      nextSnapshotId?: string,
      nextCompareSnapshotId?: string,
      nextCompareMode?: GraphCompareMode,
      restoreNodeId?: string
    ): Promise<GraphResponse | null> => {
      setLoading(true)
      setError(null)
      const previousID = preserveSelection
        ? selectedRef.current?.id
        : restoreNodeId ?? null
      if (!preserveSelection && !restoreNodeId) {
        setSelected(null)
      }

      try {
        const params = new URLSearchParams()
        if (targetId) params.set("target_id", targetId)
        if (view !== "infra") params.set("view", view)
        if (nextSnapshotId) params.set("snapshot_id", nextSnapshotId)
        if (nextCompareSnapshotId) params.set("compare_snapshot_id", nextCompareSnapshotId)
        if (nextCompareMode) params.set("compare_mode", nextCompareMode)
        const query = params.toString() ? `?${params.toString()}` : ""
        const data = await fetchApi<GraphResponse>(`/api/graph${query}`)
        setGraph(data)
        if (previousID) {
          const match = data.nodes.find((node) => node.id === previousID)
          setSelected(match ?? null)
        }
        return data
      } catch (err) {
        setGraph(null)
        setError(err instanceof Error ? err.message : "Failed to load graph")
        return null
      } finally {
        setLoading(false)
      }
    },
    []
  )

  const handleSnapshotChange = React.useCallback(
    (
      nextSnapshotId?: string,
      nextCompareSnapshotId?: string,
      mode: GraphCompareMode = "previous"
    ) => {
      setSnapshotId(nextSnapshotId)
      setCompareSnapshotId(nextCompareSnapshotId)
      setCompareMode(mode)
      if (targetFilter === "all") return
      void loadGraph(
        targetFilter,
        viewMode,
        true,
        nextSnapshotId,
        nextCompareSnapshotId,
        mode
      )
    },
    [targetFilter, viewMode, loadGraph]
  )

  React.useEffect(() => {
    void loadGraph()
    void fetchApi<PaginatedResponse<TargetSummary>>("/api/targets?limit=100")
      .then((data) => setTargets(data.data ?? []))
      .catch(() => setTargets([]))
  }, [loadGraph])

  React.useEffect(() => {
    if (targetFilter === "all") {
      setSnapshotCount(null)
      return
    }
    let cancelled = false
    void fetchApi<PaginatedResponse<SnapshotSummary>>(
      `/api/targets/${targetFilter}/snapshots?limit=2`
    )
      .then((result) => {
        if (!cancelled) setSnapshotCount(result.data?.length ?? 0)
      })
      .catch(() => {
        if (!cancelled) setSnapshotCount(null)
      })
    return () => {
      cancelled = true
    }
  }, [targetFilter])

  React.useLayoutEffect(() => {
    const element = containerRef.current
    if (!element) return

    const updateSize = () => {
      const rect = element.getBoundingClientRect()
      const width = Math.max(320, Math.floor(rect.width))
      const height = Math.max(420, Math.floor(rect.height))
      setSize((current) =>
        current.width === width && current.height === height
          ? current
          : { width, height }
      )
    }

    updateSize()
    const observer = new ResizeObserver(() => updateSize())
    observer.observe(element)
    window.addEventListener("resize", updateSize)
    return () => {
      observer.disconnect()
      window.removeEventListener("resize", updateSize)
    }
  }, [viewMode])

  const legend = VIEW_META[viewMode].legend

  const forceData = React.useMemo(() => {
    if (!graph) {
      return { nodes: [] as ForceNode[], links: [] as ForceLink[] }
    }

    const nodes = graph.nodes ?? []
    const edges = graph.edges ?? []

    return {
      nodes: nodes.map((node) => {
        const pin = pinnedNodes[node.id]
        return {
          ...node,
          color: legend[node.type] ?? "#475569",
          ...(pin ? { fx: pin.x, fy: pin.y } : {}),
        }
      }),
      links: edges.map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        label: edge.label,
        color: edge.color,
      })),
    }
  }, [graph, legend, pinnedNodes])

  function handleFilterChange(value: string) {
    setTargetFilter(value)
    if (value === "all") {
      setSnapshotId(undefined)
      setCompareSnapshotId(undefined)
      void loadGraph(undefined, viewMode)
      return
    }
    void loadGraph(value, viewMode, false, snapshotId, compareSnapshotId, compareMode)
  }

  function handleViewChange(value: string) {
    const next = (value as GraphViewMode) || "infra"
    setViewMode(next)
    void loadGraph(
      targetFilter === "all" ? undefined : targetFilter,
      next,
      true,
      snapshotId,
      compareSnapshotId,
      compareMode
    )
  }

  function refreshGraph() {
    void loadGraph(
      targetFilter === "all" ? undefined : targetFilter,
      viewMode,
      true,
      snapshotId,
      compareSnapshotId,
      compareMode
    )
  }

  function resetLayout() {
    setPinnedNodes({})
    graphApiRef.current?.d3ReheatSimulation?.()
  }

  function exportFilename(ext: string) {
    const host =
      targetFilter === "all"
        ? "all-targets"
        : (targets.find((target) => target.id === targetFilter)?.host ?? "target")
            .replace(/[^a-zA-Z0-9.-]+/g, "-")
    return `echostate-graph-${host}-${viewMode}.${ext}`
  }

  function exportPNG() {
    const canvas = containerRef.current?.querySelector("canvas") ?? null
    exportCanvasPNG(canvas, exportFilename("png"))
  }

  const getViewSnapshot = React.useCallback(
    () => ({
      viewMode,
      targetFilter,
      vantageFilter,
      snapshotId,
      compareSnapshotId,
      compareMode,
      pinnedNodes,
      selectedNodeId: selected?.id,
    }),
    [
      viewMode,
      targetFilter,
      vantageFilter,
      snapshotId,
      compareSnapshotId,
      compareMode,
      pinnedNodes,
      selected,
    ]
  )

  async function applySavedView(view: SavedGraphView) {
    const target = view.target_id ?? "all"
    const nextView = (view.view_mode as GraphViewMode) || "infra"
    const nextCompare = (view.compare_mode as GraphCompareMode) || "previous"
    const nextSnapshot = view.snapshot_id ?? undefined
    const nextCompareSnapshot = view.compare_snapshot_id ?? undefined
    const nextPins = Object.fromEntries(
      Object.entries(view.pinned_nodes ?? {}).map(([id, pin]) => [
        id,
        { x: pin.x, y: pin.y },
      ])
    )

    setTargetFilter(target)
    setViewMode(nextView)
    setVantageFilter(view.vantage_filter || "all")
    setSnapshotId(nextSnapshot)
    setCompareSnapshotId(nextCompareSnapshot)
    setCompareMode(nextCompare)
    setPinnedNodes(nextPins)

    await loadGraph(
      target === "all" ? undefined : target,
      nextView,
      false,
      nextSnapshot,
      nextCompareSnapshot,
      nextCompare,
      view.selected_node_id ?? undefined
    )
  }

  function exportSVG() {
    const api = graphApiRef.current
    const data = api?.graphData?.()
    if (!api?.graph2ScreenCoords || !data) return

    const nodes = data.nodes
      .filter((node) => typeof node.x === "number" && typeof node.y === "number")
      .map((node) => {
        const screen = api.graph2ScreenCoords!(node.x!, node.y!)
        return {
          id: node.id,
          label: node.label,
          type: node.type,
          x: screen.x,
          y: screen.y,
        }
      })

    exportGraphSVG({
      nodes,
      links: data.links,
      width: size.width,
      height: size.height,
      nodeColor: (node) => graphNodeColor({ ...node, color: legend[node.type] }),
      filename: exportFilename("svg"),
    })
  }

  const filterLabel =
    targetFilter === "all"
      ? "All targets"
      : targets.find((target) => target.id === targetFilter)?.host ??
        "Selected target"

  const allPaths = graph?.paths ?? []
  const vantageOptions = React.useMemo(() => {
    const options = new Map<string, string>()
    for (const path of allPaths) {
      const rawId = path.vantage || "local"
      const id = normalizeVantageId(rawId)
      options.set(id, formatVantageLabel(rawId, path.vantage_label))
    }
    return Array.from(options.entries()).map(([id, label]) => ({ id, label }))
  }, [allPaths])

  const paths =
    viewMode === "traceroute" && vantageFilter !== "all"
      ? allPaths.filter((path) => matchesVantageFilter(path.vantage, vantageFilter))
      : allPaths

  const geoPoints =
    viewMode === "traceroute" && vantageFilter !== "all"
      ? (graph?.geo ?? []).filter((point) => {
          const match = allPaths.find(
            (path) =>
              path.target_id === point.target_id &&
              matchesVantageFilter(path.vantage, vantageFilter)
          )
          return !!match
        })
      : (graph?.geo ?? [])
  const asPaths = graph?.as_paths ?? []
  const sharedHops = graph?.shared_hops ?? []
  const sharedSANs = graph?.shared_sans ?? []
  const clusters = graph?.clusters ?? []
  const routeDiff = graph?.route_diff
  const vantageDivergence = graph?.vantage_divergence ?? []
  const peeringIX = graph?.peering_ix ?? []
  const intelEvents = graph?.events ?? []
  const topologyDiff = graph?.topology_diff
  const addedNodeIds = new Set(topologyDiff?.added_node_ids ?? [])
  const removedNodeIds = new Set(topologyDiff?.removed_node_ids ?? [])
  const addedEdgeIds = new Set(topologyDiff?.added_edge_ids ?? [])
  const removedEdgeIds = new Set(topologyDiff?.removed_edge_ids ?? [])
  const neighborIds = React.useMemo(() => {
    if (!selected || !graph?.edges) return new Set<string>()
    const ids = new Set<string>()
    for (const edge of graph.edges) {
      if (edge.source === selected.id) ids.add(edge.target)
      if (edge.target === selected.id) ids.add(edge.source)
    }
    return ids
  }, [selected, graph?.edges])
  const showRouteDiff =
    targetFilter !== "all" &&
    routeDiff?.has_previous &&
    (viewMode === "bgp" || viewMode === "traceroute")
  const showPathPanel = viewMode === "traceroute" && paths.length > 0

  return (
    <div className="flex w-full min-w-0 flex-col gap-6">
      <Tabs value={viewMode} onValueChange={handleViewChange} className="w-full min-w-0">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <TabsList className="h-auto w-full flex-wrap justify-start gap-1 bg-muted/50 p-1 lg:w-auto">
            <TabsTrigger value="infra" className="gap-1.5">
              <NetworkIcon className="size-3.5" />
              Infrastructure
            </TabsTrigger>
            <TabsTrigger value="bgp" className="gap-1.5">
              <NetworkIcon className="size-3.5" />
              BGP routes
            </TabsTrigger>
            <TabsTrigger value="traceroute" className="gap-1.5">
              <RouteIcon className="size-3.5" />
              Traceroute
            </TabsTrigger>
            <TabsTrigger value="peering" className="gap-1.5">
              <GlobeIcon className="size-3.5" />
              Peering / IX
            </TabsTrigger>
            <TabsTrigger value="ct" className="gap-1.5">
              <ScrollTextIcon className="size-3.5" />
              CT subdomains
            </TabsTrigger>
            <TabsTrigger value="dns" className="gap-1.5">
              <GlobeIcon className="size-3.5" />
              DNS
            </TabsTrigger>
            <TabsTrigger value="cert" className="gap-1.5">
              <ShieldIcon className="size-3.5" />
              Cert SANs
            </TabsTrigger>
          </TabsList>

          <div className="flex flex-col gap-2 lg:items-end">
            <GraphSavedViews getSnapshot={getViewSnapshot} onApply={applySavedView} />
            <div className="flex flex-wrap items-center gap-2">
            {viewMode === "traceroute" && vantageOptions.length > 1 ? (
              <Select
                value={vantageFilter}
                onValueChange={(value) => value && setVantageFilter(value)}
              >
                <SelectTrigger
                  className="w-[min(100%,220px)]"
                  data-testid="graph-vantage-filter"
                >
                  <SelectValue placeholder="Vantage">
                    <span className="truncate">
                      {vantageFilter === "all"
                        ? "All vantages"
                        : vantageOptions.find((v) => v.id === vantageFilter)?.label ??
                          vantageFilter}
                    </span>
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All vantages</SelectItem>
                  {vantageOptions.map((vantage) => (
                    <SelectItem key={vantage.id} value={vantage.id}>
                      {vantage.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : null}
            <Select
              value={targetFilter}
              onValueChange={(value) => value && handleFilterChange(value)}
            >
              <SelectTrigger
                className="w-[min(100%,280px)]"
                data-testid="graph-target-filter"
              >
                <SelectValue placeholder="Filter by target">
                  <span className="truncate">{filterLabel}</span>
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All targets</SelectItem>
                {targets.map((target) => (
                  <SelectItem key={target.id} value={target.id}>
                    {target.host}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              size="sm"
              onClick={refreshGraph}
              disabled={loading}
              data-testid="graph-refresh"
            >
              <RefreshCwIcon
                className={loading ? "size-4 animate-spin" : "size-4"}
              />
              Refresh
            </Button>
            </div>
          </div>
        </div>

        {targetFilter !== "all" ? (
          <GraphHistoryBar
            targetId={targetFilter}
            routeDiff={routeDiff}
            snapshotId={snapshotId}
            compareMode={compareMode}
            onCompareModeChange={setCompareMode}
            onSnapshotChange={handleSnapshotChange}
          />
        ) : null}

        {targetFilter !== "all" && snapshotCount !== null && snapshotCount < 2 ? (
          <Alert className="border-border/60 bg-card/60">
            <InfoIcon />
            <AlertDescription className="text-sm">
              Snapshot history and compare modes need multiple scans.{" "}
              <Link
                href={`/target?id=${targetFilter}`}
                className="font-medium text-primary underline-offset-4 hover:underline"
              >
                Rescan this target
              </Link>{" "}
              to build temporal topology on the graph.
            </AlertDescription>
          </Alert>
        ) : null}

        {graph?.stats ? (
          <div className="flex flex-wrap gap-2">
            {Object.entries(graph.stats).map(([type, count]) => (
              <Badge key={type} variant="secondary" className="font-mono text-xs">
                {type} {count}
              </Badge>
            ))}
          </div>
        ) : null}

        {GRAPH_VIEW_MODES.map((mode) => (
          <TabsContent key={mode} value={mode} className="mt-0">
            <div
              className={
                showPathPanel && mode === "traceroute"
                  ? "grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]"
                  : "grid gap-4 lg:grid-cols-[minmax(0,1fr)_280px]"
              }
            >
              <Card className="min-w-0 overflow-hidden border-border/60">
                <CardHeader className="border-b border-border/50 pb-4">
                  <CardTitle className="flex items-center gap-2 font-heading text-lg">
                    {mode === "traceroute" ? (
                      <RouteIcon className="size-5 text-primary" />
                    ) : (
                      <NetworkIcon className="size-5 text-primary" />
                    )}
                    {VIEW_META[mode].title}
                  </CardTitle>
                  <CardDescription>{VIEW_META[mode].description}</CardDescription>
                </CardHeader>
                <CardContent className="p-0">
                  <div
                    ref={mode === viewMode ? containerRef : undefined}
                    className="relative h-[clamp(420px,65vh,720px)] w-full overflow-hidden bg-muted/20"
                    data-testid="graph-canvas"
                  >
                    {error ? (
                      <div className="flex h-[420px] flex-col items-center justify-center gap-3 px-6 text-center">
                        <AlertCircleIcon className="size-8 text-destructive" />
                        <p className="text-sm text-muted-foreground">{error}</p>
                      </div>
                    ) : loading && mode === viewMode ? (
                      <div className="flex h-[420px] items-center justify-center text-sm text-muted-foreground">
                        Loading graph…
                      </div>
                    ) : mode === viewMode && forceData.nodes.length === 0 ? (
                      <div className="flex h-[420px] flex-col items-center justify-center gap-2 px-6 text-center text-sm text-muted-foreground">
                        <p>No graph data yet for this view.</p>
                        <p>Run a scan to populate Neo4j relationships.</p>
                      </div>
                    ) : mode === viewMode && size.width > 0 && size.height > 0 ? (
                      <>
                        <div className="absolute right-3 top-3 z-10 flex flex-wrap justify-end gap-1">
                          <Button type="button" variant="secondary" size="icon-sm" onClick={() => graphApiRef.current?.zoom?.(1.4)} aria-label="Zoom in">
                            <PlusIcon />
                          </Button>
                          <Button type="button" variant="secondary" size="icon-sm" onClick={() => graphApiRef.current?.zoom?.(0.7)} aria-label="Zoom out">
                            <MinusIcon />
                          </Button>
                          <Button type="button" variant="secondary" size="icon-sm" onClick={() => graphApiRef.current?.zoomToFit?.(400)} aria-label="Fit graph">
                            <Maximize2Icon />
                          </Button>
                          <Button type="button" variant="secondary" size="sm" onClick={resetLayout}>
                            Reset
                          </Button>
                          <Button
                            type="button"
                            variant="secondary"
                            size="icon-sm"
                            onClick={exportPNG}
                            disabled={forceData.nodes.length === 0}
                            aria-label="Export PNG"
                            data-testid="graph-export-png"
                          >
                            <ImageIcon />
                          </Button>
                          <Button
                            type="button"
                            variant="secondary"
                            size="icon-sm"
                            onClick={exportSVG}
                            disabled={forceData.nodes.length === 0}
                            aria-label="Export SVG"
                            data-testid="graph-export-svg"
                          >
                            <DownloadIcon />
                          </Button>
                        </div>
                        <div
                          className="h-full w-full"
                          style={{ width: size.width, height: size.height }}
                        >
                          <ForceGraph2D
                          ref={graphApiRef as never}
                          width={size.width}
                          height={size.height}
                          graphData={forceData}
                          enableZoomInteraction
                          enablePanInteraction
                          nodeLabel={(node) => {
                            const n = node as ForceNode
                            return `${n.type}: ${n.label}`
                          }}
                          nodeAutoColorBy="type"
                          nodeColor={(node) => {
                            const n = node as ForceNode
                            if (removedNodeIds.has(n.id)) return "#ef4444"
                            if (addedNodeIds.has(n.id)) return "#22c55e"
                            if (selected && n.id !== selected.id && !neighborIds.has(n.id)) {
                              return "rgba(148, 163, 184, 0.35)"
                            }
                            return graphNodeColor(n)
                          }}
                          nodeVal={(node) =>
                            graphNodeSize(node as ForceNode, selected?.id)
                          }
                          linkLabel="label"
                          linkDirectionalArrowLength={3.5}
                          linkDirectionalArrowRelPos={1}
                          linkColor={(link) => {
                            const l = link as ForceLink & { id?: string }
                            const edgeId = l.id ?? `${l.source}|${l.label}|${l.target}`
                            if (removedEdgeIds.has(edgeId)) return "#ef4444"
                            if (addedEdgeIds.has(edgeId)) return "#22c55e"
                            return l.color ?? "rgba(100, 116, 139, 0.45)"
                          }}
                          onNodeClick={(node) => setSelected(node as GraphNode)}
                          onNodeDragEnd={(node) => {
                            const n = node as ForceNode & {
                              x?: number
                              y?: number
                              fx?: number
                              fy?: number
                            }
                            if (typeof n.x !== "number" || typeof n.y !== "number") return
                            // Library clears fx/fy after drag unless they were set before drag started.
                            // Re-pin on the live simulation node so it does not snap back on release.
                            n.fx = n.x
                            n.fy = n.y
                            setPinnedNodes((prev) => ({
                              ...prev,
                              [n.id]: { x: n.x, y: n.y },
                            }))
                          }}

                          cooldownTicks={80}
                          d3AlphaDecay={0.03}
                          d3VelocityDecay={0.35}
                          />
                        </div>
                      </>
                    ) : mode === viewMode ? (
                      <div className="flex h-full min-h-[420px] items-center justify-center text-sm text-muted-foreground">
                        Preparing canvas…
                      </div>
                    ) : (
                      <div className="flex h-[420px] items-center justify-center text-sm text-muted-foreground">
                        Switch to this tab to render the graph.
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>

              <div className="flex flex-col gap-4">
                <Card className="border-border/60">
                  <CardHeader className="pb-3">
                    <CardTitle className="font-heading text-base">Legend</CardTitle>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    {Object.entries(VIEW_META[mode].legend).map(([type, color]) => (
                      <div key={type} className="flex items-center gap-2 text-sm">
                        <span
                          className="size-3 rounded-full"
                          style={{ backgroundColor: color }}
                        />
                        <span>{type}</span>
                      </div>
                    ))}
                    {mode === "bgp" ? (
                      <div className="mt-2 space-y-1 border-t border-border/50 pt-3 text-xs text-muted-foreground">
                        <p>
                          <span className="font-medium text-foreground">ANNOUNCES</span>{" "}
                          — Cymru origin ASN
                        </p>
                        <p>
                          <span className="font-medium text-foreground">
                            VISIBLE_ORIGIN
                          </span>{" "}
                          — RIPEstat observed origin
                        </p>
                        <p>
                          <span className="font-medium text-foreground">
                            AS_PATH_NEXT
                          </span>{" "}
                          — RIPEstat BGP state path
                        </p>
                      </div>
                    ) : null}
                    {mode === "ct" ? (
                      <div className="mt-2 border-t border-border/50 pt-3 text-xs text-muted-foreground">
                        <p>
                          <span className="font-medium text-foreground">SCANNED_AS</span>{" "}
                          links subdomains that are already tracked targets.
                        </p>
                      </div>
                    ) : null}
                    {mode === "cert" ? (
                      <div className="mt-2 border-t border-border/50 pt-3 text-xs text-muted-foreground">
                        <p>
                          <span className="font-medium text-foreground">SCANNED_AS</span>{" "}
                          links SANs that match already tracked target hosts.
                        </p>
                      </div>
                    ) : null}
                    {mode === "dns" ? (
                      <div className="mt-2 space-y-1 border-t border-border/50 pt-3 text-xs text-muted-foreground">
                        <p>
                          <span className="font-medium text-foreground">USES_NS</span>{" "}
                          — authoritative nameservers
                        </p>
                        <p>
                          <span className="font-medium text-foreground">USES_MX</span>{" "}
                          — mail exchangers
                        </p>
                        <p>
                          <span className="font-medium text-foreground">ALIASES_TO</span>{" "}
                          — CNAME target
                        </p>
                        <p>
                          <span className="font-medium text-foreground">HAS_SOA</span>{" "}
                          — authoritative zone (serial + primary NS)
                        </p>
                      </div>
                    ) : null}
                  </CardContent>

                  <CardHeader className="border-t border-border/50 pb-3 pt-4">
                    <CardTitle className="font-heading text-base">Selection</CardTitle>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3 text-sm">
                    {selected && mode === viewMode ? (
                      <GraphNodeInspector
                        selected={selected}
                        nodes={graph?.nodes ?? []}
                        edges={graph?.edges ?? []}
                        onSelectNode={setSelected}
                        onFilterTarget={(targetId) => handleFilterChange(targetId)}
                        onJumpToView={handleViewChange}
                      />
                    ) : (
                      <GraphNodeInspector
                        selected={null}
                        nodes={[]}
                        edges={[]}
                        onSelectNode={setSelected}
                      />
                    )}
                  </CardContent>
                </Card>

                {targetFilter !== "all" && intelEvents.length > 0 ? (
                  <Card className="border-border/60" data-testid="graph-intel-events">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        Intel events
                      </CardTitle>
                      <CardDescription>
                        Diff and enrichment signals for the selected target.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[360px] flex-col gap-2 overflow-auto">
                      {intelEvents.map((event) => (
                        <IntelEventRow key={event.id} event={event} />
                      ))}
                    </CardContent>
                  </Card>
                ) : null}

                {mode === "infra" && clusters.length > 0 ? (
                  <Card className="border-border/60">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        Infrastructure clusters
                      </CardTitle>
                      <CardDescription>
                        Targets sharing two or more infra signals (ASN, JARM, favicon, issuer, SAN, IP).
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[360px] flex-col gap-3 overflow-auto">
                      {clusters.map((cluster) => (
                        <div
                          key={cluster.id}
                          className="rounded-lg border border-border/50 bg-muted/10 px-3 py-2"
                          data-testid={`infra-cluster-${cluster.id}`}
                        >
                          <div className="text-sm font-medium">
                            {cluster.target_count} targets
                          </div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {cluster.target_labels.join(", ")}
                          </div>
                          {cluster.shared_signals?.length ? (
                            <div className="mt-2 space-y-1">
                              {cluster.shared_signals.map((signal) => (
                                <div
                                  key={signal}
                                  className="font-mono text-[11px] text-muted-foreground"
                                >
                                  {signal}
                                </div>
                              ))}
                            </div>
                          ) : null}
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                ) : null}

                {mode === "cert" && sharedSANs.length > 0 ? (
                  <Card className="border-border/60">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        Shared SANs
                      </CardTitle>
                      <CardDescription>
                        Certificate names used by multiple targets.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[320px] flex-col gap-2 overflow-auto">
                      {sharedSANs.map((san) => (
                        <div
                          key={san.name}
                          className="rounded-lg border border-fuchsia-500/30 bg-fuchsia-500/5 px-3 py-2"
                          data-testid={`shared-san-${san.name}`}
                        >
                          <div className="font-mono text-sm break-all">{san.name}</div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {san.target_count} targets: {san.target_labels.join(", ")}
                          </div>
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                ) : null}

                {showRouteDiff && (mode === "bgp" || mode === "traceroute") ? (
                  <RouteDiffCard
                    mode={mode}
                    routeDiff={routeDiff}
                  />
                ) : null}

                {mode === "traceroute" && vantageDivergence.length > 0 ? (
                  <Card className="border-border/60">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        Vantage divergence
                      </CardTitle>
                      <CardDescription>
                        Hops that differ between local and external traceroute probes.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[320px] flex-col gap-2 overflow-auto">
                      {vantageDivergence.map((item, index) => (
                        <div
                          key={`${item.target_id}-${index}`}
                          className="rounded-lg border border-cyan-500/30 bg-cyan-500/5 px-3 py-2 text-sm"
                          data-testid={`vantage-divergence-${item.target_id}-${index}`}
                        >
                          <div className="font-medium">{item.target_label}</div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {item.vantage_a} vs {item.vantage_b}
                            {item.diverges_at
                              ? ` · diverges at hop ${item.diverges_at}`
                              : ""}
                          </div>
                          {item.only_in_a?.length ? (
                            <div className="mt-1 font-mono text-xs">
                              Only in {item.vantage_a}: {item.only_in_a.join(", ")}
                            </div>
                          ) : null}
                          {item.only_in_b?.length ? (
                            <div className="mt-1 font-mono text-xs">
                              Only in {item.vantage_b}: {item.only_in_b.join(", ")}
                            </div>
                          ) : null}
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                ) : null}

                {mode === "traceroute" && sharedHops.length > 0 ? (
                  <Card className="border-border/60">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        Shared hops
                      </CardTitle>
                      <CardDescription>
                        Transit IPs used by multiple targets.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[320px] flex-col gap-2 overflow-auto">
                      {sharedHops.map((hop) => (
                        <div
                          key={hop.ip}
                          className="rounded-lg border border-amber-500/30 bg-amber-500/5 px-3 py-2"
                          data-testid={`shared-hop-${hop.ip}`}
                        >
                          <div className="font-mono text-sm">{hop.ip}</div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {hop.target_count} targets: {hop.target_labels.join(", ")}
                          </div>
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                ) : null}

                {mode === "traceroute" ? (
                  <Card className="border-border/60">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        Hop geography
                      </CardTitle>
                      <CardDescription>
                        RIPEstat geolocation for public traceroute hops.
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <HopGeoMap points={mode === viewMode ? geoPoints : []} />
                    </CardContent>
                  </Card>
                ) : null}

                {mode === "traceroute" && paths.length > 0 ? (
                  <Card className="border-border/60 xl:col-span-1">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        Hop paths
                      </CardTitle>
                      <CardDescription>
                        Ordered traceroute chains per target.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[520px] flex-col gap-4 overflow-auto">
                      {paths.map((path) => (
                        <TraceroutePathCard
                          key={`${path.target_id}-${path.vantage || "local"}`}
                          path={path}
                        />
                      ))}
                    </CardContent>
                  </Card>
                ) : null}

                {mode === "peering" && peeringIX.length > 0 ? (
                  <Card className="border-border/60">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        IX presence
                      </CardTitle>
                      <CardDescription>
                        PeeringDB internet exchange attachments for each ASN.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[520px] flex-col gap-2 overflow-auto">
                      {peeringIX.map((entry) => (
                        <div
                          key={`${entry.target_id}-${entry.ix_id}`}
                          className="rounded-lg border border-border/50 bg-muted/10 px-3 py-2"
                          data-testid={`peering-ix-${entry.ix_id}`}
                        >
                          <div className="text-sm font-medium">{entry.ix_name}</div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {entry.target_label} · AS{entry.asn}
                            {entry.city || entry.country
                              ? ` · ${[entry.city, entry.country].filter(Boolean).join(", ")}`
                              : ""}
                          </div>
                          {entry.speed_mbps ? (
                            <div className="mt-1 font-mono text-xs">
                              {entry.speed_mbps} Mbps
                            </div>
                          ) : null}
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                ) : null}

                {mode === "bgp" && asPaths.length > 0 ? (
                  <Card className="border-border/60">
                    <CardHeader className="pb-3">
                      <CardTitle className="font-heading text-base">
                        AS paths
                      </CardTitle>
                      <CardDescription>
                        Unique BGP paths from RIPEstat for each prefix.
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="flex max-h-[520px] flex-col gap-3 overflow-auto">
                      {asPaths.map((path, index) => (
                        <div
                          key={`${path.target_id}-${index}`}
                          className="rounded-lg border border-border/50 bg-muted/10 p-3"
                          data-testid={`as-path-${path.target_id}-${index}`}
                        >
                          <div className="text-sm font-medium">
                            {path.target_label}
                          </div>
                          {path.prefix ? (
                            <div className="mt-1 font-mono text-xs text-muted-foreground">
                              {path.prefix}
                            </div>
                          ) : null}
                          <div className="mt-2 font-mono text-xs leading-relaxed">
                            {path.path_label}
                          </div>
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                ) : null}
              </div>
            </div>
          </TabsContent>
        ))}
      </Tabs>
    </div>
  )
}

function graphNodeColor(node: ForceNode): string {
  if (node.type === "Prefix" && typeof node.props?.risk_color === "string") {
    return node.props.risk_color
  }
  if (
    (node.type === "Hop" || node.type === "SharedHop" || node.type === "CertSAN") &&
    node.props?.shared === true
  ) {
    return node.type === "CertSAN" ? "#c026d3" : "#f97316"
  }
  return node.color ?? "#475569"
}

function graphNodeSize(node: ForceNode, selectedId?: string): number {
  const base =
    node.type === "Target"
      ? 4
      : node.type === "SharedHop" || node.type === "CertSAN"
        ? 3
        : node.props?.shared === true
          ? 3
          : 2
  return node.id === selectedId ? base + 2 : base
}

function IntelEventRow({ event }: { event: GraphIntelEvent }) {
  const when =
    event.detected_at && event.detected_at > 0
      ? new Date(event.detected_at * 1000).toLocaleString()
      : null

  return (
    <div className="rounded-lg border border-border/50 bg-muted/10 px-3 py-2">
      <div className="flex flex-wrap items-center gap-2">
        <Badge variant="outline" className="font-mono text-[10px]">
          {event.type}
        </Badge>
        {event.severity ? (
          <Badge variant="secondary" className="text-[10px]">
            {event.severity}
          </Badge>
        ) : null}
        {event.source ? (
          <span className="text-[10px] uppercase tracking-wide text-muted-foreground">
            {event.source}
          </span>
        ) : null}
      </div>
      <p className="mt-2 text-sm leading-snug">{event.summary}</p>
      {when ? (
        <p className="mt-1 text-[11px] text-muted-foreground">{when}</p>
      ) : null}
      {event.snapshot_id ? (
        <Button
          variant="link"
          size="sm"
          className="mt-1 h-auto px-0 text-xs"
          render={<Link href={`/snapshot?id=${event.snapshot_id}`} />}
        >
          View snapshot
        </Button>
      ) : null}
    </div>
  )
}

function RouteDiffCard({
  mode,
  routeDiff,
}: {
  mode: "bgp" | "traceroute"
  routeDiff?: GraphResponse["route_diff"]
}) {
  if (!routeDiff?.has_previous) return null

  const bgp = routeDiff.bgp
  const trace = routeDiff.traceroute

  return (
    <Card className="border-border/60" data-testid="route-diff-card">
      <CardHeader className="pb-3">
        <CardTitle className="font-heading text-base">Route history</CardTitle>
        <CardDescription>
          Changes since the previous snapshot for this target.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3 text-sm">
        {mode === "bgp" && bgp ? (
          <div className="space-y-2">
            {bgp.changed ? (
              <Badge variant="secondary">BGP routing changed</Badge>
            ) : (
              <Badge variant="outline">No BGP routing changes</Badge>
            )}
            {bgp.hijack_risk_from || bgp.hijack_risk_to ? (
              <p>
                Hijack risk:{" "}
                <span className="font-mono">{bgp.hijack_risk_from || "—"}</span>
                {" → "}
                <span className="font-mono">{bgp.hijack_risk_to || "—"}</span>
              </p>
            ) : null}
            {bgp.rpki_from || bgp.rpki_to ? (
              <p>
                RPKI:{" "}
                <span className="font-mono">{bgp.rpki_from || "—"}</span>
                {" → "}
                <span className="font-mono">{bgp.rpki_to || "—"}</span>
              </p>
            ) : null}
            {bgp.visible_origins_added?.length ? (
              <p>
                Origins added:{" "}
                <span className="font-mono">
                  {bgp.visible_origins_added.map((asn) => `AS${asn}`).join(", ")}
                </span>
              </p>
            ) : null}
            {bgp.visible_origins_removed?.length ? (
              <p>
                Origins removed:{" "}
                <span className="font-mono">
                  {bgp.visible_origins_removed.map((asn) => `AS${asn}`).join(", ")}
                </span>
              </p>
            ) : null}
            {bgp.as_paths_added?.length ? (
              <div>
                <div className="text-xs uppercase tracking-wide text-muted-foreground">
                  AS paths added
                </div>
                {bgp.as_paths_added.map((path) => (
                  <div key={path} className="font-mono text-xs">
                    {path}
                  </div>
                ))}
              </div>
            ) : null}
          </div>
        ) : null}

        {mode === "traceroute" && trace ? (
          <div className="space-y-2">
            {trace.changed ? (
              <Badge variant="secondary">Traceroute changed</Badge>
            ) : (
              <Badge variant="outline">No traceroute changes</Badge>
            )}
            <p>
              Hop count:{" "}
              <span className="font-mono">{trace.hop_count_from ?? 0}</span>
              {" → "}
              <span className="font-mono">{trace.hop_count_to ?? 0}</span>
            </p>
            {trace.hops_added?.length ? (
              <p>
                Hops added:{" "}
                <span className="font-mono">{trace.hops_added.join(", ")}</span>
              </p>
            ) : null}
            {trace.hops_removed?.length ? (
              <p>
                Hops removed:{" "}
                <span className="font-mono">{trace.hops_removed.join(", ")}</span>
              </p>
            ) : null}
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}

function TraceroutePathCard({ path }: { path: GraphPath }) {
  return (
    <div
      className="rounded-lg border border-border/50 bg-muted/10 p-3"
      data-testid={`graph-path-${path.target_id}-${path.vantage || "local"}`}
    >
      <div className="mb-1 text-sm font-medium">{path.target_label}</div>
      <div className="mb-3 text-xs text-muted-foreground">
        {formatVantageLabel(path.vantage || "local", path.vantage_label)}
      </div>
      <div className="flex flex-wrap items-center gap-1.5">
        {path.hops.map((hop, index) => (
          <React.Fragment key={`${path.target_id}-${hop.hop}`}>
            <div className="flex min-w-0 flex-col rounded-md border border-border/50 bg-background/70 px-2 py-1.5">
              <span className="font-mono text-[10px] text-muted-foreground">
                hop {hop.hop}
              </span>
              <span className="font-mono text-xs">
                {hop.timeout ? "*" : hop.ip ?? "—"}
              </span>
              {hop.rtt_ms != null ? (
                <span className="text-[10px] text-muted-foreground">
                  {hop.rtt_ms.toFixed(1)} ms
                </span>
              ) : null}
              {hop.country ? (
                <span className="text-[10px] text-muted-foreground">
                  {hop.city ? `${hop.city}, ` : ""}
                  {hop.country}
                </span>
              ) : null}
            </div>
            {index < path.hops.length - 1 ? (
              <ArrowRightIcon className="size-3.5 shrink-0 text-muted-foreground" />
            ) : null}
          </React.Fragment>
        ))}
      </div>
    </div>
  )
}