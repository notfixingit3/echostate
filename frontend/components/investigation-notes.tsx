"use client"

import * as React from "react"
import Link from "next/link"
import { formatDistanceToNow } from "date-fns"

import { HelpTip } from "@/components/help-tip"
import { Alert, AlertDescription } from "@/components/ui/alert"
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
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import { fetchApi, ApiError } from "@/lib/api"
import type {
  InvestigationNote,
  NoteReference,
  NoteStatus,
  PaginatedResponse,
} from "@/lib/types"
import {
  AlertCircleIcon,
  ArchiveIcon,
  ArchiveRestoreIcon,
  LinkIcon,
  PencilIcon,
  PlusIcon,
  StickyNoteIcon,
  Trash2Icon,
  Undo2Icon,
} from "lucide-react"

type NoteAnchor = {
  targetId?: string
  snapshotId?: string
  collectionId?: string
  graphNode?: { id: string; label: string; type: string }
}

type InvestigationNotesProps = NoteAnchor & {
  title?: string
  compact?: boolean
  allowFreeform?: boolean
}

type ReferenceDraft = {
  type: NoteReference["type"]
  label: string
  url: string
  id: string
}

const STATUS_TABS: NoteStatus[] = ["active", "archived", "trashed"]

function referenceHref(ref: NoteReference): string | null {
  switch (ref.type) {
    case "url":
      return ref.url || null
    case "target":
      return ref.id ? `/target?id=${ref.id}` : null
    case "snapshot":
      return ref.id ? `/snapshot?id=${ref.id}` : null
    case "collection":
      return "/collections"
    default:
      return null
  }
}

function referenceLabel(ref: NoteReference): string {
  if (ref.label) return ref.label
  if (ref.url) return ref.url
  if (ref.id) return ref.id
  return ref.type
}

function NoteBody({ body }: { body: string }) {
  return (
    <div
      className="prose prose-sm dark:prose-invert max-w-none break-words text-sm leading-relaxed [&_a]:text-primary [&_a]:underline-offset-4 hover:[&_a]:underline"
      dangerouslySetInnerHTML={{ __html: body }}
    />
  )
}

function NoteReferences({ references }: { references: NoteReference[] }) {
  if (!references.length) return null
  return (
    <div className="flex flex-wrap gap-2">
      {references.map((ref, index) => {
        const href = referenceHref(ref)
        const label = referenceLabel(ref)
        return href ? (
          <Badge key={`${ref.type}-${index}`} variant="outline" className="gap-1 font-normal">
            <LinkIcon className="size-3" />
            <a href={href} target={ref.type === "url" ? "_blank" : undefined} rel="noopener noreferrer">
              {label}
            </a>
          </Badge>
        ) : (
          <Badge key={`${ref.type}-${index}`} variant="secondary" className="font-normal">
            {label}
          </Badge>
        )
      })}
    </div>
  )
}

