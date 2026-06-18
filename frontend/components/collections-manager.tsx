"use client"

import * as React from "react"
import Link from "next/link"

import { HelpTip } from "@/components/help-tip"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

import { fetchApi, ApiError } from "@/lib/api"
import type {
  CollectionRescanResponse,
  PaginatedResponse,
  TargetCollectionDetail,
  TargetCollectionSummary,
  TargetSummary,
} from "@/lib/types"
import {
  FolderKanbanIcon,
  PlusIcon,
  RefreshCwIcon,
  Trash2Icon,
  AlertCircleIcon,
} from "lucide-react"

export function CollectionsManager() {
  const [collections, setCollections] = React.useState<TargetCollectionSummary[]>([])
  const [selectedId, setSelectedId] = React.useState<string>("")
  const [detail, setDetail] = React.useState<TargetCollectionDetail | null>(null)
  const [targets, setTargets] = React.useState<TargetSummary[]>([])
  const [name, setName] = React.useState("")
  const [description, setDescription] = React.useState("")
  const [addTargetId, setAddTargetId] = React.useState("")
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [rescanning, setRescanning] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [rescanMessage, setRescanMessage] = React.useState<string | null>(null)

  const loadCollections = React.useCallback(async () => {
    const data = await fetchApi<TargetCollectionSummary[]>("/api/collections")
    setCollections(data)
    if (!selectedId && data.length > 0) {
      setSelectedId(data[0].id)
    }
  }, [selectedId])

  const loadDetail = React.useCallback(async (collectionId: string) => {
    if (!collectionId) {
      setDetail(null)
      return
    }
    const data = await fetchApi<TargetCollectionDetail>(
      `/api/collections/${collectionId}`
    )
    setDetail(data)
  }, [])

  React.useEffect(() => {
    let cancelled = false
    async function load() {
      setLoading(true)
      setError(null)
      try {
        const [collectionData, targetData] = await Promise.all([
          fetchApi<TargetCollectionSummary[]>("/api/collections"),
          fetchApi<PaginatedResponse<TargetSummary>>("/api/targets?limit=200"),
        ])
        if (cancelled) return
        setCollections(collectionData)
        setTargets(targetData.data ?? [])
        if (collectionData.length > 0) {
          setSelectedId((current) => current || collectionData[0].id)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof ApiError ? err.message : "Failed to load collections")
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  React.useEffect(() => {
    if (!selectedId) return
    void loadDetail(selectedId).catch((err) => {
      setError(err instanceof ApiError ? err.message : "Failed to load collection")
    })
  }, [selectedId, loadDetail])

  async function createCollection(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setSaving(true)
    setError(null)
    try {
      const created = await fetchApi<TargetCollectionDetail>("/api/collections", {
        method: "POST",
        body: JSON.stringify({ name: name.trim(), description: description.trim() }),
      })
      setName("")
      setDescription("")
      await loadCollections()
      setSelectedId(created.id)
      setDetail(created)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to create collection")
    } finally {
      setSaving(false)
    }
  }

  async function addTarget() {
    if (!selectedId || !addTargetId) return
    setSaving(true)
    setError(null)
    try {
      const updated = await fetchApi<TargetCollectionDetail>(
        `/api/collections/${selectedId}/targets`,
        {
          method: "POST",
          body: JSON.stringify({ target_id: addTargetId }),
        }
      )
      setDetail(updated)
      setAddTargetId("")
      await loadCollections()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to add target")
    } finally {
      setSaving(false)
    }
  }

  async function removeTarget(targetId: string) {
    if (!selectedId) return
    setSaving(true)
    setError(null)
    try {
      await fetchApi(`/api/collections/${selectedId}/targets/${targetId}`, {
        method: "DELETE",
      })
      await loadDetail(selectedId)
      await loadCollections()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to remove target")
    } finally {
      setSaving(false)
    }
  }

  async function rescanCollection() {
    if (!selectedId) return
    setRescanning(true)
    setError(null)
    setRescanMessage(null)
    try {
      const result = await fetchApi<CollectionRescanResponse>(
        `/api/collections/${selectedId}/rescan`,
        { method: "POST" }
      )
      setRescanMessage(
        result.jobs.length > 0
          ? `Queued ${result.jobs.length} rescan job${result.jobs.length === 1 ? "" : "s"}.`
          : "No targets in this collection to rescan."
      )
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to queue rescans")
    } finally {
      setRescanning(false)
    }
  }

  async function deleteCollection() {
    if (!selectedId || !detail) return
    if (!confirm(`Delete collection "${detail.name}"?`)) return
    setSaving(true)
    setError(null)
    try {
      await fetchApi(`/api/collections/${selectedId}`, { method: "DELETE" })
      setDetail(null)
      setSelectedId("")
      await loadCollections()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to delete collection")
    } finally {
      setSaving(false)
    }
  }

  const memberIds = new Set((detail?.targets ?? []).map((target) => target.id))
  const availableTargets = targets.filter((target) => !memberIds.has(target.id))

  return (
    <div className="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FolderKanbanIcon className="size-5 text-primary" />
            New collection
            <HelpTip id="collections.create" />
          </CardTitle>
          <CardDescription>
            Group related targets for watchlists and bulk rescans.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={createCollection} className="flex flex-col gap-4">
            <div className="space-y-2">
              <Label htmlFor="collection-name">Name</Label>
              <Input
                id="collection-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Acme perimeter"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="collection-description">Description</Label>
              <textarea
                id="collection-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional notes for this watchlist"
                rows={3}
                className="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
              />
            </div>
            <Button type="submit" disabled={saving || !name.trim()} className="w-fit">
              <PlusIcon data-icon="inline-start" />
              Create collection
            </Button>
          </form>
        </CardContent>
      </Card>

      <div className="flex flex-col gap-4">
        {error ? (
          <Alert variant="destructive">
            <AlertCircleIcon />
            <AlertTitle>Collection error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
        {rescanMessage ? (
          <Alert>
            <AlertTitle>Rescan queued</AlertTitle>
            <AlertDescription>{rescanMessage}</AlertDescription>
          </Alert>
        ) : null}

        <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
          <CardHeader className="flex flex-row items-start justify-between gap-4">
            <div>
              <CardTitle className="flex items-center gap-2">
                Collections
                <HelpTip id="collections.list" />
              </CardTitle>
              <CardDescription>
                Select a collection to manage members or queue a bulk rescan.
              </CardDescription>
            </div>
            {detail ? (
              <div className="inline-flex items-center gap-1.5">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void rescanCollection()}
                  disabled={rescanning || detail.targets.length === 0}
                >
                  <RefreshCwIcon data-icon="inline-start" />
                  Rescan all
                </Button>
                <HelpTip id="collections.rescan" />
              </div>
            ) : null}
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {loading ? (
              <p className="text-sm text-muted-foreground">Loading collections…</p>
            ) : collections.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No collections yet. Create one to start a watchlist.
              </p>
            ) : (
              <Select value={selectedId} onValueChange={(value) => value && setSelectedId(value)}>
                <SelectTrigger>
                  <SelectValue placeholder="Select collection" />
                </SelectTrigger>
                <SelectContent>
                  {collections.map((collection) => (
                    <SelectItem key={collection.id} value={collection.id}>
                      {collection.name} ({collection.target_count})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}

            {detail ? (
              <>
                <div className="flex items-start justify-between gap-3 rounded-lg border border-border/50 bg-muted/10 p-4">
                  <div>
                    <div className="font-medium">{detail.name}</div>
                    {detail.description ? (
                      <p className="mt-1 text-sm text-muted-foreground">{detail.description}</p>
                    ) : null}
                    <p className="mt-2 text-xs text-muted-foreground">
                      {detail.target_count} target{detail.target_count === 1 ? "" : "s"}
                    </p>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                    onClick={() => void deleteCollection()}
                    disabled={saving}
                    aria-label="Delete collection"
                  >
                    <Trash2Icon />
                  </Button>
                </div>

                <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
                  <div className="min-w-0 flex-1 space-y-2">
                    <Label>Add target</Label>
                    <Select value={addTargetId} onValueChange={(value) => value && setAddTargetId(value)}>
                      <SelectTrigger>
                        <SelectValue placeholder="Choose a target" />
                      </SelectTrigger>
                      <SelectContent>
                        {availableTargets.map((target) => (
                          <SelectItem key={target.id} value={target.id}>
                            {target.host}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <Button
                    type="button"
                    onClick={() => void addTarget()}
                    disabled={saving || !addTargetId}
                  >
                    Add
                  </Button>
                </div>

                <div className="flex flex-col gap-2">
                  {detail.targets.length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                      No targets in this collection yet.
                    </p>
                  ) : (
                    detail.targets.map((target) => (
                      <div
                        key={target.id}
                        className="flex items-center justify-between gap-3 rounded-lg border border-border/50 bg-background/60 px-3 py-2"
                      >
                        <div className="min-w-0">
                          <Link
                            href={`/target?id=${target.id}`}
                            className="font-mono text-sm hover:text-primary"
                          >
                            {target.host}
                          </Link>
                          <div className="mt-1 flex flex-wrap gap-2 text-xs text-muted-foreground">
                            <span>{target.snapshot_count} snapshots</span>
                            {(target.tags ?? []).map((tag) => (
                              <Badge key={tag} variant="secondary" className="text-[10px]">
                                {tag}
                              </Badge>
                            ))}
                          </div>
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => void removeTarget(target.id)}
                          disabled={saving}
                          aria-label={`Remove ${target.host}`}
                        >
                          <Trash2Icon />
                        </Button>
                      </div>
                    ))
                  )}
                </div>
              </>
            ) : null}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}