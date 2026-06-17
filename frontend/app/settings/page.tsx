"use client"

import * as React from "react"
import { Trash2Icon, PlusIcon, WebhookIcon, SettingsIcon, SaveIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { fetchApi } from "@/lib/api"

type Webhook = {
  id: string
  name: string
  type: string
  url: string
  enabled: boolean
}

type SystemSettings = {
  dns_servers: string
  pwhois_server: string
  rate_limit: number
}

export default function SettingsPage() {
  const [webhooks, setWebhooks] = React.useState<Webhook[]>([])
  const [loading, setLoading] = React.useState(true)

  const [name, setName] = React.useState("")
  const [type, setType] = React.useState("slack")
  const [url, setUrl] = React.useState("")

  const [sysSettings, setSysSettings] = React.useState<SystemSettings>({
    dns_servers: "8.8.8.8,1.1.1.1",
    pwhois_server: "whois.pwhois.org",
    rate_limit: 30,
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
        body: JSON.stringify({
          dns_servers: sysSettings.dns_servers,
          pwhois_server: sysSettings.pwhois_server,
          rate_limit: Number(sysSettings.rate_limit)
        }),
      })
    } catch (e) {
      console.error(e)
    } finally {
      setSavingSys(false)
    }
  }

  async function createWebhook(e: React.FormEvent) {
    e.preventDefault()
    if (!name || !url) return
    try {
      await fetchApi("/api/webhooks", {
        method: "POST",
        body: JSON.stringify({ name, type, url, enabled: true }),
      })
      setName("")
      setUrl("")
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

  return (
    <div className="container py-10 max-w-4xl">
      <div className="mb-8 flex items-center gap-2">
        <WebhookIcon className="size-6 text-primary" />
        <h1 className="text-3xl font-bold tracking-tight">Settings & Integrations</h1>
      </div>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm mb-8">
        <CardHeader>
          <CardTitle>Add New Webhook</CardTitle>
          <CardDescription>
            Configure Slack, Discord, or MS Teams webhooks to receive alerts when target changes are detected.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={createWebhook} className="flex flex-col gap-4 sm:flex-row sm:items-end">
            <div className="flex-1 space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="e.g. SOC Alert Channel"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div className="w-full sm:w-48 space-y-2">
              <Label htmlFor="type">Platform</Label>
              <Select value={type} onValueChange={(val) => val && setType(val)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="slack">Slack</SelectItem>
                  <SelectItem value="discord">Discord</SelectItem>
                  <SelectItem value="teams">MS Teams</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex-1 space-y-2">
              <Label htmlFor="url">Webhook URL</Label>
              <Input
                id="url"
                placeholder="https://..."
                value={url}
                onChange={(e) => setUrl(e.target.value)}
              />
            </div>
            <Button type="submit" disabled={!name || !url}>
              <PlusIcon data-icon="inline-start" />
              Add
            </Button>
          </form>
        </CardContent>
      </Card>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold tracking-tight">Active Webhooks</h2>
        {loading ? (
          <p className="text-muted-foreground">Loading...</p>
        ) : webhooks.length === 0 ? (
          <div className="rounded-lg border border-dashed p-8 text-center text-muted-foreground">
            No webhooks configured. Add one above to start receiving alerts.
          </div>
        ) : (
          <div className="grid gap-4">
            {webhooks.map((w) => (
              <Card key={w.id} className="border-border/60 bg-card/40">
                <CardContent className="flex items-center justify-between p-4">
                  <div className="flex flex-col">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold">{w.name}</span>
                      <span className="rounded bg-muted px-2 py-0.5 text-xs font-medium uppercase text-muted-foreground">
                        {w.type}
                      </span>
                    </div>
                    <span className="font-mono text-xs text-muted-foreground mt-1 truncate max-w-md">
                      {w.url}
                    </span>
                  </div>
                  <div className="flex items-center gap-4">
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

      <div className="mb-8 flex items-center gap-2 mt-16">
        <SettingsIcon className="size-6 text-primary" />
        <h2 className="text-2xl font-bold tracking-tight">System Configuration</h2>
      </div>

      <Card className="border-border/60 bg-card/60 backdrop-blur-sm mb-8">
        <CardHeader>
          <CardTitle>Global Settings</CardTitle>
          <CardDescription>
            Configure DNS resolution, WHOIS enrichment options, and API limits.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={saveSysSettings} className="flex flex-col gap-6">
            <div className="grid gap-6 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="dns_servers">DNS Resolvers</Label>
                <Input
                  id="dns_servers"
                  placeholder="e.g. 8.8.8.8, 1.1.1.1"
                  value={sysSettings.dns_servers}
                  onChange={(e) => setSysSettings({ ...sysSettings, dns_servers: e.target.value })}
                />
                <p className="text-xs text-muted-foreground">Comma-separated IPs. Leave blank for system default.</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="pwhois_server">pWhois Address</Label>
                <Input
                  id="pwhois_server"
                  placeholder="whois.pwhois.org"
                  value={sysSettings.pwhois_server}
                  onChange={(e) => setSysSettings({ ...sysSettings, pwhois_server: e.target.value })}
                />
                <p className="text-xs text-muted-foreground">Used for IP ASN/org enrichment.</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="rate_limit">API Rate Limit</Label>
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
            <Button type="submit" disabled={savingSys} className="w-fit">
              <SaveIcon data-icon="inline-start" />
              {savingSys ? "Saving..." : "Save Configuration"}
            </Button>
          </form>
        </CardContent>
      </Card>

    </div>
  )
}
