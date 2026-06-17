"use client"

import * as React from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { fetchApi } from "@/lib/api"
import { PlusIcon, XIcon, TagIcon } from "lucide-react"

export function TargetTags({ targetId, initialTags = [] }: { targetId: string, initialTags?: string[] }) {
  const [tags, setTags] = React.useState<string[]>(initialTags || [])
  const [adding, setAdding] = React.useState(false)
  const [newTag, setNewTag] = React.useState("")

  async function updateTags(newTags: string[]) {
    try {
      await fetchApi(`/api/targets/${targetId}/tags`, {
        method: "PUT",
        body: JSON.stringify({ tags: newTags }),
      })
      setTags(newTags)
    } catch (e) {
      console.error("Failed to update tags", e)
    }
  }

  function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    const t = newTag.trim().toLowerCase()
    if (!t || tags.includes(t)) {
      setAdding(false)
      setNewTag("")
      return
    }
    const updated = [...tags, t]
    setTags(updated)
    updateTags(updated)
    setNewTag("")
    setAdding(false)
  }

  function handleRemove(tagToRemove: string) {
    const updated = tags.filter((t) => t !== tagToRemove)
    setTags(updated)
    updateTags(updated)
  }

  return (
    <div className="flex flex-wrap items-center gap-2 mt-2">
      <TagIcon className="size-4 text-muted-foreground mr-1" />
      {tags.map((tag) => (
        <Badge key={tag} variant="outline" className="bg-primary/5 hover:bg-primary/10 transition-colors gap-1 pr-1 border-primary/20 text-primary">
          {tag}
          <div
            role="button"
            className="rounded-full hover:bg-primary/20 p-0.5 cursor-pointer"
            onClick={() => handleRemove(tag)}
          >
            <XIcon className="size-3" />
          </div>
        </Badge>
      ))}
      
      {adding ? (
        <form onSubmit={handleAdd} className="flex items-center gap-1">
          <Input
            autoFocus
            size={12}
            className="h-6 w-24 text-xs px-2 py-0"
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            onBlur={() => {
              if (!newTag) setAdding(false)
            }}
          />
        </form>
      ) : (
        <Button variant="ghost" size="sm" className="h-6 text-xs px-2 text-muted-foreground hover:text-foreground" onClick={() => setAdding(true)}>
          <PlusIcon className="size-3 mr-1" /> Add Tag
        </Button>
      )}
    </div>
  )
}
