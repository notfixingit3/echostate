"use client"

import * as React from "react"

export function PwaRegister() {
  React.useEffect(() => {
    if (process.env.NODE_ENV !== "production") return
    if (!("serviceWorker" in navigator)) return

    void navigator.serviceWorker.register("/sw.js", { scope: "/" }).catch(() => {
      // PWA install still works on some platforms without SW; ignore registration errors.
    })
  }, [])

  return null
}