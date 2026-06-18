"use client"

import * as React from "react"
import { useTheme } from "next-themes"
import {
  KeyRoundIcon,
  MonitorIcon,
  MoonIcon,
  PencilIcon,
  PlusIcon,
  SaveIcon,
  SunIcon,
  Trash2Icon,
  XIcon,
} from "lucide-react"

import { useAuth } from "@/components/auth-provider"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { HelpTip } from "@/components/help-tip"
import { fetchApi } from "@/lib/api"
import {
  issueDeviceCode,
  registerPasskey,
  renameCredential,
  updateProfile,
  type AuthCredential,
  type IssuedEnrollmentCode,
  type UserTheme,
} from "@/lib/auth"

const THEME_OPTIONS: { value: UserTheme; label: string; icon: typeof SunIcon }[] = [
  { value: "light", label: "Light", icon: SunIcon },
  { value: "dark", label: "Dark", icon: MoonIcon },
  { value: "system", label: "System", icon: MonitorIcon },
]

const COMMON_TIMEZONES = [
  "UTC",
  "America/Los_Angeles",
  "America/Denver",
  "America/Chicago",
  "America/New_York",
  "America/Toronto",
  "America/Sao_Paulo",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Europe/Amsterdam",
  "Asia/Dubai",
  "Asia/Kolkata",
  "Asia/Singapore",
  "Asia/Tokyo",
  "Australia/Sydney",
  "Pacific/Auckland",
] as const

function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC"
  } catch {
    return "UTC"
  }
}

function formatTimestamp(value: string, timezone: string) {
  try {
    return new Intl.DateTimeFormat(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
      timeZone: timezone || "UTC",
    }).format(new Date(value))
  } catch {
    return new Date(value).toLocaleString()
  }
}

