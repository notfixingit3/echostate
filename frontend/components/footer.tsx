export function Footer() {
  return (
    <footer className="border-t border-border/60 py-6">
      <div className="container flex flex-col items-center justify-between gap-3 text-sm text-muted-foreground sm:flex-row">
        <span className="font-heading text-foreground/80">EchoState</span>
        <span className="text-xs">
          Passive reconnaissance · Snapshot history · PDF reports
        </span>
        <a
          href="https://github.com/notfixingit3/echostate"
          target="_blank"
          rel="noopener noreferrer"
          className="transition-colors hover:text-primary"
        >
          GitHub
        </a>
      </div>
    </footer>
  )
}