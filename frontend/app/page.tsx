import { Suspense } from "react"

import { HelpTip } from "@/components/help-tip"
import { ScanForm } from "@/components/scan-form"
import { Badge } from "@/components/ui/badge"
import { GlobeIcon, NetworkIcon, MonitorIcon } from "lucide-react"

const features = [
  {
    icon: GlobeIcon,
    title: "WHOIS",
    helpId: "feature.whois",
    description: "Registrar, nameservers, expiration, and raw registry data.",
  },
  {
    icon: NetworkIcon,
    title: "ASN / BGP",
    helpId: "feature.asn_bgp",
    description: "Origin ASN, prefix, country, and allocation via Team Cymru.",
  },
  {
    icon: MonitorIcon,
    title: "Web intel",
    helpId: "feature.web_intel",
    description: "Page title, final URL, and copyright signals from the landing page.",
  },
] as const

export default function Page() {
  return (
    <section className="container flex flex-1 flex-col py-8 md:py-10">
      <div className="grid w-full items-start gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)] lg:gap-10 xl:gap-12">
        <div className="flex flex-col gap-5 lg:gap-6 lg:pt-1">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="secondary" className="font-mono text-xs">
              Passive recon
            </Badge>
            <Badge variant="outline" className="font-mono text-xs">
              WHOIS · ASN · Web
            </Badge>
          </div>
          <div className="flex flex-col gap-3">
            <h1 className="font-heading text-3xl font-semibold tracking-tight text-balance sm:text-4xl lg:text-[2.75rem] lg:leading-tight">
              Know your target before you touch it.
            </h1>
            <p className="max-w-xl text-base text-muted-foreground text-balance md:text-lg">
              Feed EchoState a host, IP, or URL and gather WHOIS, BGP/ASN, and
              webpage intelligence in seconds — with full snapshot history.
            </p>
          </div>
          <ul className="flex flex-col gap-3 border-t border-border/50 pt-5">
            {features.map((feature) => {
              const Icon = feature.icon
              return (
                <li key={feature.title} className="flex items-start gap-3">
                  <span className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md bg-primary/10">
                    <Icon className="size-3.5 text-primary" />
                  </span>
                  <div className="min-w-0">
                    <p className="flex items-center gap-1.5 font-heading text-sm font-semibold">
                      {feature.title}
                      <HelpTip id={feature.helpId} />
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {feature.description}
                    </p>
                  </div>
                </li>
              )
            })}
          </ul>
        </div>

        <div className="rounded-2xl border border-border/60 bg-card/80 p-5 shadow-sm backdrop-blur-sm md:p-6 lg:sticky lg:top-24">
          <Suspense
            fallback={
              <div className="h-36 animate-pulse rounded-xl bg-muted/40" />
            }
          >
            <ScanForm />
          </Suspense>
        </div>
      </div>
    </section>
  )
}