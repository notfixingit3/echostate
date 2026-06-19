"use client"

import * as React from "react"

import { HelpTip } from "@/components/help-tip"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useConfirm } from "@/components/confirm-dialog"
import { fetchApi, ApiError } from "@/lib/api"
import type { GraphCompareMode, GraphViewMode, SavedGraphView } from "@/lib/types"
import { BookmarkIcon, PlusIcon, SaveIcon, Trash2Icon } from "lucide-react"

export type GraphViewSnapshot = {
  viewMode: GraphViewMode
  targetFilter: string
  vantageFilter: string
  snapshotId?: string
  compareSnapshotId?: string
  compareMode: GraphCompareMode
  pinnedNodes: Record<string, { x: number; y: number }>
  selectedNodeId?: string
}

export function GraphSavedViews({
  getSnapshot,
  onApply,
}: {
  getSnapshot: () => GraphViewSnapshot
  onApply: (view: SavedGraphView) => void | Promise<void>
}) {
  const confirm = useConfirm()
  const [views, setViews] = React.useState<SavedGraphView[]>([])
  const [selectedId, setSelectedId] = React.useState("")
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [showCreate, setShowCreate] = React.useState(false)
  const [name, setName] = React.useState("")
  const [description, setDescription] = React.useState("")

  const loadViews = React.useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const rows = await fetchApi<SavedGraphView[]>("/api/graph/views")
      setViews(rows ?? [])
    } catch (err) {
      setViews([])
      setError(err instanceof ApiError ? err.message : "Failed to load saved views")
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void loadViews()
  }, [loadViews])

  function buildPayload(viewName: string, viewDescription: string) {
    const snapshot = getSnapshot()
    const targetId = snapshot.targetFilter === "all" ? undefined : snapshot.targetFilter
    return {
      name: viewName.trim(),
      description: viewDescription.trim(),
      view_mode: snapshot.viewMode,
      target_id: targetId,
      vantage_filter: snapshot.vantageFilter,
      snapshot_id: snapshot.snapshotId,
      compare_snapshot_id: snapshot.compareSnapshotId,
      compare_mode: snapshot.compareMode,
      pinned_nodes: snapshot.pinnedNodes,
      selected_node_id: snapshot.selectedNodeId,
    }
  }

  async function createView(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setSaving(true)
    setError(null)
    try {
      const created = await fetchApi<SavedGraphView>("/api/graph/views", {
        method: "POST",
        body: JSON.stringify(buildPayload(name, description)),
      })
      setSelectedId(created.id)
      setName("")
      setDescription("")
      setShowCreate(false)
      await loadViews()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to save view")
    } finally {
      setSaving(false)
    }
  }

  async function updateSelected() {
    if (!selectedId) return
    const current = views.find((view) => view.id === selectedId)
    if (!current) return
    setSaving(true)
    setError(null)
    try {
      await fetchApi(`/api/graph/views/${selectedId}`, {
        method: "PUT",
        body: JSON.stringify(buildPayload(current.name, current.description)),
      })
      await loadViews()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to update view")
    } finally {
      setSaving(false)
    }
  }

  async function deleteSelected() {
    if (!selectedId) return
    const current = views.find((view) => view.id === selectedId)
    if (!current) return
    const ok = await confirm({
      title: "Delete saved view?",
      description: `Delete saved view "${current.name}"?`,
      confirmLabel: "Delete",
      destructive: true,
    })
    if (!ok) return
    setSaving(true)
    setError(null)
    try {
      await fetchApi(`/api/graph/views/${selectedId}`, { method: "DELETE" })
      setSelectedId("")
      await loadViews()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to delete view")
    } finally {
      setSaving(false)
    }
  }

  async function applySelected() {
    if (!selectedId) return
    setError(null)
    try {
      const view = await fetchApi<SavedGraphView>(`/api/graph/views/${selectedId}`)
      await onApply(view)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to load view")
    }
  }

  const selected = views.find((view) => view.id === selectedId)

  return (
    <div className="flex flex-wrap items-center gap-2" data-testid="graph-saved-views">
      <Select value={selectedId} onValueChange={(value) => value && setSelectedId(value)}>
        <SelectTrigger className="w-[min(100%,240px)]">
          <SelectValue placeholder={loading ? "Loading views…" : "Saved views"}>
            <span className="inline-flex items-center gap-1.5 truncate">
              <BookmarkIcon className="size-3.5 shrink-0" />
              {selected ? selected.name : "Saved views"}
            </span>
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          {views.map((view) => (
            <SelectItem key={view.id} value={view.id}>
              {view.name}
              {view.target_host ? ` · ${view.target_host}` : ""}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() => void applySelected()}
        disabled={!selectedId || saving}
        data-testid="graph-view-apply"
      >
        Load
      </Button>

      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() => setShowCreate((open) => !open)}
        disabled={saving}
      >
        <PlusIcon data-icon="inline-start" />
        Save
      </Button>

      {selectedId ? (
        <>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => void updateSelected()}
            disabled={saving}
            data-testid="graph-view-update"
          >
            <SaveIcon data-icon="inline-start" />
            Update
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            onClick={() => void deleteSelected()}
            disabled={saving}
            aria-label="Delete saved view"
          >
            <Trash2Icon />
          </Button>
        </>
      ) : null}

      <HelpTip id="graph.saved_views" />

      {error ? (
        <span className="w-full text-xs text-destructive">{error}</span>
      ) : null}

      {showCreate ? (
        <form
          onSubmit={createView}
          className="flex w-full flex-col gap-2 rounded-lg border border-border/60 bg-card/80 p-3"
        >
          <div className="grid gap-2 sm:grid-cols-2">
            <div className="space-y-1">
              <Label htmlFor="graph-view-name">Name</Label>
              <Input
                id="graph-view-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. acme-bgp-drift"
                required
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="graph-view-description">Description</Label>
              <Input
                id="graph-view-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional"
              />
            </div>
          </div>
          <div className="flex gap-2">
            <Button type="submit" size="sm" disabled={saving || !name.trim()}>
              {saving ? "Saving…" : "Save current view"}
            </Button>
            <Button type="button" variant="ghost" size="sm" onClick={() => setShowCreate(false)}>
              Cancel
            </Button>
          </div>
        </form>
      ) : null}
    </div>
  )
}