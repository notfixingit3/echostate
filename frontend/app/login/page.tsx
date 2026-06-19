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
  clearStoredUserId,
  getStoredUserId,
  loginWithPasskey,
  registerPasskey,
  storeUserId,
  verifyEnrollmentCode,
} from "@/lib/auth"
import { ApiError } from "@/lib/api"

type Step = "passkey" | "code" | "code-verified" | "register"

export default function LoginPage() {
  const router = useRouter()
  const { config, user, refresh, setUser } = useAuth()
  const [storedUserId, setStoredUserId] = React.useState<string | null>(null)
  const [code, setCode] = React.useState("")
  const [nickname, setNickname] = React.useState("")
  const [pendingUserId, setPendingUserId] = React.useState<string | null>(null)
  const [step, setStep] = React.useState<Step>("passkey")
  const [error, setError] = React.useState<string | null>(null)
  const [busy, setBusy] = React.useState(false)

  React.useEffect(() => {
    const saved = getStoredUserId()
    setStoredUserId(saved)
    setStep(saved ? "passkey" : "code")
  }, [])

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
      storeUserId(result.user.id)
      setStoredUserId(result.user.id)

      if (result.needs_passkey) {
        setStep("register")
        return
      }

      setStep("code-verified")
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
        setStep("register")
        return
      }
      if (result.user) {
        storeUserId(result.user.id)
        setStoredUserId(result.user.id)
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
      storeUserId(result.user.id)
      setStoredUserId(result.user.id)
      setUser(result.user)
      await refresh()
      router.replace("/")
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Passkey registration failed")
    } finally {
      setBusy(false)
    }
  }

  function handleForgetDevice() {
    clearStoredUserId()
    setStoredUserId(null)
    setPendingUserId(null)
    setStep("code")
    setError(null)
  }

  const codeLength = config?.enrollment_code_length ?? 8

  return (
    <Card className="w-full border-border/60 bg-card/70 shadow-lg backdrop-blur-sm">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-2xl">
          <ShieldCheckIcon className="size-6 text-primary" />
          Sign in to EchoState
        </CardTitle>
        <CardDescription>
          {step === "passkey"
            ? "Use your passkey to sign in."
            : "Passkey-first access with single-use enrollment codes."}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {error ? (
          <p className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error}
          </p>
        ) : null}

        {step === "passkey" ? (
          <div className="space-y-4">
            <Button
              type="button"
              className="w-full"
              disabled={busy || !storedUserId}
              onClick={() => storedUserId && handlePasskeyLogin(storedUserId)}
            >
              <KeyRoundIcon data-icon="inline-start" />
              Sign in with passkey
            </Button>
            {!storedUserId ? (
              <p className="text-center text-sm text-muted-foreground">
                No saved account on this browser yet. Use an enrollment or recovery code below.
              </p>
            ) : null}
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border/60" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground">or</span>
              </div>
            </div>
            <Button
              type="button"
              variant="outline"
              className="w-full"
              data-testid="login-use-code-button"
              onClick={() => setStep("code")}
            >
              Use enrollment or recovery code
            </Button>
            {storedUserId ? (
              <Button type="button" variant="ghost" className="w-full text-muted-foreground" onClick={handleForgetDevice}>
                Not your account? Clear saved sign-in
              </Button>
            ) : null}
          </div>
        ) : null}

        {step === "code" ? (
          <>
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
                  data-testid="login-code-input"
                />
              </div>
              <Button
                type="submit"
                className="w-full"
                disabled={busy || !code.trim()}
                data-testid="login-verify-code-button"
              >
                {busy ? "Verifying..." : "Verify code"}
              </Button>
            </form>
            {storedUserId ? (
              <Button type="button" variant="ghost" className="w-full" onClick={() => setStep("passkey")}>
                Back to passkey sign-in
              </Button>
            ) : null}
          </>
        ) : null}

        {step === "code-verified" ? (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Code accepted. Sign in with an existing passkey or register this device.
            </p>
            <Button
              type="button"
              className="w-full"
              disabled={busy}
              onClick={() => pendingUserId && handlePasskeyLogin(pendingUserId)}
            >
              <KeyRoundIcon data-icon="inline-start" />
              Sign in with passkey
            </Button>
            <Button
              type="button"
              variant="outline"
              className="w-full"
              disabled={busy}
              data-testid="login-register-passkey-button"
              onClick={() => setStep("register")}
            >
              Register passkey on this device
            </Button>
          </div>
        ) : null}

        {step === "register" ? (
          <form onSubmit={handleRegisterPasskey} className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Register a passkey for this device. You can add more devices later in Profile.
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
            <Button type="submit" className="w-full" disabled={busy} data-testid="login-create-passkey-button">
              <KeyRoundIcon data-icon="inline-start" />
              {busy ? "Registering..." : "Create passkey"}
            </Button>
            {pendingUserId ? (
              <Button type="button" variant="ghost" className="w-full" onClick={() => setStep("code-verified")}>
                Back
              </Button>
            ) : null}
          </form>
        ) : null}
      </CardContent>
    </Card>
  )
}