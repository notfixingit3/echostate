"use client"

import * as React from "react"
import { KeyRoundIcon, PlusIcon, UserPlusIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { HelpTip } from "@/components/help-tip"
import { fetchApi } from "@/lib/api"
import type { AuthUser } from "@/lib/auth"

type IssuedCode = {
  code: string
  expires_at: string
  purpose: string
  user_id: string
}

export function AuthUsersPanel() {
  const [users, setUsers] = React.useState<AuthUser[]>([])
  const [loading, setLoading] = React.useState(true)
  const [displayName, setDisplayName] = React.useState("")
  const [role, setRole] = React.useState<"admin" | "scanner">("scanner")
  const [issuedCode, setIssuedCode] = React.useState<IssuedCode | null>(null)
  const [busy, setBusy] = React.useState(false)

  async function loadUsers() {
    try {
      const data = await fetchApi<AuthUser[]>("/api/users")
      setUsers(data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    loadUsers()
  }, [])

  async function createUser(e: React.FormEvent) {
    e.preventDefault()
    if (!displayName.trim()) return
    setBusy(true)
    try {
      await fetchApi("/api/users", {
        method: "POST",
        body: JSON.stringify({ display_name: displayName.trim(), role }),
      })
      setDisplayName("")
      await loadUsers()
    } catch (e) {
      console.error(e)
    } finally {
      setBusy(false)
    }
  }

  async function issueCode(userId: string, recovery = false) {
    setBusy(true)
    setIssuedCode(null)
    try {
      const result = await fetchApi<IssuedCode>(`/api/users/${userId}/codes`, {
        method: "POST",
        body: JSON.stringify({ recovery }),
      })
      setIssuedCode(result)
    } catch (e) {
      console.error(e)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
      <CardHeader>
        <CardTitle className="inline-flex items-center gap-1.5">
          Users & enrollment codes
          <HelpTip id="settings.auth_users" />
        </CardTitle>
        <CardDescription>
          Create scanner or admin accounts and issue single-use enrollment or recovery codes.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <form onSubmit={createUser} className="grid gap-4 sm:grid-cols-[1fr_auto_auto] sm:items-end">
          <div className="space-y-2">
            <Label htmlFor="new_user_name">Display name</Label>
            <Input
              id="new_user_name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Analyst name"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="new_user_role">Role</Label>
            <Select value={role} onValueChange={(val) => val && setRole(val as "admin" | "scanner")}>
              <SelectTrigger id="new_user_role" className="w-full sm:w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="scanner">Scanner</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button type="submit" disabled={busy || !displayName.trim()} className="w-full sm:w-auto">
            <UserPlusIcon data-icon="inline-start" />
            Add user
          </Button>
        </form>

        {issuedCode ? (
          <div className="rounded-lg border border-primary/30 bg-primary/5 p-4 text-sm">
            <p className="font-medium">Code issued ({issuedCode.purpose})</p>
            <p className="mt-1 font-mono text-lg tracking-widest">{issuedCode.code}</p>
            <p className="mt-2 text-muted-foreground">
              Expires {new Date(issuedCode.expires_at).toLocaleString()}. Copy now — it cannot be shown again.
            </p>
          </div>
        ) : null}

        {loading ? (
          <p className="text-sm text-muted-foreground">Loading users…</p>
        ) : users.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No users yet. Bootstrap the first admin with{" "}
            <code className="rounded bg-muted px-1 py-0.5 text-xs">echostate auth bootstrap-admin</code>.
          </p>
        ) : (
          <div className="space-y-3">
            {users.map((user) => (
              <div
                key={user.id}
                className="flex flex-col gap-3 rounded-lg border p-4 sm:flex-row sm:items-center sm:justify-between"
              >
                <div>
                  <p className="font-medium">{user.display_name}</p>
                  <p className="text-xs text-muted-foreground">
                    {user.role} · {user.id}
                    {user.disabled ? " · disabled" : ""}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={busy}
                    onClick={() => issueCode(user.id, false)}
                  >
                    <PlusIcon data-icon="inline-start" />
                    Enrollment code
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={busy}
                    onClick={() => issueCode(user.id, true)}
                  >
                    <KeyRoundIcon data-icon="inline-start" />
                    Recovery code
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}