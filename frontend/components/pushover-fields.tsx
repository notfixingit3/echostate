import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export type PushoverConfig = {
  app_token: string
  user_key: string
  priority: number
  sound: string
  device: string
  title: string
  url_title: string
  retry: number
  expire: number
}

export const defaultPushoverConfig = (): PushoverConfig => ({
  app_token: "",
  user_key: "",
  priority: 0,
  sound: "",
  device: "",
  title: "EchoState Alert",
  url_title: "View target",
  retry: 60,
  expire: 1800,
})

const PUSHOVER_SOUNDS = [
  { value: "", label: "User default" },
  { value: "pushover", label: "Pushover (default tone)" },
  { value: "bike", label: "Bike" },
  { value: "bugle", label: "Bugle" },
  { value: "classical", label: "Classical" },
  { value: "cosmic", label: "Cosmic" },
  { value: "siren", label: "Siren" },
  { value: "spacealarm", label: "Space Alarm" },
  { value: "persistent", label: "Persistent (long)" },
  { value: "echo", label: "Pushover Echo (long)" },
  { value: "vibrate", label: "Vibrate only" },
  { value: "none", label: "Silent" },
]

function FieldHelp({ children }: { children: React.ReactNode }) {
  return <p className="text-xs leading-relaxed text-muted-foreground">{children}</p>
}

export function PushoverFields({
  config,
  onChange,
}: {
  config: PushoverConfig
  onChange: (next: PushoverConfig) => void
}) {
  const set = <K extends keyof PushoverConfig>(key: K, value: PushoverConfig[K]) => {
    onChange({ ...config, [key]: value })
  }

  return (
    <div className="grid gap-5 sm:grid-cols-2">
      <div className="space-y-2 sm:col-span-2">
        <Label htmlFor="pushover-app-token">Application API Token</Label>
        <Input
          id="pushover-app-token"
          placeholder="30-character token from pushover.net/apps"
          value={config.app_token}
          onChange={(e) => set("app_token", e.target.value)}
          className="font-mono"
        />
        <FieldHelp>
          Create an application at{" "}
          <a
            href="https://pushover.net/apps/build"
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:underline"
          >
            pushover.net/apps/build
          </a>
          . This token identifies EchoState as the sender.
        </FieldHelp>
      </div>

      <div className="space-y-2 sm:col-span-2">
        <Label htmlFor="pushover-user-key">User or Group Key</Label>
        <Input
          id="pushover-user-key"
          placeholder="30-character key from your Pushover dashboard"
          value={config.user_key}
          onChange={(e) => set("user_key", e.target.value)}
          className="font-mono"
        />
        <FieldHelp>
          Found on your{" "}
          <a
            href="https://pushover.net/dashboard"
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:underline"
          >
            Pushover dashboard
          </a>
          . Use a group key to notify multiple on-call users at once.
        </FieldHelp>
      </div>

      <div className="space-y-2">
        <Label htmlFor="pushover-priority">Priority</Label>
        <Select
          value={String(config.priority)}
          onValueChange={(val) => val && set("priority", Number(val))}
        >
          <SelectTrigger id="pushover-priority">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="-2">Lowest (-2) — badge only, no popup</SelectItem>
            <SelectItem value="-1">Low (-1) — quiet notification, no sound</SelectItem>
            <SelectItem value="0">Normal (0) — default alert behavior</SelectItem>
            <SelectItem value="1">High (1) — bypasses quiet hours</SelectItem>
            <SelectItem value="2">Emergency (2) — repeats until acknowledged</SelectItem>
          </SelectContent>
        </Select>
        <FieldHelp>
          Normal is best for routine recon changes. Use High for important shifts
          (e.g. cert expiry, new open ports). Emergency repeats with sound until
          someone acknowledges — reserve for critical incidents.
        </FieldHelp>
      </div>

      <div className="space-y-2">
        <Label htmlFor="pushover-sound">Sound</Label>
        <Select
          value={config.sound || "__default__"}
          onValueChange={(val) =>
            set("sound", val === "__default__" ? "" : (val ?? ""))
          }
        >
          <SelectTrigger id="pushover-sound">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {PUSHOVER_SOUNDS.map((sound) => (
              <SelectItem
                key={sound.value || "__default__"}
                value={sound.value || "__default__"}
              >
                {sound.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <FieldHelp>
          Overrides the recipient&apos;s default tone for this alert. Leave as
          &quot;User default&quot; unless you want a distinct sound profile for
          EchoState changes.
        </FieldHelp>
      </div>

      <div className="space-y-2">
        <Label htmlFor="pushover-device">Device (optional)</Label>
        <Input
          id="pushover-device"
          placeholder="e.g. iphone, pixel, desktop"
          value={config.device}
          onChange={(e) => set("device", e.target.value)}
        />
        <FieldHelp>
          Send only to a named device from the Pushover client (max 25 chars).
          Leave blank to deliver to all active devices — recommended unless you
          want phone-only or desktop-only alerts.
        </FieldHelp>
      </div>

      <div className="space-y-2">
        <Label htmlFor="pushover-title">Notification title</Label>
        <Input
          id="pushover-title"
          placeholder="EchoState Alert"
          value={config.title}
          onChange={(e) => set("title", e.target.value)}
        />
        <FieldHelp>
          Shown as the push headline. Defaults to &quot;EchoState Alert&quot;;
          the message body still includes the target host and detected changes.
        </FieldHelp>
      </div>

      <div className="space-y-2 sm:col-span-2">
        <Label htmlFor="pushover-url-title">Link button label</Label>
        <Input
          id="pushover-url-title"
          placeholder="View target"
          value={config.url_title}
          onChange={(e) => set("url_title", e.target.value)}
        />
        <FieldHelp>
          Text for the supplementary link EchoState attaches to each alert,
          pointing to the target detail page in the web UI.
        </FieldHelp>
      </div>

      {config.priority === 2 ? (
        <>
          <div className="space-y-2">
            <Label htmlFor="pushover-retry">Retry interval (seconds)</Label>
            <Input
              id="pushover-retry"
              type="number"
              min={30}
              value={config.retry}
              onChange={(e) => set("retry", Number(e.target.value))}
            />
            <FieldHelp>
              Required for Emergency priority. How often Pushover re-sends the
              alert until acknowledged. Minimum 30 seconds.
            </FieldHelp>
          </div>
          <div className="space-y-2">
            <Label htmlFor="pushover-expire">Expire after (seconds)</Label>
            <Input
              id="pushover-expire"
              type="number"
              min={1}
              max={10800}
              value={config.expire}
              onChange={(e) => set("expire", Number(e.target.value))}
            />
            <FieldHelp>
              Required for Emergency priority. Stop retrying after this many
              seconds (max 10,800 / 3 hours). The notification remains visible
              but won&apos;t keep demanding acknowledgement.
            </FieldHelp>
          </div>
        </>
      ) : null}
    </div>
  )
}

export function maskSecret(value: string) {
  if (!value) return "—"
  if (value.length <= 8) return "••••••••"
  return `••••${value.slice(-4)}`
}