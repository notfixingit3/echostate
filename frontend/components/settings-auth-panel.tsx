"use client"

import * as React from "react"
import { SaveIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { HelpTip, LabelWithHelp } from "@/components/help-tip"
import { fetchApi } from "@/lib/api"
import { defaultSystemSettings, type SystemSettings } from "@/lib/settings-types"

export function SettingsAuthPanel() {
  const [sysSettings, setSysSettings] = React.useState<SystemSettings>(defaultSystemSettings())
  const [saving, setSaving] = React.useState(false)

  React.useEffect(() => {
    void fetchApi<SystemSettings>("/api/settings").then((data) => setSysSettings(data)).catch(console.error)
  }, [])

  async function save(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    try {
      await fetchApi("/api/settings", { method: "PUT", body: JSON.stringify(sysSettings) })
      setSysSettings(await fetchApi<SystemSettings>("/api/settings"))
    } catch (err) {
      console.error(err)
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={save} className="flex flex-col gap-6">
      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
              <CardHeader>
                <CardTitle className="inline-flex items-center gap-1.5">
                  Authentication
                  <HelpTip id="settings.auth" />
                </CardTitle>
                <CardDescription>
                  Passkey enrollment defaults, session lifetime, and WebAuthn relying party settings.
                </CardDescription>
              </CardHeader>
              <CardContent className="grid gap-4 sm:grid-cols-2">
                <div className="flex items-center gap-2 sm:col-span-2">
                  <Switch
                    checked={!!sysSettings.auth_enabled}
                    onCheckedChange={(checked) => setSysSettings({ ...sysSettings, auth_enabled: checked })}
                    id="auth_enabled"
                  />
                  <Label htmlFor="auth_enabled">
                    <LabelWithHelp label="Require authentication when users exist" helpId="settings.auth_enabled" />
                  </Label>
                </div>
                {[
                  ["enrollment_code_ttl_hours", "Enrollment code TTL (hours)", "settings.enrollment_code_ttl"],
                  ["enrollment_code_length", "Enrollment code length (digits)", "settings.enrollment_code_length"],
                  ["recovery_code_length", "Recovery code length", "settings.recovery_code_length"],
                  ["session_ttl_hours", "Session TTL (hours)", "settings.session_ttl"],
                  ["max_code_attempts", "Max code attempts", "settings.max_code_attempts"],
                  ["code_attempt_window_minutes", "Code attempt window (minutes)", "settings.code_attempt_window"],
                ].map(([key, label, helpId]) => (
                  <div className="space-y-2" key={key}>
                    <Label htmlFor={key}>
                      <LabelWithHelp label={label} helpId={helpId} />
                    </Label>
                    <Input
                      id={key}
                      type="number"
                      min="1"
                      value={Number(sysSettings[key as keyof SystemSettings] ?? 0)}
                      onChange={(e) =>
                        setSysSettings({
                          ...sysSettings,
                          [key]: Number(e.target.value),
                        })
                      }
                    />
                  </div>
                ))}
                <div className="space-y-2">
                  <Label htmlFor="webauthn_rp_id">
                    <LabelWithHelp label="WebAuthn RP ID" helpId="settings.webauthn_rp_id" />
                  </Label>
                  <Input
                    id="webauthn_rp_id"
                    value={sysSettings.webauthn_rp_id || ""}
                    onChange={(e) => setSysSettings({ ...sysSettings, webauthn_rp_id: e.target.value })}
                    placeholder="localhost or app.example.com"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="webauthn_rp_origin">
                    <LabelWithHelp label="WebAuthn RP origin" helpId="settings.webauthn_rp_origin" />
                  </Label>
                  <Input
                    id="webauthn_rp_origin"
                    value={sysSettings.webauthn_rp_origin || ""}
                    onChange={(e) => setSysSettings({ ...sysSettings, webauthn_rp_origin: e.target.value })}
                    placeholder="http://localhost:3001"
                  />
                </div>
              </CardContent>
            </Card>

      <Button type="submit" disabled={saving} className="w-fit">
        <SaveIcon data-icon="inline-start" />
        {saving ? "Saving…" : "Save authentication settings"}
      </Button>
    </form>
  )
}
