package config

import (
	"sync"
)

type AlertRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Enabled     bool     `json:"enabled"`
	MinSeverity string   `json:"min_severity"`
	MatchTypes  []string `json:"match_types"`
	WebhookIDs  []string `json:"webhook_ids"`
}

type SystemSettings struct {
	DNSServers                string      `json:"dns_servers"`
	DKIMSelectors             string      `json:"dkim_selectors,omitempty"`
	PwhoisServer              string      `json:"pwhois_server"`
	RateLimit                 float64     `json:"rate_limit"`
	APIKey                    string      `json:"api_key,omitempty"`
	ShodanAPIKey              string      `json:"shodan_api_key,omitempty"`
	CensysAPIID               string      `json:"censys_api_id,omitempty"`
	CensysAPISecret           string      `json:"censys_api_secret,omitempty"`
	HIBPAPIKey                string      `json:"hibp_api_key,omitempty"`
	RiskIQAPIUser             string      `json:"riskiq_api_user,omitempty"`
	RiskIQAPIKey              string      `json:"riskiq_api_key,omitempty"`
	ScanConcurrency           int         `json:"scan_concurrency"`
	ScanJobTimeoutSec         int         `json:"scan_job_timeout_sec"`
	DefaultGathererTimeoutSec int         `json:"default_gatherer_timeout_sec"`
	CTHTTPTimeoutSec          int         `json:"ct_http_timeout_sec"`
	TracerouteTimeoutSec      int         `json:"traceroute_timeout_sec"`
	ScreenshotTimeoutSec      int         `json:"screenshot_timeout_sec"`
	RetentionMaxSnapshots     int         `json:"retention_max_snapshots"`
	ScheduleEnabled           bool        `json:"schedule_enabled"`
	ScheduleIntervalMinutes   int         `json:"schedule_interval_minutes"`
	ScheduleStaleHours        int         `json:"schedule_stale_hours"`
	ScheduleTags              []string    `json:"schedule_tags"`
	AlertRules                []AlertRule `json:"alert_rules"`
	AuthEnabled               bool        `json:"auth_enabled"`
	EnrollmentCodeTTLHours    int         `json:"enrollment_code_ttl_hours"`
	EnrollmentCodeLength      int         `json:"enrollment_code_length"`
	RecoveryCodeLength        int         `json:"recovery_code_length"`
	SessionTTLHours           int         `json:"session_ttl_hours"`
	MaxCodeAttempts           int         `json:"max_code_attempts"`
	CodeAttemptWindowMinutes  int         `json:"code_attempt_window_minutes"`
	WebAuthnRPID              string      `json:"webauthn_rp_id"`
	WebAuthnRPOrigin          string      `json:"webauthn_rp_origin"`
}

var (
	defaultSettings = SystemSettings{
		DNSServers:                "8.8.8.8,1.1.1.1",
		PwhoisServer:              "whois.pwhois.org",
		RateLimit:                 30.0,
		ScanConcurrency:           2,
		ScanJobTimeoutSec:         120,
		DefaultGathererTimeoutSec: 20,
		CTHTTPTimeoutSec:          60,
		TracerouteTimeoutSec:      40,
		ScreenshotTimeoutSec:        25,
		RetentionMaxSnapshots:     0,
		ScheduleEnabled:           false,
		ScheduleIntervalMinutes:   60,
		ScheduleStaleHours:        24,
		AlertRules:                defaultAlertRules(),
		AuthEnabled:               true,
		EnrollmentCodeTTLHours:    24,
		EnrollmentCodeLength:      8,
		RecoveryCodeLength:        12,
		SessionTTLHours:           168,
		MaxCodeAttempts:           5,
		CodeAttemptWindowMinutes:  15,
	}
	GlobalSettings = defaultSettings
	settingsMu     sync.RWMutex
)

