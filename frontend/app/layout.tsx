import { Plus_Jakarta_Sans, JetBrains_Mono, Sora } from "next/font/google"

import "./globals.css"
import { AuthProvider } from "@/components/auth-provider"
import { ThemeProvider } from "@/components/theme-provider"
import { TooltipProvider } from "@/components/ui/tooltip"
import { Nav } from "@/components/nav"
import { Footer } from "@/components/footer"
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
      <body>
        <ThemeProvider>
          <AuthProvider>
            <TooltipProvider>
              <div className="flex min-h-svh flex-col">
                <Nav />
                <main className="flex-1">{children}</main>
                <Footer />
              </div>
            </TooltipProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}