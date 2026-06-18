"use client"

import * as React from "react"
import { DownloadIcon, UploadIcon, AlertCircleIcon } from "lucide-react"

import { HelpTip } from "@/components/help-tip"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { API_BASE, ApiError, fetchApi } from "@/lib/api"
import type { ExportBundle, ImportResult } from "@/lib/types"

export function SettingsDataPanel() {
  const [includeBlobs, setIncludeBlobs] = React.useState(true)
  const [includeConfig, setIncludeConfig] = React.useState(false)
  const [conflict, setConflict] = React.useState<"overwrite" | "skip">("overwrite")
  const [rebuildGraph, setRebuildGraph] = React.useState(true)
  const [exporting, setExporting] = React.useState(false)
  const [importing, setImporting] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [result, setResult] = React.useState<ImportResult | null>(null)
  const fileInputRef = React.useRef<HTMLInputElement>(null)

  async function handleExport() {
    setExporting(true)
    setError(null)
    try {
      const params = new URLSearchParams({
        include_blobs: String(includeBlobs),
        include_config: String(includeConfig),
      })
      const response = await fetch(`${API_BASE}/api/export?${params.toString()}`, {
        credentials: "include",
      })
      if (!response.ok) {
        const text = await response.text()
        throw new ApiError(text || response.statusText, response.status)
      }
      const bundle = (await response.json()) as ExportBundle
      const blob = new Blob([JSON.stringify(bundle, null, 2)], {
        type: "application/json",
      })
      const stamp = new Date().toISOString().replace(/[:.]/g, "-")
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement("a")
      anchor.href = url
      anchor.download = `echostate-export-${stamp}.json`
      anchor.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Export failed")
    } finally {
      setExporting(false)
    }
  }

  async function handleImportFile(file: File) {
    setImporting(true)
    setError(null)
    setResult(null)
    try {
      const text = await file.text()
      const parsed = JSON.parse(text) as ExportBundle | { bundle: ExportBundle }
      const bundle =
        "format_version" in parsed && parsed.format_version
          ? parsed
          : "bundle" in parsed
            ? parsed.bundle
            : null
      if (!bundle?.format_version) {
        throw new Error("File does not contain a valid EchoState export bundle")
      }

      const importResult = await fetchApi<ImportResult>("/api/import", {
        method: "POST",
        body: JSON.stringify({
          conflict,
          rebuild_graph: rebuildGraph,
          bundle,
        }),
      })
      setResult(importResult)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : err instanceof Error ? err.message : "Import failed")
    } finally {
      setImporting(false)
      if (fileInputRef.current) {
        fileInputRef.current.value = ""
      }
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Data transfer error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {result ? (
        <Alert>
          <AlertTitle>Import complete</AlertTitle>
          <AlertDescription className="space-y-1">
            <p>
              Imported {result.targets_imported} target{result.targets_imported === 1 ? "" : "s"},{" "}
              {result.snapshots_imported} snapshot{result.snapshots_imported === 1 ? "" : "s"},{" "}
              {result.collections_imported} collection{result.collections_imported === 1 ? "" : "s"},{" "}
              {result.notes_imported} note{result.notes_imported === 1 ? "" : "s"}.
            </p>
            {result.graph_sync_queued > 0 ? (
              <p>Queued {result.graph_sync_queued} Neo4j graph sync job{result.graph_sync_queued === 1 ? "" : "s"}.</p>
            ) : null}
            {result.warnings?.length ? (
              <ul className="list-disc pl-5 text-xs text-muted-foreground">
                {result.warnings.map((warning) => (
                  <li key={warning}>{warning}</li>
                ))}
              </ul>
            ) : null}
          </AlertDescription>
        </Alert>
      ) : null}

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-2">
            Export all
            <HelpTip id="settings.export_all" />
          </CardTitle>
          <CardDescription>
            Download targets, snapshots, collections, notes, and saved graph views as a portable JSON bundle.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex items-center justify-between gap-4 rounded-lg border border-border/50 bg-muted/10 px-4 py-3">
            <div>
              <Label htmlFor="include-blobs" className="text-sm font-medium">
                Include screenshot thumbnails
              </Label>
              <p className="text-xs text-muted-foreground">Adds blob data; export files will be larger.</p>
            </div>
            <Switch id="include-blobs" checked={includeBlobs} onCheckedChange={setIncludeBlobs} />
          </div>
          <div className="flex items-center justify-between gap-4 rounded-lg border border-border/50 bg-muted/10 px-4 py-3">
            <div>
              <Label htmlFor="include-config" className="text-sm font-medium">
                Include server configuration
              </Label>
              <p className="text-xs text-muted-foreground">Exports app settings and webhooks (not user accounts).</p>
            </div>
            <Switch id="include-config" checked={includeConfig} onCheckedChange={setIncludeConfig} />
          </div>
          <Button type="button" onClick={() => void handleExport()} disabled={exporting} className="w-fit">
            <DownloadIcon data-icon="inline-start" />
            {exporting ? "Preparing export…" : "Download export bundle"}
          </Button>
        </CardContent>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-2">
            Import bundle
            <HelpTip id="settings.import_bundle" />
          </CardTitle>
          <CardDescription>
            Restore investigation data from another EchoState install. User accounts and passkeys are never included.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="space-y-2">
            <Label>Conflict strategy</Label>
            <Select
              value={conflict}
              onValueChange={(value) => value && setConflict(value as "overwrite" | "skip")}
            >
              <SelectTrigger className="w-full max-w-sm">
                <SelectValue placeholder="Choose conflict strategy">
                  <span className="truncate">
                    {conflict === "overwrite" ? "Overwrite existing records" : "Skip existing records"}
                  </span>
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="overwrite">Overwrite existing records</SelectItem>
                <SelectItem value="skip">Skip existing records</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center justify-between gap-4 rounded-lg border border-border/50 bg-muted/10 px-4 py-3">
            <div>
              <Label htmlFor="rebuild-graph" className="text-sm font-medium">
                Rebuild Neo4j graph after import
              </Label>
              <p className="text-xs text-muted-foreground">Queues graph sync for imported snapshots when Neo4j is enabled.</p>
            </div>
            <Switch id="rebuild-graph" checked={rebuildGraph} onCheckedChange={setRebuildGraph} />
          </div>
          <input
            ref={fileInputRef}
            type="file"
            accept="application/json,.json"
            className="hidden"
            onChange={(event) => {
              const file = event.target.files?.[0]
              if (file) void handleImportFile(file)
            }}
          />
          <Button
            type="button"
            variant="outline"
            disabled={importing}
            className="w-fit"
            onClick={() => fileInputRef.current?.click()}
          >
            <UploadIcon data-icon="inline-start" />
            {importing ? "Importing…" : "Choose export file"}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}