import type { Metadata, Viewport } from "next"
import { Plus_Jakarta_Sans, JetBrains_Mono, Sora } from "next/font/google"

import "./globals.css"
import { AppShell } from "@/components/app-shell"
import { AuthProvider } from "@/components/auth-provider"
import { PwaRegister } from "@/components/pwa-register"
import { ThemeProvider } from "@/components/theme-provider"
import { UserThemeSync } from "@/components/user-theme-sync"
import { TooltipProvider } from "@/components/ui/tooltip"
import {
  PWA_BACKGROUND_COLOR,
  PWA_THEME_COLOR,
  SITE_DESCRIPTION,
  SITE_NAME,
} from "@/lib/site"
import { cn } from "@/lib/utils"

const sans = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-sans",
})

const heading = Sora({
  subsets: ["latin"],
  variable: "--font-heading",
  weight: ["400", "500", "600", "700"],
})

const mono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
})

export const metadata: Metadata = {
  title: {
    default: SITE_NAME,
    template: `%s · ${SITE_NAME}`,
  },
  description: SITE_DESCRIPTION,
  applicationName: SITE_NAME,
  manifest: "/manifest.webmanifest",
  appleWebApp: {
    capable: true,
    title: SITE_NAME,
    statusBarStyle: "default",
  },
  formatDetection: {
    telephone: false,
  },
  icons: {
    icon: [
      { url: "/icon.png", sizes: "32x32", type: "image/png" },
      { url: "/icon.svg", type: "image/svg+xml" },
    ],
    apple: [{ url: "/apple-icon.png", sizes: "180x180", type: "image/png" }],
  },
}

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: PWA_THEME_COLOR },
    { media: "(prefers-color-scheme: dark)", color: "#0f766e" },
  ],
  colorScheme: "light dark",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={cn(
        "antialiased",
        sans.variable,
        heading.variable,
        mono.variable,
        "font-sans"
      )}
    >
      <body className="min-h-svh overscroll-y-none">
        <PwaRegister />
        <ThemeProvider>
          <AuthProvider>
            <UserThemeSync />
            <TooltipProvider>
              <AppShell>{children}</AppShell>
            </TooltipProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}