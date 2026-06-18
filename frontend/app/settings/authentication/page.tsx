import { redirect } from "next/navigation"

export default function LegacyAuthenticationSettingsPage() {
  redirect("/admin/authentication")
}