export function AccountProfilePanel() {
  const { user, refresh, setUser } = useAuth()
  const { setTheme } = useTheme()
  const [credentials, setCredentials] = React.useState<AuthCredential[]>([])
  const [loading, setLoading] = React.useState(true)
  const [displayName, setDisplayName] = React.useState("")
  const [nickname, setNickname] = React.useState("")
  const [timezone, setTimezone] = React.useState("UTC")
  const [theme, setThemePreference] = React.useState<UserTheme>("system")
  const [issuedDeviceCode, setIssuedDeviceCode] = React.useState<IssuedEnrollmentCode | null>(null)
  const [savingAccount, setSavingAccount] = React.useState(false)
  const [savingPreferences, setSavingPreferences] = React.useState(false)
  const [editingCredId, setEditingCredId] = React.useState<string | null>(null)
  const [editingNickname, setEditingNickname] = React.useState("")
  const [busy, setBusy] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [message, setMessage] = React.useState<string | null>(null)

  const timezoneOptions = React.useMemo(() => {
    const detected = browserTimezone()
    const options = new Set<string>(COMMON_TIMEZONES)
    if (user?.timezone) options.add(user.timezone)
    options.add(detected)
    return Array.from(options).sort()
  }, [user?.timezone])

  React.useEffect(() => {
    setDisplayName(user?.display_name || "")
    setTimezone(user?.timezone || browserTimezone())
    setThemePreference(user?.theme || "system")
  }, [user?.display_name, user?.timezone, user?.theme])

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
    setMessage(null)
    try {
      await registerPasskey(nickname)
      setNickname("")
      setMessage("Passkey registered on this device.")
      await loadCredentials()
      await refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to register passkey")
    } finally {
      setBusy(false)
    }
  }

  async function handleSaveAccount(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = displayName.trim()
    if (!trimmed) {
      setError("Display name cannot be empty.")
      return
    }
    setSavingAccount(true)
    setError(null)
    setMessage(null)
    try {
      const result = await updateProfile({ display_name: trimmed })
      setUser(result.user)
      setMessage("Account updated.")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update account")
    } finally {
      setSavingAccount(false)
    }
  }

  async function handleIssueDeviceCode() {
    setBusy(true)
    setError(null)
    setMessage(null)
    setIssuedDeviceCode(null)
    try {
      const result = await issueDeviceCode()
      setIssuedDeviceCode(result)
      setMessage("Device enrollment code issued.")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to issue device code")
    } finally {
      setBusy(false)
    }
  }

  async function handleSavePreferences(e: React.FormEvent) {
    e.preventDefault()
    setSavingPreferences(true)
    setError(null)
    setMessage(null)
    try {
      const result = await updateProfile({ timezone, theme })
      setUser(result.user)
      setTheme(theme)
      setMessage("Profile preferences saved.")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save preferences")
    } finally {
      setSavingPreferences(false)
    }
  }

  async function handleRename(credId: string) {
    if (!user) return
    setBusy(true)
    setError(null)
    setMessage(null)
    try {
      await renameCredential(user.id, credId, editingNickname)
      setEditingCredId(null)
      setEditingNickname("")
      setMessage("Passkey renamed.")
      await loadCredentials()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to rename passkey")
    } finally {
      setBusy(false)
    }
  }

  async function handleDelete(credId: string) {
    if (!user || !confirm("Remove this passkey from your account?")) return
    setBusy(true)
    setError(null)
    setMessage(null)
    try {
      await fetchApi(`/api/users/${user.id}/credentials/${credId}`, { method: "DELETE" })
      setMessage("Passkey removed.")
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
          Sign in to manage your profile.
        </CardContent>
      </Card>
    )
  }

  const displayTimezone = user.timezone || timezone

  return (
    <div className="space-y-6">
      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Account
            <HelpTip id="settings.profile" />
          </CardTitle>
          <CardDescription>
            Role: {user.role}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSaveAccount} className="flex flex-col gap-4 sm:flex-row sm:items-end">
            <div className="flex-1 space-y-2">
              <Label htmlFor="display_name">Display name</Label>
              <Input
                id="display_name"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="Your name"
                maxLength={80}
              />
            </div>
            <Button type="submit" disabled={savingAccount || !displayName.trim()} className="w-full sm:w-auto">
              <SaveIcon data-icon="inline-start" />
              {savingAccount ? "Saving…" : "Save name"}
            </Button>
          </form>
        </CardContent>
      </Card>

      {error ? (
        <p className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </p>
      ) : null}
      {message ? (
        <p className="rounded-md border border-primary/30 bg-primary/10 px-3 py-2 text-sm text-primary">
          {message}
        </p>
      ) : null}

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Preferences
            <HelpTip id="settings.preferences" />
          </CardTitle>
          <CardDescription>Theme and timezone follow you across browsers when signed in.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSavePreferences} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="theme">Theme</Label>
              <Select value={theme} onValueChange={(val) => val && setThemePreference(val as UserTheme)}>
                <SelectTrigger id="theme" className="w-full">
                  <SelectValue placeholder="Select theme" />
                </SelectTrigger>
                <SelectContent>
                  {THEME_OPTIONS.map((option) => {
                    const Icon = option.icon
                    return (
                      <SelectItem key={option.value} value={option.value}>
                        <span className="inline-flex items-center gap-2">
                          <Icon className="size-4" />
                          {option.label}
                        </span>
                      </SelectItem>
                    )
                  })}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="timezone">Timezone</Label>
              <Select value={timezone} onValueChange={(val) => val && setTimezone(val)}>
                <SelectTrigger id="timezone" className="w-full">
                  <SelectValue placeholder="Select timezone" />
                </SelectTrigger>
                <SelectContent>
                  {timezoneOptions.map((tz) => (
                    <SelectItem key={tz} value={tz}>
                      {tz}
                      {tz === browserTimezone() ? " (browser)" : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button type="submit" disabled={savingPreferences} className="w-full sm:w-auto">
              <SaveIcon data-icon="inline-start" />
              {savingPreferences ? "Saving…" : "Save preferences"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Passkeys & devices
            <HelpTip id="settings.passkeys" />
          </CardTitle>
          <CardDescription>
            Register and manage the passkeys you use to sign in on each browser or security key.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="rounded-lg border border-dashed p-4 space-y-4">
            <div>
              <p className="inline-flex items-center gap-1.5 font-medium">
                Enroll another device
                <HelpTip id="settings.device_code" />
              </p>
              <p className="text-sm text-muted-foreground">
                Issue a one-time code for a phone or browser that does not have your passkey yet.
              </p>
            </div>
            <Button type="button" variant="outline" disabled={busy} onClick={() => void handleIssueDeviceCode()}>
              <PlusIcon data-icon="inline-start" />
              {busy ? "Issuing…" : "Issue device code"}
            </Button>
            {issuedDeviceCode ? (
              <div className="rounded-lg border border-primary/30 bg-primary/5 p-4 text-sm">
                <p className="font-medium">Device code ({issuedDeviceCode.purpose})</p>
                <p className="mt-1 font-mono text-lg tracking-widest">{issuedDeviceCode.code}</p>
                <p className="mt-2 text-muted-foreground">
                  Expires {formatTimestamp(issuedDeviceCode.expires_at, displayTimezone)}. Copy now — it cannot be
                  shown again. Enter at <code className="rounded bg-muted px-1 py-0.5 text-xs">/login</code> on the new
                  device.
                </p>
              </div>
            ) : null}
          </div>

          <form onSubmit={handleRegister} className="flex flex-col gap-4 rounded-lg border border-dashed p-4 sm:flex-row sm:items-end">
            <div className="flex-1 space-y-2">
              <Label htmlFor="passkey_nickname">Add passkey on this device</Label>
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
              No passkeys registered yet. Use the form above to enroll this browser or security key.
            </p>
          ) : (
            <div className="space-y-3">
              {credentials.map((cred) => (
                <div
                  key={cred.id}
                  className="flex flex-col gap-3 rounded-lg border p-4 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div className="min-w-0 flex-1 space-y-2">
                    {editingCredId === cred.id ? (
                      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                        <Input
                          value={editingNickname}
                          onChange={(e) => setEditingNickname(e.target.value)}
                          placeholder="Device nickname"
                          className="sm:max-w-xs"
                        />
                        <div className="flex gap-2">
                          <Button
                            type="button"
                            size="sm"
                            disabled={busy}
                            onClick={() => handleRename(cred.id)}
                          >
                            <SaveIcon data-icon="inline-start" />
                            Save
                          </Button>
                          <Button
                            type="button"
                            size="sm"
                            variant="ghost"
                            disabled={busy}
                            onClick={() => {
                              setEditingCredId(null)
                              setEditingNickname("")
                            }}
                          >
                            <XIcon data-icon="inline-start" />
                            Cancel
                          </Button>
                        </div>
                      </div>
                    ) : (
                      <p className="font-medium">{cred.nickname || "Unnamed device"}</p>
                    )}
                    <p className="text-xs text-muted-foreground">
                      Added {formatTimestamp(cred.created_at, displayTimezone)}
                      {cred.last_used_at
                        ? ` · Last used ${formatTimestamp(cred.last_used_at, displayTimezone)}`
                        : ""}
                    </p>
                  </div>
                  {editingCredId !== cred.id ? (
                    <div className="flex gap-2">
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        disabled={busy}
                        onClick={() => {
                          setEditingCredId(cred.id)
                          setEditingNickname(cred.nickname || "")
                        }}
                      >
                        <PencilIcon data-icon="inline-start" />
                        Rename
                      </Button>
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
                  ) : null}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}