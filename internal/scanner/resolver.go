package scanner

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
)

func getResolver() *net.Resolver {
	settings := config.GetSettings()
	if settings.DNSServers != "" {
		servers := strings.Split(settings.DNSServers, ",")
		return &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 5 * time.Second}
				return d.DialContext(ctx, "udp", strings.TrimSpace(servers[0])+":53")
			},
		}
	}
	return &net.Resolver{}
}
