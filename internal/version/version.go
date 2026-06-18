package version

// Version is injected at link time via -ldflags; defaults to "dev" for local builds.
var Version = "dev"