import { redirect } from "next/navigation"

export default function LegacyIntegrationsSettingsPage() {
  redirect("/admin/integrations")
}