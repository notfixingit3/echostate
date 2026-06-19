"use client"

import * as React from "react"

export function useTableSelection(visibleIds: readonly string[]) {
  const [selected, setSelected] = React.useState<Set<string>>(() => new Set())

  const visibleKey = React.useMemo(() => visibleIds.join("\n"), [visibleIds])

  React.useEffect(() => {
    const visible = new Set(visibleKey ? visibleKey.split("\n") : [])
    setSelected((current) => {
      const next = new Set([...current].filter((id) => visible.has(id)))
      return next.size === current.size ? current : next
    })
  }, [visibleKey])

  const toggle = React.useCallback((id: string) => {
    setSelected((current) => {
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const toggleAll = React.useCallback(() => {
    setSelected((current) => {
      if (visibleIds.length === 0) return current
      const allVisible = visibleIds.every((id) => current.has(id))
      if (allVisible) return new Set()
      return new Set(visibleIds)
    })
  }, [visibleIds])

  const clear = React.useCallback(() => setSelected(new Set()), [])

  const allSelected =
    visibleIds.length > 0 && visibleIds.every((id) => selected.has(id))
  const someSelected = visibleIds.some((id) => selected.has(id))

  return {
    selected,
    toggle,
    toggleAll,
    clear,
    allSelected,
    someSelected,
    count: selected.size,
  }
}