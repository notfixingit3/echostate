"use client"

import * as React from "react"
import { Trash2Icon, PlusIcon, WebhookIcon, SettingsIcon, SaveIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { HelpTip, LabelWithHelp } from "@/components/help-tip"
import {
  PushoverFields,
  defaultPushoverConfig,
  maskSecret,
  type PushoverConfig,
} from "@/components/pushover-fields"
import { AuthUsersPanel } from "@/components/auth-users"
import { fetchApi } from "@/lib/api"

type Webhook = {
  id: string
  name: string
  type: string
  url: string
  config?: Record<string, unknown>
  enabled: boolean
}

const GRAPH_DRIFT_TYPES = [
  "graph_bgp_origin_added",
  "graph_bgp_origin_removed",
  "graph_as_path_added",
  "graph_as_path_removed",
  "graph_rpki_changed",
  "graph_traceroute_hop_added",
  "graph_traceroute_hop_removed",
  "graph_traceroute_reordered",
]

const CHANGE_TYPE_HINTS =
  "cert_expiry, bgp_origin, bgp_hijack_risk, new_ct_subdomain, dmarc_policy, web_title, graph_bgp_origin_added, graph_traceroute_hop_removed"

type AlertRule = {
  id: string
  name: string
  enabled: boolean
  min_severity: string
  match_types: string[]
  webhook_ids: string[]
}

type SystemSettings = {
  dns_servers: string
  pwhois_server: string
  rate_limit: number
  api_key?: string
  shodan_api_key?: string
  hibp_api_key?: string
  riskiq_api_user?: string
  riskiq_api_key?: string
  censys_api_id?: string
  censys_api_secret?: string
  scan_concurrency?: number
  scan_job_timeout_sec?: number
  default_gatherer_timeout_sec?: number
  ct_http_timeout_sec?: number
  traceroute_timeout_sec?: number
  screenshot_timeout_sec?: number
  retention_max_snapshots?: number
  schedule_enabled?: boolean
  schedule_interval_minutes?: number
  schedule_stale_hours?: number
  schedule_tags?: string[]
  alert_rules?: AlertRule[]
  auth_enabled?: boolean
  enrollment_code_ttl_hours?: number
  enrollment_code_length?: number
  recovery_code_length?: number
  session_ttl_hours?: number
  max_code_attempts?: number
  code_attempt_window_minutes?: number
  webauthn_rp_id?: string
  webauthn_rp_origin?: string
}

export default function SettingsPage() {
  const [webhooks, setWebhooks] = React.useState<Webhook[]>([])
  const [loading, setLoading] = React.useState(true)

  const [name, setName] = React.useState("")
  const [type, setType] = React.useState("slack")
  const [url, setUrl] = React.useState("")
  const [pushover, setPushover] = React.useState<PushoverConfig>(defaultPushoverConfig())

  const [sysSettings, setSysSettings] = React.useState<SystemSettings>({
    dns_servers: "8.8.8.8,1.1.1.1",
    pwhois_server: "whois.pwhois.org",
    rate_limit: 30,
    scan_concurrency: 2,
    scan_job_timeout_sec: 120,
    default_gatherer_timeout_sec: 20,
    ct_http_timeout_sec: 60,
    traceroute_timeout_sec: 40,
    screenshot_timeout_sec: 25,
    retention_max_snapshots: 0,
    schedule_enabled: false,
    schedule_interval_minutes: 60,
    schedule_stale_hours: 24,
    schedule_tags: [],
    alert_rules: [],
    auth_enabled: true,
    enrollment_code_ttl_hours: 24,
    enrollment_code_length: 8,
    recovery_code_length: 12,
    session_ttl_hours: 168,
    max_code_attempts: 5,
    code_attempt_window_minutes: 15,
    webauthn_rp_id: "",
    webauthn_rp_origin: "",
  })
  const [savingSys, setSavingSys] = React.useState(false)

  React.useEffect(() => {
    loadWebhooks()
  }, [])

  async function loadWebhooks() {
    try {
      const [whData, sysData] = await Promise.all([
        fetchApi<Webhook[]>("/api/webhooks").catch(() => []),
        fetchApi<SystemSettings>("/api/settings").catch(() => null),
      ])
      setWebhooks(whData || [])
      if (sysData) setSysSettings(sysData)
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  async function saveSysSettings(e: React.FormEvent) {
    e.preventDefault()
    setSavingSys(true)
    try {
      await fetchApi("/api/settings", {
        method: "PUT",
        body: JSON.stringify(sysSettings),
      })
      const refreshed = await fetchApi<SystemSettings>("/api/settings")
      setSysSettings(refreshed)
    } catch (e) {
      console.error(e)
    } finally {
      setSavingSys(false)
    }
  }

  function buildWebhookPayload() {
    if (type === "pushover") {
      return {
        name,
        type,
        url: "https://api.pushover.net/1/messages.json",
        enabled: true,
        config: {
          app_token: pushover.app_token.trim(),
          user_key: pushover.user_key.trim(),
          priority: pushover.priority,
          sound: pushover.sound.trim(),
          device: pushover.device.trim(),
          title: pushover.title.trim() || "EchoState Alert",
          url_title: pushover.url_title.trim() || "View target",
          ...(pushover.priority === 2
            ? { retry: pushover.retry, expire: pushover.expire }
            : {}),
        },
      }
    }

    return { name, type, url, enabled: true, config: {} }
  }

  async function createWebhook(e: React.FormEvent) {
    e.preventDefault()
    if (!name) return
    if (type !== "pushover" && !url) return
    if (type === "pushover" && (!pushover.app_token || !pushover.user_key)) return

    try {
      await fetchApi("/api/webhooks", {
        method: "POST",
        body: JSON.stringify(buildWebhookPayload()),
      })
      setName("")
      setUrl("")
      setPushover(defaultPushoverConfig())
      loadWebhooks()
    } catch (e) {
      console.error(e)
    }
  }

  async function toggleWebhook(w: Webhook) {
    try {
      await fetchApi(`/api/webhooks/${w.id}`, {
        method: "PUT",
        body: JSON.stringify({ ...w, enabled: !w.enabled }),
      })
      loadWebhooks()
    } catch (e) {
      console.error(e)
    }
  }

  async function deleteWebhook(id: string) {
    if (!confirm("Are you sure you want to delete this webhook?")) return
    try {
      await fetchApi(`/api/webhooks/${id}`, { method: "DELETE" })
      loadWebhooks()
    } catch (e) {
      console.error(e)
    }
  }

  function webhookSummary(w: Webhook) {
    if (w.type === "pushover") {
      const cfg = w.config || {}
      const priority = typeof cfg.priority === "number" ? cfg.priority : 0
      const sound = typeof cfg.sound === "string" && cfg.sound ? cfg.sound : "default"
      return `App ${maskSecret(String(cfg.app_token || ""))} · User ${maskSecret(String(cfg.user_key || ""))} · Priority ${priority} · Sound ${sound}`
    }
    return w.url
  }

  return (
    <div className="container py-10 max-w-4xl">
      <div className="mb-8 flex items-center gap-2">
        <WebhookIcon className="size-6 text-primary" />
        <h1 className="text-3xl font-bold tracking-tight">Settings & Integrations</h1>
        <HelpTip id="page.settings" />
      </div>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm mb-8">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Add Notification Integration
            <HelpTip id="settings.webhooks" />
          </CardTitle>
          <CardDescription>
            Receive alerts when snapshot diffs are detected. Supports Slack, Discord,
            MS Teams webhooks, and Pushover mobile/desktop notifications.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={createWebhook} className="flex flex-col gap-6">
            <div className="grid gap-4 sm:grid-cols-[1fr_auto] sm:items-end">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  placeholder="e.g. SOC Alert Channel"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div className="w-full sm:w-56 space-y-2">
                <Label htmlFor="type">Platform</Label>
                <Select value={type} onValueChange={(val) => val && setType(val)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="slack">Slack</SelectItem>
                    <SelectItem value="discord">Discord</SelectItem>
                    <SelectItem value="teams">MS Teams</SelectItem>
                    <SelectItem value="pushover">Pushover</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {type === "pushover" ? (
              <PushoverFields config={pushover} onChange={setPushover} />
            ) : (
              <div className="space-y-2">
                <Label htmlFor="url">
                  <LabelWithHelp label="Webhook URL" helpId="settings.webhook_url" />
                </Label>
                <Input
                  id="url"
                  placeholder="https://..."
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  Paste the incoming webhook URL from your chat platform. EchoState
                  posts a summary of detected field changes with a link to the target.
                </p>
              </div>
            )}

            <Button
              type="submit"
              disabled={
                !name ||
                (type === "pushover"
                  ? !pushover.app_token || !pushover.user_key
                  : !url)
              }
              className="w-fit"
            >
              <PlusIcon data-icon="inline-start" />
              Add integration
            </Button>
          </form>
        </CardContent>
      </Card>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold tracking-tight">Active Integrations</h2>
        {loading ? (
          <p className="text-muted-foreground">Loading...</p>
        ) : webhooks.length === 0 ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-muted-foreground">
            No integrations configured. Add one above to start receiving alerts.
          </div>
        ) : (
          <div className="grid gap-4">
            {webhooks.map((w) => (
              <Card key={w.id} className="border-border/60 bg-card/40">
                <CardContent className="flex items-center justify-between gap-4 p-4">
                  <div className="min-w-0 flex-1 flex-col">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold">{w.name}</span>
                      <span className="rounded bg-muted px-2 py-0.5 text-xs font-medium uppercase text-muted-foreground">
                        {w.type}
                      </span>
                    </div>
                    <span className="mt-1 block break-all font-mono text-xs text-muted-foreground">
                      {webhookSummary(w)}
                    </span>
                  </div>
                  <div className="flex shrink-0 items-center gap-4">
                    <div className="flex items-center gap-2">
                      <Switch
                         checked={w.enabled}
                         onCheckedChange={() => toggleWebhook(w)}
                         id={`webhook-${w.id}`}
                      />
                      <Label htmlFor={`webhook-${w.id}`} className="text-xs">
                        {w.enabled ? "Active" : "Disabled"}
                      </Label>
                    </div>
                    <Button variant="ghost" size="icon" onClick={() => deleteWebhook(w.id)} className="text-destructive hover:bg-destructive/10 hover:text-destructive">
                      <Trash2Icon />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>

      <div className="mb-8 mt-16 flex items-center gap-2">
        <SettingsIcon className="size-6 text-primary" />
        <h2 className="text-2xl font-bold tracking-tight">System Configuration</h2>
        <HelpTip id="page.settings" />
      </div>

      <form onSubmit={saveSysSettings} className="flex flex-col gap-8">
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

      <AuthUsersPanel />

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle>Global Settings</CardTitle>
          <CardDescription>
            Configure DNS resolution, WHOIS enrichment options, and API limits.
          </CardDescription>
        </CardHeader>
        <CardContent>
            <div className="grid gap-6 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="dns_servers">
                  <LabelWithHelp label="DNS Resolvers" helpId="settings.dns_servers" />
                </Label>
                <Input
                  id="dns_servers"
                  placeholder="e.g. 8.8.8.8, 1.1.1.1"
                  value={sysSettings.dns_servers}
                  onChange={(e) => setSysSettings({ ...sysSettings, dns_servers: e.target.value })}
                />
                <p className="text-xs text-muted-foreground">Comma-separated IPs. Leave blank for system default.</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="pwhois_server">
                  <LabelWithHelp label="pWhois Address" helpId="settings.pwhois_server" />
                </Label>
                <Input
                  id="pwhois_server"
                  placeholder="whois.pwhois.org"
                  value={sysSettings.pwhois_server}
                  onChange={(e) => setSysSettings({ ...sysSettings, pwhois_server: e.target.value })}
                />
                <p className="text-xs text-muted-foreground">Used for IP ASN/org enrichment.</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="rate_limit">
                  <LabelWithHelp label="API Rate Limit" helpId="settings.rate_limit" />
                </Label>
                <Input
                  id="rate_limit"
                  type="number"
                  min="1"
                  value={sysSettings.rate_limit}
                  onChange={(e) => setSysSettings({ ...sysSettings, rate_limit: Number(e.target.value) })}
                />
                <p className="text-xs text-muted-foreground">Requests allowed per minute.</p>
              </div>
            </div>
        </CardContent>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            API Access & Enrichment Keys
            <HelpTip id="settings.api_key" />
          </CardTitle>
          <CardDescription>
            Write-endpoint API key plus passive Shodan, Censys, HIBP, and RiskIQ (PassiveTotal) correlation.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="api_key">
              <LabelWithHelp label="EchoState API key" helpId="settings.api_key" />
            </Label>
            <Input id="api_key" value={sysSettings.api_key || ""} onChange={(e) => setSysSettings({ ...sysSettings, api_key: e.target.value })} placeholder="Required for POST/PUT/DELETE when set" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="shodan_api_key">
              <LabelWithHelp label="Shodan API key" helpId="settings.shodan_key" />
            </Label>
            <Input id="shodan_api_key" value={sysSettings.shodan_api_key || ""} onChange={(e) => setSysSettings({ ...sysSettings, shodan_api_key: e.target.value })} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="censys_api_id">
              <LabelWithHelp label="Censys API ID" helpId="settings.censys_keys" />
            </Label>
            <Input id="censys_api_id" value={sysSettings.censys_api_id || ""} onChange={(e) => setSysSettings({ ...sysSettings, censys_api_id: e.target.value })} />
          </div>
          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="censys_api_secret">Censys API secret</Label>
            <Input id="censys_api_secret" value={sysSettings.censys_api_secret || ""} onChange={(e) => setSysSettings({ ...sysSettings, censys_api_secret: e.target.value })} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="hibp_api_key">
              <LabelWithHelp label="HIBP API key" helpId="settings.hibp_key" />
            </Label>
            <Input id="hibp_api_key" value={sysSettings.hibp_api_key || ""} onChange={(e) => setSysSettings({ ...sysSettings, hibp_api_key: e.target.value })} placeholder="Breach checks for WHOIS emails" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="riskiq_api_user">
              <LabelWithHelp label="RiskIQ / PassiveTotal user" helpId="settings.riskiq_keys" />
            </Label>
            <Input id="riskiq_api_user" value={sysSettings.riskiq_api_user || ""} onChange={(e) => setSysSettings({ ...sysSettings, riskiq_api_user: e.target.value })} />
          </div>
          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="riskiq_api_key">RiskIQ / PassiveTotal API key</Label>
            <Input id="riskiq_api_key" value={sysSettings.riskiq_api_key || ""} onChange={(e) => setSysSettings({ ...sysSettings, riskiq_api_key: e.target.value })} placeholder="Passive DNS for scanned host" />
          </div>
        </CardContent>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Scan Performance
            <HelpTip id="settings.scan_concurrency" />
          </CardTitle>
          <CardDescription>
            Worker concurrency and per-gatherer timeouts (seconds).
            <HelpTip id="settings.scan_timeouts" className="ml-1.5" />
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          {[
            ["scan_concurrency", "Scan concurrency"],
            ["scan_job_timeout_sec", "Scan job timeout"],
            ["default_gatherer_timeout_sec", "Default gatherer timeout"],
            ["ct_http_timeout_sec", "CT HTTP timeout"],
            ["traceroute_timeout_sec", "Traceroute timeout"],
            ["screenshot_timeout_sec", "Screenshot timeout"],
          ].map(([key, label]) => (
            <div className="space-y-2" key={key}>
              <Label htmlFor={key}>
                <LabelWithHelp
                  label={label}
                  helpId={key === "scan_concurrency" ? "settings.scan_concurrency" : "settings.scan_timeouts"}
                />
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
        </CardContent>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Scheduled Rescans
            <HelpTip id="settings.scheduler" />
          </CardTitle>
          <CardDescription>Enqueue stale targets automatically (tag filter optional).</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="flex items-center gap-2 sm:col-span-2">
            <Switch
              checked={!!sysSettings.schedule_enabled}
              onCheckedChange={(checked) => setSysSettings({ ...sysSettings, schedule_enabled: checked })}
              id="schedule_enabled"
            />
            <Label htmlFor="schedule_enabled">Enable scheduler</Label>
          </div>
          <div className="space-y-2">
            <Label htmlFor="schedule_interval_minutes">Interval (minutes)</Label>
            <Input id="schedule_interval_minutes" type="number" min="5" value={sysSettings.schedule_interval_minutes || 60} onChange={(e) => setSysSettings({ ...sysSettings, schedule_interval_minutes: Number(e.target.value) })} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="schedule_stale_hours">Stale after (hours)</Label>
            <Input id="schedule_stale_hours" type="number" min="1" value={sysSettings.schedule_stale_hours || 24} onChange={(e) => setSysSettings({ ...sysSettings, schedule_stale_hours: Number(e.target.value) })} />
          </div>
          <div className="space-y-2 sm:col-span-2">
            <Label htmlFor="schedule_tags">
              <LabelWithHelp label="Limit to tags (comma-separated)" helpId="settings.schedule_tags" />
            </Label>
            <Input id="schedule_tags" value={(sysSettings.schedule_tags || []).join(", ")} onChange={(e) => setSysSettings({ ...sysSettings, schedule_tags: e.target.value.split(",").map((t) => t.trim()).filter(Boolean) })} />
          </div>
        </CardContent>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Retention
            <HelpTip id="settings.retention" />
          </CardTitle>
          <CardDescription>Keep only the newest N snapshots per target (0 = unlimited).</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 max-w-xs">
            <Label htmlFor="retention_max_snapshots">
              <LabelWithHelp label="Max snapshots per target" helpId="settings.retention" />
            </Label>
            <Input id="retention_max_snapshots" type="number" min="0" value={sysSettings.retention_max_snapshots || 0} onChange={(e) => setSysSettings({ ...sysSettings, retention_max_snapshots: Number(e.target.value) })} />
          </div>
        </CardContent>
      </Card>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Alert Rules
            <HelpTip id="settings.alert_rules" />
          </CardTitle>
          <CardDescription>
            Filter webhook notifications by severity, change type, and destination integration.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {(sysSettings.alert_rules || []).map((rule, index) => (
            <div key={rule.id || index} className="grid gap-3 rounded-lg border p-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label>Name</Label>
                <Input value={rule.name} onChange={(e) => {
                  const rules = [...(sysSettings.alert_rules || [])]
                  rules[index] = { ...rule, name: e.target.value }
                  setSysSettings({ ...sysSettings, alert_rules: rules })
                }} />
              </div>
              <div className="space-y-2">
                <Label>Min severity</Label>
                <Select value={rule.min_severity} onValueChange={(val) => val && (() => {
                  const rules = [...(sysSettings.alert_rules || [])]
                  rules[index] = { ...rule, min_severity: val }
                  setSysSettings({ ...sysSettings, alert_rules: rules })
                })()}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="info">info</SelectItem>
                    <SelectItem value="warning">warning</SelectItem>
                    <SelectItem value="critical">critical</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2 sm:col-span-2">
                <Label>
                  <LabelWithHelp
                    label="Match types (comma-separated, empty = all)"
                    helpId="settings.alert_match_types"
                  />
                </Label>
                <Input value={(rule.match_types || []).join(", ")} onChange={(e) => {
                  const rules = [...(sysSettings.alert_rules || [])]
                  rules[index] = { ...rule, match_types: e.target.value.split(",").map((t) => t.trim()).filter(Boolean) }
                  setSysSettings({ ...sysSettings, alert_rules: rules })
                }} />
                <p className="text-xs text-muted-foreground">{CHANGE_TYPE_HINTS}</p>
              </div>
              {webhooks.length > 0 ? (
                <div className="space-y-2 sm:col-span-2">
                  <Label>
                    <LabelWithHelp
                      label="Deliver to integrations (empty = all enabled)"
                      helpId="settings.alert_webhooks"
                    />
                  </Label>
                  <div className="flex flex-wrap gap-3">
                    {webhooks.map((wh) => {
                      const selected = (rule.webhook_ids || []).includes(wh.id)
                      return (
                        <label key={wh.id} className="flex items-center gap-2 text-sm">
                          <input
                            type="checkbox"
                            checked={selected}
                            onChange={() => {
                              const rules = [...(sysSettings.alert_rules || [])]
                              const ids = new Set(rule.webhook_ids || [])
                              if (selected) ids.delete(wh.id)
                              else ids.add(wh.id)
                              rules[index] = { ...rule, webhook_ids: Array.from(ids) }
                              setSysSettings({ ...sysSettings, alert_rules: rules })
                            }}
                          />
                          {wh.name}
                        </label>
                      )
                    })}
                  </div>
                </div>
              ) : null}
              <div className="flex items-center gap-2 sm:col-span-2">
                <Switch checked={rule.enabled} onCheckedChange={(checked) => {
                  const rules = [...(sysSettings.alert_rules || [])]
                  rules[index] = { ...rule, enabled: checked }
                  setSysSettings({ ...sysSettings, alert_rules: rules })
                }} id={`rule-${index}`} />
                <Label htmlFor={`rule-${index}`}>Enabled</Label>
              </div>
            </div>
          ))}
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={() => setSysSettings({
              ...sysSettings,
              alert_rules: [...(sysSettings.alert_rules || []), {
                id: `rule-${Date.now()}`,
                name: "New rule",
                enabled: true,
                min_severity: "warning",
                match_types: [],
                webhook_ids: [],
              }],
            })}>
              <PlusIcon data-icon="inline-start" />
              Add alert rule
            </Button>
            <Button type="button" variant="outline" onClick={() => setSysSettings({
              ...sysSettings,
              alert_rules: [...(sysSettings.alert_rules || []), {
                id: `rule-graph-${Date.now()}`,
                name: "Graph path drift",
                enabled: true,
                min_severity: "warning",
                match_types: GRAPH_DRIFT_TYPES,
                webhook_ids: [],
              }],
            })}>
              <PlusIcon data-icon="inline-start" />
              Add graph drift preset
            </Button>
          </div>
        </CardContent>
      </Card>

      <Button type="submit" disabled={savingSys} className="w-fit">
        <SaveIcon data-icon="inline-start" />
        {savingSys ? "Saving..." : "Save system configuration"}
      </Button>
      </form>

    </div>
  )
}