"use client"

import { Button } from "@/components/ui/button"
import { Trash2Icon } from "lucide-react"

type BulkActionBarProps = {
  count: number
  noun: string
  deleting?: boolean
  onClear: () => void
  onDelete: () => void
}

export function BulkActionBar({
  count,
  noun,
  deleting = false,
  onClear,
  onDelete,
}: BulkActionBarProps) {
  if (count === 0) return null

  const label = count === 1 ? noun : `${noun}s`

  return (
    <div
      className="flex flex-wrap items-center gap-2 rounded-lg border border-border/60 bg-muted/40 px-3 py-2"
      data-testid="bulk-action-bar"
    >
      <span className="text-sm text-muted-foreground">
        {count} {label} selected
      </span>
      <Button variant="outline" size="sm" onClick={onClear} disabled={deleting}>
        Clear
      </Button>
      <Button
        variant="destructive"
        size="sm"
        disabled={deleting}
        data-testid="bulk-delete-button"
        onClick={onDelete}
      >
        <Trash2Icon data-icon="inline-start" />
        Delete selected
      </Button>
    </div>
  )
}