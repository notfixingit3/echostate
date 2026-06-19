"use client"

import * as React from "react"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"

export type ConfirmOptions = {
  title?: string
  description: string
  confirmLabel?: string
  cancelLabel?: string
  destructive?: boolean
}

type ConfirmState = ConfirmOptions & {
  open: boolean
}

type ConfirmContextValue = {
  confirm: (options: ConfirmOptions | string) => Promise<boolean>
}

const ConfirmContext = React.createContext<ConfirmContextValue | null>(null)

const defaultState: ConfirmState = {
  open: false,
  title: "Are you sure?",
  description: "",
  confirmLabel: "Continue",
  cancelLabel: "Cancel",
  destructive: false,
}

export function ConfirmProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = React.useState<ConfirmState>(defaultState)
  const resolveRef = React.useRef<((value: boolean) => void) | null>(null)

  const finish = React.useCallback((result: boolean) => {
    const resolve = resolveRef.current
    if (!resolve) return
    resolveRef.current = null
    setState(defaultState)
    resolve(result)
  }, [])

  const confirm = React.useCallback((options: ConfirmOptions | string) => {
    const opts: ConfirmOptions =
      typeof options === "string" ? { description: options } : options

    return new Promise<boolean>((resolve) => {
      resolveRef.current = resolve
      setState({
        open: true,
        title: opts.title ?? "Are you sure?",
        description: opts.description,
        confirmLabel: opts.confirmLabel ?? "Continue",
        cancelLabel: opts.cancelLabel ?? "Cancel",
        destructive: opts.destructive ?? false,
      })
    })
  }, [])

  return (
    <ConfirmContext.Provider value={{ confirm }}>
      {children}
      <AlertDialog
        open={state.open}
        onOpenChange={(open) => {
          if (!open) finish(false)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{state.title}</AlertDialogTitle>
            <AlertDialogDescription>{state.description}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{state.cancelLabel}</AlertDialogCancel>
            <AlertDialogAction
              variant={state.destructive ? "destructive" : "default"}
              onClick={() => finish(true)}
            >
              {state.confirmLabel}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </ConfirmContext.Provider>
  )
}

export function useConfirm() {
  const context = React.useContext(ConfirmContext)
  if (!context) {
    throw new Error("useConfirm must be used within ConfirmProvider")
  }
  return context.confirm
}