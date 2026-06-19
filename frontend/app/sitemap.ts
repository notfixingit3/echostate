import type { MetadataRoute } from "next"

import { siteBaseUrl } from "@/lib/site-url"

export const dynamic = "force-static"

export default function sitemap(): MetadataRoute.Sitemap {
  const base = siteBaseUrl()
  const lastModified = new Date()

  return [
    {
      url: base,
      lastModified,
      changeFrequency: "weekly",
      priority: 1,
    },
  ]
}