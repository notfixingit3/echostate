"use client"

import * as React from "react"
import { Trash2Icon, PlusIcon } from "lucide-react"

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
import { fetchApi } from "@/lib/api"
import type { Webhook } from "@/lib/settings-types"

export function SettingsIntegrationsPanel() {
  const [webhooks, setWebhooks] = React.useState<Webhook[]>([])
  const [loading, setLoading] = React.useState(true)
  const [name, setName] = React.useState("")
  const [type, setType] = React.useState("slack")
  const [url, setUrl] = React.useState("")
  const [pushover, setPushover] = React.useState<PushoverConfig>(defaultPushoverConfig())

  React.useEffect(() => {
    void loadWebhooks()
  }, [])

  async function loadWebhooks() {
    try {
      const whData = await fetchApi<Webhook[]>("/api/webhooks").catch(() => [])
      setWebhooks(whData || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
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
          ...(pushover.priority === 2 ? { retry: pushover.retry, expire: pushover.expire } : {}),
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
      await fetchApi("/api/webhooks", { method: "POST", body: JSON.stringify(buildWebhookPayload()) })
      setName("")
      setUrl("")
      setPushover(defaultPushoverConfig())
      await loadWebhooks()
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
      await loadWebhooks()
    } catch (e) {
      console.error(e)
    }
  }

  async function deleteWebhook(id: string) {
    if (!confirm("Are you sure you want to delete this webhook?")) return
    try {
      await fetchApi(`/api/webhooks/${id}`, { method: "DELETE" })
      await loadWebhooks()
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
    <div className="space-y-8">
      <Card className="border-border/60 bg-card/60 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="inline-flex items-center gap-1.5">
            Add notification integration
            <HelpTip id="settings.webhooks" />
          </CardTitle>
          <CardDescription>
            Receive alerts when snapshot diffs are detected. Supports Slack, Discord, MS Teams, and Pushover.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={createWebhook} className="flex flex-col gap-6">
            <div className="grid gap-4 sm:grid-cols-[1fr_auto] sm:items-end">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input id="name" placeholder="e.g. SOC Alert Channel" value={name} onChange={(e) => setName(e.target.value)} />
              </div>
              <div className="w-full sm:w-56 space-y-2">
                <Label htmlFor="type">Platform</Label>
                <Select value={type} onValueChange={(val) => val && setType(val)}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
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
                <Input id="url" placeholder="https://..." value={url} onChange={(e) => setUrl(e.target.value)} />
              </div>
            )}
            <Button
              type="submit"
              disabled={!name || (type === "pushover" ? !pushover.app_token || !pushover.user_key : !url)}
              className="w-fit"
            >
              <PlusIcon data-icon="inline-start" />
              Add integration
            </Button>
          </form>
        </CardContent>
      </Card>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold tracking-tight">Active integrations</h2>
        {loading ? (
          <p className="text-muted-foreground">Loading…</p>
        ) : webhooks.length === 0 ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-muted-foreground">
            No integrations configured.
          </div>
        ) : (
          <div className="grid gap-4">
            {webhooks.map((w) => (
              <Card key={w.id} className="border-border/60 bg-card/40">
                <CardContent className="flex items-center justify-between gap-4 p-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold">{w.name}</span>
                      <span className="rounded bg-muted px-2 py-0.5 text-xs font-medium uppercase text-muted-foreground">{w.type}</span>
                    </div>
                    <span className="mt-1 block break-all font-mono text-xs text-muted-foreground">{webhookSummary(w)}</span>
                  </div>
                  <div className="flex shrink-0 items-center gap-4">
                    <div className="flex items-center gap-2">
                      <Switch checked={w.enabled} onCheckedChange={() => toggleWebhook(w)} id={`webhook-${w.id}`} />
                      <Label htmlFor={`webhook-${w.id}`} className="text-xs">{w.enabled ? "Active" : "Disabled"}</Label>
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
    </div>
  )
}