export function InvestigationNotes({
  title = "Investigation notes",
  compact = false,
  targetId,
  snapshotId,
  collectionId,
  graphNode,
  allowFreeform = false,
}: InvestigationNotesProps) {
  const [status, setStatus] = React.useState<NoteStatus>("active")
  const [notes, setNotes] = React.useState<InvestigationNote[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [editingId, setEditingId] = React.useState<string | null>(null)
  const [draftTitle, setDraftTitle] = React.useState("")
  const [draftBody, setDraftBody] = React.useState("")
  const [draftRefs, setDraftRefs] = React.useState<ReferenceDraft[]>([])
  const [refType, setRefType] = React.useState<ReferenceDraft["type"]>("url")
  const [refLabel, setRefLabel] = React.useState("")
  const [refUrl, setRefUrl] = React.useState("")
  const [refId, setRefId] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [draftTargetId, setDraftTargetId] = React.useState(targetId || "")
  const [draftSnapshotId, setDraftSnapshotId] = React.useState(snapshotId || "")
  const [draftCollectionId, setDraftCollectionId] = React.useState(collectionId || "")

  const query = React.useMemo(() => {
    const params = new URLSearchParams({ status, limit: "50" })
    if (targetId) params.set("target_id", targetId)
    if (snapshotId) params.set("snapshot_id", snapshotId)
    if (collectionId) params.set("collection_id", collectionId)
    if (graphNode?.id) params.set("graph_node_id", graphNode.id)
    return params.toString()
  }, [status, targetId, snapshotId, collectionId, graphNode?.id])

  const loadNotes = React.useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await fetchApi<PaginatedResponse<InvestigationNote>>(
        `/api/notes?${query}`
      )
      setNotes(result.data ?? [])
    } catch (err) {
      setNotes([])
      setError(err instanceof ApiError ? err.message : "Failed to load notes")
    } finally {
      setLoading(false)
    }
  }, [query])

  React.useEffect(() => {
    void loadNotes()
  }, [loadNotes])

  function resetComposer() {
    setEditingId(null)
    setDraftTitle("")
    setDraftBody("")
    setDraftRefs([])
    setRefLabel("")
    setRefUrl("")
    setRefId("")
  }

  function startEdit(note: InvestigationNote) {
    setEditingId(note.id)
    setDraftTitle(note.title)
    setDraftBody(note.body)
    setDraftRefs(
      note.references.map((ref) => ({
        type: ref.type,
        label: ref.label || "",
        url: ref.url || "",
        id: ref.id || "",
      }))
    )
  }

  function addReferenceDraft() {
    const next: ReferenceDraft = {
      type: refType,
      label: refLabel.trim(),
      url: refUrl.trim(),
      id: refId.trim(),
    }
    if (refType === "url" && !next.url) return
    if (refType !== "url" && !next.id) return
    setDraftRefs((prev) => [...prev, next])
    setRefLabel("")
    setRefUrl("")
    setRefId("")
  }

  function buildPayload() {
    const references: NoteReference[] = draftRefs.map((ref) => ({
      type: ref.type,
      ...(ref.label ? { label: ref.label } : {}),
      ...(ref.url ? { url: ref.url } : {}),
      ...(ref.id ? { id: ref.id } : {}),
    }))
    return {
      title: draftTitle.trim(),
      body: draftBody.trim(),
      target_id: (targetId || draftTargetId.trim()) || undefined,
      snapshot_id: (snapshotId || draftSnapshotId.trim()) || undefined,
      collection_id: (collectionId || draftCollectionId.trim()) || undefined,
      graph_node_id: graphNode?.id,
      graph_node_label: graphNode?.label,
      graph_node_type: graphNode?.type,
      references,
    }
  }

  async function saveNote() {
    if (!draftBody.trim()) return
    setSaving(true)
    setError(null)
    try {
      const payload = buildPayload()
      if (editingId) {
        await fetchApi(`/api/notes/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(payload),
        })
      } else {
        await fetchApi("/api/notes", {
          method: "POST",
          body: JSON.stringify(payload),
        })
      }
      resetComposer()
      await loadNotes()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to save note")
    } finally {
      setSaving(false)
    }
  }

  async function transition(id: string, action: "archive" | "unarchive" | "trash" | "restore") {
    setError(null)
    try {
      await fetchApi(`/api/notes/${id}/${action}`, { method: "POST" })
      if (editingId === id) resetComposer()
      await loadNotes()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to update note")
    }
  }

  async function deleteForever(id: string) {
    if (!confirm("Permanently delete this note? This cannot be undone.")) return
    setError(null)
    try {
      await fetchApi(`/api/notes/${id}`, { method: "DELETE" })
      if (editingId === id) resetComposer()
      await loadNotes()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to delete note")
    }
  }

  const hasContextAnchor =
    !!targetId || !!snapshotId || !!collectionId || !!graphNode?.id
  const canCompose = hasContextAnchor || allowFreeform

  return (
    <Card className="border-border/60 bg-card/60">
      <CardHeader className={compact ? "pb-3" : undefined}>
        <CardTitle className="inline-flex items-center gap-1.5 text-lg">
          <StickyNoteIcon className="size-5 text-primary" />
          {title}
          <HelpTip id="notes.panel" />
        </CardTitle>
        <CardDescription>
          Analyst notes with links and references. Trashed notes are purged after 30 days.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {error ? (
          <Alert variant="destructive">
            <AlertCircleIcon />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}

        <Tabs
          value={status}
          onValueChange={(value) => value && setStatus(value as NoteStatus)}
        >
          <TabsList>
            {STATUS_TABS.map((tab) => (
              <TabsTrigger key={tab} value={tab} className="capitalize">
                {tab}
              </TabsTrigger>
            ))}
          </TabsList>

          {STATUS_TABS.map((tab) => (
            <TabsContent key={tab} value={tab} className="mt-4 flex flex-col gap-4">
              {tab === "active" && canCompose ? (
                <div className="rounded-lg border border-border/60 bg-muted/10 p-4">
                  <div className="mb-3 flex items-center gap-1.5 text-sm font-medium">
                    {editingId ? "Edit note" : "New note"}
                    <HelpTip id="notes.compose" />
                  </div>
                  <div className="grid gap-3">
                    {allowFreeform && !hasContextAnchor ? (
                      <div className="grid gap-3 sm:grid-cols-2">
                        <div className="space-y-2">
                          <Label htmlFor={`note-target-${tab}`}>Target ID</Label>
                          <Input
                            id={`note-target-${tab}`}
                            value={draftTargetId}
                            onChange={(e) => setDraftTargetId(e.target.value)}
                            placeholder="Target UUID"
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor={`note-collection-${tab}`}>Collection ID</Label>
                          <Input
                            id={`note-collection-${tab}`}
                            value={draftCollectionId}
                            onChange={(e) => setDraftCollectionId(e.target.value)}
                            placeholder="Collection UUID"
                          />
                        </div>
                        <div className="space-y-2 sm:col-span-2">
                          <Label htmlFor={`note-snapshot-${tab}`}>Snapshot ID (optional)</Label>
                          <Input
                            id={`note-snapshot-${tab}`}
                            value={draftSnapshotId}
                            onChange={(e) => setDraftSnapshotId(e.target.value)}
                            placeholder="Snapshot UUID"
                          />
                        </div>
                      </div>
                    ) : null}

                    <div className="space-y-2">
                      <Label htmlFor={`note-title-${tab}`}>Title</Label>
                      <Input
                        id={`note-title-${tab}`}
                        value={draftTitle}
                        onChange={(e) => setDraftTitle(e.target.value)}
                        placeholder="Optional title"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor={`note-body-${tab}`}>Body</Label>
                      <textarea
                        id={`note-body-${tab}`}
                        value={draftBody}
                        onChange={(e) => setDraftBody(e.target.value)}
                        rows={compact ? 4 : 5}
                        className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30"
                        placeholder='Write findings. Paste URLs or use <a href="https://...">label</a> for links.'
                      />
                    </div>

                    <div className="space-y-2 rounded-md border border-dashed border-border/60 p-3">
                      <div className="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                        References
                        <HelpTip id="notes.references" />
                      </div>
                      <div className="grid gap-2 sm:grid-cols-[120px_1fr_1fr_auto] sm:items-end">
                        <div className="space-y-1">
                          <Label className="text-xs">Type</Label>
                          <select
                            value={refType}
                            onChange={(e) => setRefType(e.target.value as ReferenceDraft["type"])}
                            className="h-9 w-full rounded-md border border-input bg-transparent px-2 text-sm"
                          >
                            <option value="url">URL</option>
                            <option value="target">Target</option>
                            <option value="snapshot">Snapshot</option>
                            <option value="collection">Collection</option>
                            <option value="graph_node">Graph node</option>
                          </select>
                        </div>
                        <div className="space-y-1">
                          <Label className="text-xs">Label</Label>
                          <Input value={refLabel} onChange={(e) => setRefLabel(e.target.value)} placeholder="Display label" />
                        </div>
                        <div className="space-y-1">
                          <Label className="text-xs">{refType === "url" ? "URL" : "ID"}</Label>
                          <Input
                            value={refType === "url" ? refUrl : refId}
                            onChange={(e) =>
                              refType === "url"
                                ? setRefUrl(e.target.value)
                                : setRefId(e.target.value)
                            }
                            placeholder={refType === "url" ? "https://..." : "UUID or node id"}
                          />
                        </div>
                        <Button type="button" variant="outline" size="sm" onClick={addReferenceDraft}>
                          <PlusIcon data-icon="inline-start" />
                          Add
                        </Button>
                      </div>
                      {draftRefs.length ? (
                        <div className="flex flex-wrap gap-2">
                          {draftRefs.map((ref, index) => (
                            <Badge key={`${ref.type}-${index}`} variant="secondary" className="gap-1">
                              {ref.label || ref.url || ref.id}
                              <button
                                type="button"
                                className="opacity-70 hover:opacity-100"
                                onClick={() =>
                                  setDraftRefs((prev) => prev.filter((_, i) => i !== index))
                                }
                              >
                                ×
                              </button>
                            </Badge>
                          ))}
                        </div>
                      ) : null}
                    </div>

                    <div className="flex flex-wrap gap-2">
                      <Button type="button" onClick={() => void saveNote()} disabled={saving || !draftBody.trim()}>
                        {saving ? "Saving…" : editingId ? "Save changes" : "Add note"}
                      </Button>
                      {editingId ? (
                        <Button type="button" variant="outline" onClick={resetComposer}>
                          Cancel
                        </Button>
                      ) : null}
                    </div>
                  </div>
                </div>
              ) : null}

              {loading ? (
                <p className="text-sm text-muted-foreground">Loading notes…</p>
              ) : notes.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  {tab === "active"
                    ? "No notes yet for this context."
                    : `No ${tab} notes.`}
                </p>
              ) : (
                <div className="flex flex-col gap-3">
                  {notes.map((note) => (
                    <div
                      key={note.id}
                      className="rounded-lg border border-border/60 bg-background/60 p-4"
                      data-testid={`note-card-${note.id}`}
                    >
                      <div className="flex flex-wrap items-start justify-between gap-2">
                        <div>
                          <div className="font-medium">{note.title || "Untitled note"}</div>
                          <div className="text-xs text-muted-foreground">
                            Updated {formatDistanceToNow(new Date(note.updated_at), { addSuffix: true })}
                          </div>
                        </div>
                        <div className="flex flex-wrap gap-1">
                          {tab === "active" ? (
                            <>
                              <Button type="button" variant="ghost" size="icon-sm" onClick={() => startEdit(note)} aria-label="Edit note">
                                <PencilIcon />
                              </Button>
                              <Button type="button" variant="ghost" size="icon-sm" onClick={() => void transition(note.id, "archive")} aria-label="Archive note">
                                <ArchiveIcon />
                              </Button>
                              <Button type="button" variant="ghost" size="icon-sm" onClick={() => void transition(note.id, "trash")} aria-label="Move to trash">
                                <Trash2Icon />
                              </Button>
                            </>
                          ) : null}
                          {tab === "archived" ? (
                            <>
                              <Button type="button" variant="ghost" size="sm" onClick={() => void transition(note.id, "unarchive")}>
                                <ArchiveRestoreIcon data-icon="inline-start" />
                                Restore
                              </Button>
                              <Button type="button" variant="ghost" size="icon-sm" onClick={() => void transition(note.id, "trash")} aria-label="Move to trash">
                                <Trash2Icon />
                              </Button>
                            </>
                          ) : null}
                          {tab === "trashed" ? (
                            <>
                              <Button type="button" variant="ghost" size="sm" onClick={() => void transition(note.id, "restore")}>
                                <Undo2Icon data-icon="inline-start" />
                                Restore
                              </Button>
                              <Button type="button" variant="ghost" size="sm" onClick={() => void deleteForever(note.id)}>
                                <Trash2Icon data-icon="inline-start" />
                                Delete forever
                              </Button>
                            </>
                          ) : null}
                        </div>
                      </div>

                      <div className="mt-3">
                        <NoteBody body={note.body} />
                      </div>

                      <div className="mt-3 flex flex-col gap-2">
                        <NoteReferences references={note.references} />
                        <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                          {note.target_host ? (
                            <Link href={`/target?id=${note.target_id}`} className="hover:text-primary">
                              Target: {note.target_host}
                            </Link>
                          ) : null}
                          {note.collection_name ? (
                            <Link href="/collections" className="hover:text-primary">
                              Collection: {note.collection_name}
                            </Link>
                          ) : null}
                          {note.graph_node_label ? (
                            <span>
                              Graph: {note.graph_node_type}: {note.graph_node_label}
                            </span>
                          ) : null}
                          {note.trashed_at && tab === "trashed" ? (
                            <span>
                              Trashed {formatDistanceToNow(new Date(note.trashed_at), { addSuffix: true })}
                              {" · "}
                              purged after 30 days
                            </span>
                          ) : null}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>
          ))}
        </Tabs>
      </CardContent>
    </Card>
  )
}