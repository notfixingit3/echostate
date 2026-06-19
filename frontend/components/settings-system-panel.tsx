"use client"

import * as React from "react"
import { PlusIcon, SaveIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { HelpTip, LabelWithHelp } from "@/components/help-tip"
import { fetchApi } from "@/lib/api"
import {
  CHANGE_TYPE_HINTS,
  GRAPH_DRIFT_TYPES,
  MAIL_SECURITY_CHANGE_TYPES,
  defaultSystemSettings,
  type SystemSettings,
  type Webhook,
} from "@/lib/settings-types"

export function SettingsSystemPanel() {
  const [webhooks, setWebhooks] = React.useState<Webhook[]>([])
  const [sysSettings, setSysSettings] = React.useState<SystemSettings>(defaultSystemSettings())
  const [saving, setSaving] = React.useState(false)

  React.useEffect(() => {
    void Promise.all([
      fetchApi<Webhook[]>("/api/webhooks").catch(() => []),
      fetchApi<SystemSettings>("/api/settings").catch(() => null),
    ]).then(([wh, sys]) => {
      setWebhooks(wh || [])
      if (sys) setSysSettings(sys)
    })
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
    <form onSubmit={save} className="flex flex-col gap-8">
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
                      <Label htmlFor="dkim_selectors">
                        <LabelWithHelp label="DKIM Selectors" helpId="settings.dkim_selectors" />
                      </Label>
                      <Input
                        id="dkim_selectors"
                        placeholder="e.g. selector3, custom1"
                        value={sysSettings.dkim_selectors || ""}
                        onChange={(e) =>
                          setSysSettings({ ...sysSettings, dkim_selectors: e.target.value })
                        }
                      />
                      <p className="text-xs text-muted-foreground">
                        Comma-separated extra DKIM selectors probed beyond built-in defaults.
                      </p>
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
                  Write-endpoint API key plus passive Shodan, Censys, HIBP, RiskIQ, and VirusTotal correlation.
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
                <div className="space-y-2 sm:col-span-2">
                  <Label htmlFor="virustotal_api_key">
                    <LabelWithHelp label="VirusTotal API key" helpId="settings.virustotal_key" />
                  </Label>
                  <Input
                    id="virustotal_api_key"
                    value={sysSettings.virustotal_api_key || ""}
                    onChange={(e) =>
                      setSysSettings({
                        ...sysSettings,
                        virustotal_api_key: e.target.value,
                      })
                    }
                    placeholder="Optional passive DNS resolutions"
                  />
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
                <div className="flex items-center gap-2 sm:col-span-2">
                  <Switch
                    checked={sysSettings.traceroute_redact_scanner_prefix !== false}
                    onCheckedChange={(checked) =>
                      setSysSettings({
                        ...sysSettings,
                        traceroute_redact_scanner_prefix: checked,
                      })
                    }
                    id="traceroute_redact_scanner_prefix"
                  />
                  <Label htmlFor="traceroute_redact_scanner_prefix">
                    <LabelWithHelp
                      label="Hide scanner network in local traceroutes"
                      helpId="settings.traceroute_redact_scanner_prefix"
                    />
                  </Label>
                </div>
                {sysSettings.traceroute_redact_scanner_prefix !== false ? (
                  <div className="space-y-2">
                    <Label htmlFor="traceroute_redact_local_extra_hops">
                      <LabelWithHelp
                        label="Extra local hops to hide"
                        helpId="settings.traceroute_redact_local_extra_hops"
                      />
                    </Label>
                    <Input
                      id="traceroute_redact_local_extra_hops"
                      type="number"
                      min="0"
                      value={Number(sysSettings.traceroute_redact_local_extra_hops ?? 0)}
                      onChange={(e) =>
                        setSysSettings({
                          ...sysSettings,
                          traceroute_redact_local_extra_hops: Number(e.target.value),
                        })
                      }
                    />
                  </div>
                ) : null}
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
                <CardDescription>Snapshot keep-N per target and audit log retention (0 = unlimited).</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-4 sm:grid-cols-2 max-w-2xl">
                <div className="space-y-2">
                  <Label htmlFor="retention_max_snapshots">
                    <LabelWithHelp label="Max snapshots per target" helpId="settings.retention" />
                  </Label>
                  <Input id="retention_max_snapshots" type="number" min="0" value={sysSettings.retention_max_snapshots || 0} onChange={(e) => setSysSettings({ ...sysSettings, retention_max_snapshots: Number(e.target.value) })} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="audit_retention_days">
                    <LabelWithHelp label="Audit log retention (days)" helpId="settings.audit_retention" />
                  </Label>
                  <Input id="audit_retention_days" type="number" min="0" value={sysSettings.audit_retention_days ?? 90} onChange={(e) => setSysSettings({ ...sysSettings, audit_retention_days: Number(e.target.value) })} />
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
                  <Button type="button" variant="outline" onClick={() => setSysSettings({
                    ...sysSettings,
                    alert_rules: [...(sysSettings.alert_rules || []), {
                      id: `rule-mail-${Date.now()}`,
                      name: "Mail & DNS security",
                      enabled: true,
                      min_severity: "info",
                      match_types: MAIL_SECURITY_CHANGE_TYPES,
                      webhook_ids: [],
                    }],
                  })}>
                    <PlusIcon data-icon="inline-start" />
                    Add mail security preset
                  </Button>
                </div>
              </CardContent>
            </Card>

      <Button type="submit" disabled={saving} className="w-fit">
        <SaveIcon data-icon="inline-start" />
        {saving ? "Saving…" : "Save system configuration"}
      </Button>
    </form>
  )
}
