"use client"

import * as React from "react"
import { CopyIcon, CheckIcon } from "lucide-react"

import { Button } from "@/components/ui/button"

export function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      className="size-6 ml-2 text-muted-foreground hover:text-foreground"
      onClick={handleCopy}
      title="Copy to clipboard"
    >
      {copied ? <CheckIcon className="size-3 text-green-500" /> : <CopyIcon className="size-3" />}
      <span className="sr-only">Copy</span>
    </Button>
  )
}
