import { BrandLogo } from "@/components/brand-logo"
import { ScanForm } from "@/components/scan-form"
import { Badge } from "@/components/ui/badge"
import { GlobeIcon, NetworkIcon, MonitorIcon } from "lucide-react"

const features = [
  {
    icon: GlobeIcon,
    title: "WHOIS",
    description: "Registrar, nameservers, expiration, and raw registry data.",
  },
  {
    icon: NetworkIcon,
    title: "ASN / BGP",
    description: "Origin ASN, prefix, country, and allocation via Team Cymru.",
  },
  {
    icon: MonitorIcon,
    title: "Web intel",
    description: "Page title, final URL, and copyright signals from the landing page.",
  },
]

export default function Page() {
  return (
    <section className="container flex flex-1 flex-col items-center gap-12 py-12 md:gap-16 md:py-20">
      <div className="flex max-w-3xl flex-col items-center gap-6 text-center">
        <BrandLogo variant="banner" className="h-16 w-auto md:h-20" />
        <div className="flex flex-wrap items-center justify-center gap-2">
          <Badge variant="secondary" className="font-mono text-xs">
            Passive recon
          </Badge>
          <Badge variant="outline" className="font-mono text-xs">
            WHOIS · ASN · Web
          </Badge>
        </div>
        <h1 className="font-heading text-4xl font-semibold tracking-tight text-balance md:text-5xl lg:text-6xl">
          Know your target before you touch it.
        </h1>
        <p className="max-w-xl text-lg text-muted-foreground text-balance md:text-xl">
          Feed EchoState a host, IP, or URL and gather WHOIS, BGP/ASN, and
          webpage intelligence in seconds — with full snapshot history.
        </p>
      </div>

      <div className="w-full max-w-2xl rounded-2xl border border-border/60 bg-card/80 p-6 shadow-sm backdrop-blur-sm md:p-8">
        <ScanForm />
      </div>

      <div className="grid w-full max-w-4xl gap-4 sm:grid-cols-3">
        {features.map((feature) => (
          <div
            key={feature.title}
            className="flex flex-col gap-3 rounded-xl border border-border/50 bg-card/60 p-5 backdrop-blur-sm"
          >
            <feature.icon className="size-5 text-primary" />
            <h3 className="font-heading text-sm font-semibold">{feature.title}</h3>
            <p className="text-sm text-muted-foreground">{feature.description}</p>
          </div>
        ))}
      </div>
    </section>
  )
}