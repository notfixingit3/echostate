import { redirect } from "next/navigation"

export default function LegacyUsersSettingsPage() {
  redirect("/admin/users")
}