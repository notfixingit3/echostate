package config

import (
	"sync"
)

type SystemSettings struct {
	DNSServers   string  `json:"dns_servers"`    // Comma-separated
	PwhoisServer string  `json:"pwhois_server"`
	RateLimit    float64 `json:"rate_limit"`     // Tokens per minute
}

var (
	defaultSettings = SystemSettings{
		DNSServers:   "8.8.8.8,1.1.1.1",
		PwhoisServer: "whois.pwhois.org",
		RateLimit:    30.0,
	}
	GlobalSettings = defaultSettings
	settingsMu     sync.RWMutex
)

func GetSettings() SystemSettings {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	return GlobalSettings
}

func UpdateSettings(s SystemSettings) {
	settingsMu.Lock()
	GlobalSettings = s
	settingsMu.Unlock()
}
