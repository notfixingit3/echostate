"use client"

import { CircleHelpIcon } from "lucide-react"

import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { getHelpCopy } from "@/lib/help-copy"
import { cn } from "@/lib/utils"

export function HelpTip({
  id,
  content,
  className,
  side = "top",
}: {
  id?: string
  content?: string
  className?: string
  side?: "top" | "right" | "bottom" | "left"
}) {
  const text = content ?? (id ? getHelpCopy(id) : undefined)
  if (!text) return null

  return (
    <Tooltip>
      <TooltipTrigger
        type="button"
        className={cn(
          "inline-flex size-4 shrink-0 items-center justify-center rounded-full text-muted-foreground transition-colors hover:text-foreground",
          className
        )}
        aria-label="Help"
        onClick={(event) => event.stopPropagation()}
      >
        <CircleHelpIcon className="size-3.5" />
      </TooltipTrigger>
      <TooltipContent side={side} className="max-w-sm text-left leading-relaxed">
        {text}
      </TooltipContent>
    </Tooltip>
  )
}

export function LabelWithHelp({
  label,
  helpId,
  helpContent,
  className,
}: {
  label: React.ReactNode
  helpId?: string
  helpContent?: string
  className?: string
}) {
  return (
    <span className={cn("inline-flex items-center gap-1.5", className)}>
      {label}
      <HelpTip id={helpId} content={helpContent} />
    </span>
  )
}