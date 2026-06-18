import { HelpTip } from "@/components/help-tip"

export function PageHeader({
  title,
  description,
  helpId,
  children,
}: {
  title: string
  description?: string
  helpId?: string
  children?: React.ReactNode
}) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="flex flex-col gap-2">
        <h1 className="inline-flex items-center gap-2 font-heading text-2xl font-semibold tracking-tight md:text-3xl">
          {title}
          {helpId ? <HelpTip id={helpId} /> : null}
        </h1>
        {description ? (
          <p className="max-w-2xl text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {children}
    </div>
  )
}