func defaultAlertRules() []AlertRule {
	return []AlertRule{
		{ID: "critical-all", Name: "Critical changes", Enabled: true, MinSeverity: "critical"},
		{ID: "warnings", Name: "Warnings and above", Enabled: true, MinSeverity: "warning"},
		{
			ID: "graph-drift", Name: "Graph path drift", Enabled: true, MinSeverity: "warning",
			MatchTypes: []string{
				"graph_bgp_origin_added", "graph_bgp_origin_removed",
				"graph_as_path_added", "graph_as_path_removed",
				"graph_rpki_changed",
				"graph_traceroute_hop_added", "graph_traceroute_hop_removed", "graph_traceroute_reordered",
			},
		},
	}
}

func GetSettings() SystemSettings {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	return GlobalSettings
}

func UpdateSettings(s SystemSettings) {
	settingsMu.Lock()
	GlobalSettings = normalizeSettings(s)
	settingsMu.Unlock()
}

func normalizeSettings(s SystemSettings) SystemSettings {
	if s.DNSServers == "" {
		s.DNSServers = defaultSettings.DNSServers
	}
	if s.PwhoisServer == "" {
		s.PwhoisServer = defaultSettings.PwhoisServer
	}
	if s.RateLimit <= 0 {
		s.RateLimit = defaultSettings.RateLimit
	}
	if s.ScanConcurrency <= 0 {
		s.ScanConcurrency = defaultSettings.ScanConcurrency
	}
	if s.ScanJobTimeoutSec <= 0 {
		s.ScanJobTimeoutSec = defaultSettings.ScanJobTimeoutSec
	}
	if s.DefaultGathererTimeoutSec <= 0 {
		s.DefaultGathererTimeoutSec = defaultSettings.DefaultGathererTimeoutSec
	}
	if s.CTHTTPTimeoutSec <= 0 {
		s.CTHTTPTimeoutSec = defaultSettings.CTHTTPTimeoutSec
	}
	if s.TracerouteTimeoutSec <= 0 {
		s.TracerouteTimeoutSec = defaultSettings.TracerouteTimeoutSec
	}
	if s.ScreenshotTimeoutSec <= 0 {
		s.ScreenshotTimeoutSec = defaultSettings.ScreenshotTimeoutSec
	}
	if s.ScheduleIntervalMinutes <= 0 {
		s.ScheduleIntervalMinutes = defaultSettings.ScheduleIntervalMinutes
	}
	if s.ScheduleStaleHours <= 0 {
		s.ScheduleStaleHours = defaultSettings.ScheduleStaleHours
	}
	if len(s.AlertRules) == 0 {
		s.AlertRules = defaultAlertRules()
	}
	if s.EnrollmentCodeTTLHours <= 0 {
		s.EnrollmentCodeTTLHours = defaultSettings.EnrollmentCodeTTLHours
	}
	if s.EnrollmentCodeLength <= 0 {
		s.EnrollmentCodeLength = defaultSettings.EnrollmentCodeLength
	}
	if s.RecoveryCodeLength <= 0 {
		s.RecoveryCodeLength = defaultSettings.RecoveryCodeLength
	}
	if s.SessionTTLHours <= 0 {
		s.SessionTTLHours = defaultSettings.SessionTTLHours
	}
	if s.MaxCodeAttempts <= 0 {
		s.MaxCodeAttempts = defaultSettings.MaxCodeAttempts
	}
	if s.CodeAttemptWindowMinutes <= 0 {
		s.CodeAttemptWindowMinutes = defaultSettings.CodeAttemptWindowMinutes
	}
	return s
}

// PublicSettings returns settings safe for API responses (secrets masked).
func PublicSettings(s SystemSettings) SystemSettings {
	out := s
	out.APIKey = maskSecret(out.APIKey)
	out.ShodanAPIKey = maskSecret(out.ShodanAPIKey)
	out.CensysAPISecret = maskSecret(out.CensysAPISecret)
	out.HIBPAPIKey = maskSecret(out.HIBPAPIKey)
	out.RiskIQAPIKey = maskSecret(out.RiskIQAPIKey)
	return out
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:4] + "****"
}