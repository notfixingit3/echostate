"use client"

import * as React from "react"
import { KeyRoundIcon, Trash2Icon } from "lucide-react"

import { useAuth } from "@/components/auth-provider"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { HelpTip } from "@/components/help-tip"
import { fetchApi } from "@/lib/api"
import { registerPasskey, type AuthCredential } from "@/lib/auth"

export function AccountPasskeysPanel() {
  const { user, refresh } = useAuth()
  const [credentials, setCredentials] = React.useState<AuthCredential[]>([])
  const [loading, setLoading] = React.useState(true)
  const [nickname, setNickname] = React.useState("")
  const [busy, setBusy] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  async function loadCredentials() {
    if (!user) return
    setLoading(true)
    try {
      const session = await fetchApi<{
        authenticated: boolean
        credentials?: AuthCredential[]
      }>("/api/auth/session")
      setCredentials(session.credentials || [])
    } catch (e) {
      console.error(e)
      setCredentials([])
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    void loadCredentials()
  }, [user?.id])

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      await registerPasskey(nickname)
      setNickname("")
      await loadCredentials()
      await refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to register passkey")
    } finally {
      setBusy(false)
    }
  }

  async function handleDelete(credId: string) {
    if (!user || !confirm("Remove this passkey from your account?")) return
    setBusy(true)
    setError(null)
    try {
      await fetchApi(`/api/users/${user.id}/credentials/${credId}`, { method: "DELETE" })
      await loadCredentials()
      await refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to remove passkey")
    } finally {
      setBusy(false)
    }
  }

  if (!user) {
    return (
      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardContent className="py-8 text-sm text-muted-foreground">
          Sign in to manage passkeys on this account.
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Account
            <HelpTip id="settings.account" />
          </CardTitle>
          <CardDescription>
            Signed in as {user.display_name} ({user.role})
          </CardDescription>
        </CardHeader>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Passkeys on this account
            <HelpTip id="settings.passkeys" />
          </CardTitle>
          <CardDescription>
            Register a passkey for each device or security key you use to sign in.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {error ? (
            <p className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </p>
          ) : null}

          <form onSubmit={handleRegister} className="flex flex-col gap-4 rounded-lg border border-dashed p-4 sm:flex-row sm:items-end">
            <div className="flex-1 space-y-2">
              <Label htmlFor="passkey_nickname">Add passkey to this device</Label>
              <Input
                id="passkey_nickname"
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                placeholder="MacBook, YubiKey, phone…"
              />
            </div>
            <Button type="submit" disabled={busy} className="w-full sm:w-auto">
              <KeyRoundIcon data-icon="inline-start" />
              {busy ? "Registering…" : "Register passkey"}
            </Button>
          </form>

          {loading ? (
            <p className="text-sm text-muted-foreground">Loading passkeys…</p>
          ) : credentials.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No passkeys registered yet. Use the button above to enroll this browser or security key.
            </p>
          ) : (
            <div className="space-y-3">
              {credentials.map((cred) => (
                <div
                  key={cred.id}
                  className="flex flex-col gap-3 rounded-lg border p-4 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div>
                    <p className="font-medium">{cred.nickname || "Unnamed device"}</p>
                    <p className="text-xs text-muted-foreground">
                      Added {new Date(cred.created_at).toLocaleString()}
                      {cred.last_used_at
                        ? ` · Last used ${new Date(cred.last_used_at).toLocaleString()}`
                        : ""}
                    </p>
                  </div>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={busy}
                    onClick={() => handleDelete(cred.id)}
                  >
                    <Trash2Icon data-icon="inline-start" />
                    Remove
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}