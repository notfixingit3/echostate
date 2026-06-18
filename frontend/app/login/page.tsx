"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { KeyRoundIcon, ShieldCheckIcon } from "lucide-react"

import { useAuth } from "@/components/auth-provider"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  getStoredUserId,
  loginWithPasskey,
  registerPasskey,
  verifyEnrollmentCode,
} from "@/lib/auth"
import { ApiError } from "@/lib/api"

export default function LoginPage() {
  const router = useRouter()
  const { config, user, refresh, setUser } = useAuth()
  const [code, setCode] = React.useState("")
  const [nickname, setNickname] = React.useState("")
  const [pendingUserId, setPendingUserId] = React.useState<string | null>(null)
  const [needsPasskey, setNeedsPasskey] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [busy, setBusy] = React.useState(false)

  const storedUserId = React.useMemo(() => getStoredUserId(), [])

  React.useEffect(() => {
    if (user) {
      router.replace("/")
    }
  }, [user, router])

  async function handleVerifyCode(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      const result = await verifyEnrollmentCode(code.trim())
      setPendingUserId(result.user.id)
      setNeedsPasskey(result.needs_passkey)
      setUser(result.user)
      if (!result.needs_passkey) {
        const login = await loginWithPasskey(result.user.id)
        if (!login.needsPasskey && login.user) {
          setUser(login.user)
          await refresh()
          router.replace("/")
        }
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Verification failed")
    } finally {
      setBusy(false)
    }
  }

  async function handlePasskeyLogin(userId: string) {
    setBusy(true)
    setError(null)
    try {
      const result = await loginWithPasskey(userId)
      if (result.needsPasskey) {
        setPendingUserId(userId)
        setNeedsPasskey(true)
        return
      }
      if (result.user) {
        setUser(result.user)
        await refresh()
        router.replace("/")
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Passkey sign-in failed")
    } finally {
      setBusy(false)
    }
  }

  async function handleRegisterPasskey(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      const result = await registerPasskey(nickname)
      setUser(result.user)
      await refresh()
      router.replace("/")
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Passkey registration failed")
    } finally {
      setBusy(false)
    }
  }

  const codeLength = config?.enrollment_code_length ?? 8

  return (
    <div className="container flex min-h-[70vh] max-w-lg items-center py-12">
      <Card className="w-full border-border/60 bg-card/70 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-2xl">
            <ShieldCheckIcon className="size-6 text-primary" />
            Sign in to EchoState
          </CardTitle>
          <CardDescription>
            Passkey-first access with single-use enrollment codes. Admins manage users and
            defaults from Settings.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {error ? (
            <p className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </p>
          ) : null}

          {!needsPasskey ? (
            <>
              {storedUserId ? (
                <div className="space-y-3">
                  <Button
                    type="button"
                    className="w-full"
                    disabled={busy}
                    onClick={() => handlePasskeyLogin(storedUserId)}
                  >
                    <KeyRoundIcon data-icon="inline-start" />
                    Sign in with passkey
                  </Button>
                  <p className="text-center text-xs text-muted-foreground">
                    Uses the passkey registered on this browser.
                  </p>
                </div>
              ) : null}

              <form onSubmit={handleVerifyCode} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="code">Enrollment or recovery code</Label>
                  <Input
                    id="code"
                    value={code}
                    onChange={(e) => setCode(e.target.value)}
                    placeholder={`${codeLength}-digit code or recovery code`}
                    autoComplete="one-time-code"
                    inputMode="text"
                  />
                </div>
                <Button type="submit" className="w-full" disabled={busy || !code.trim()}>
                  {busy ? "Verifying..." : "Verify code"}
                </Button>
              </form>
            </>
          ) : (
            <form onSubmit={handleRegisterPasskey} className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Register a passkey for this device to finish enrollment
                {pendingUserId ? ` (${pendingUserId.slice(0, 8)}…)` : ""}.
              </p>
              <div className="space-y-2">
                <Label htmlFor="nickname">Device nickname (optional)</Label>
                <Input
                  id="nickname"
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                  placeholder="MacBook, YubiKey, phone…"
                />
              </div>
              <Button type="submit" className="w-full" disabled={busy}>
                <KeyRoundIcon data-icon="inline-start" />
                {busy ? "Registering..." : "Create passkey"}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}