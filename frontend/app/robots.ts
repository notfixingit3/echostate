import type { MetadataRoute } from "next"

import { siteBaseUrl } from "@/lib/site-url"

export const dynamic = "force-static"

export default function robots(): MetadataRoute.Robots {
  const base = siteBaseUrl()

  return {
    rules: {
      userAgent: "*",
      allow: "/",
      disallow: ["/admin/", "/api/", "/login", "/profile", "/settings"],
    },
    sitemap: `${base}/sitemap.xml`,
  }
}