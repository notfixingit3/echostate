package auth

import "github.com/notfixingit3/echostate/internal/config"

// Settings are auth tunables stored in SystemSettings (admin-configurable).
type Settings struct {
	AuthEnabled              bool
	EnrollmentCodeTTLHours   int
	EnrollmentCodeLength     int
	RecoveryCodeLength       int
	SessionTTLHours          int
	MaxCodeAttempts          int
	CodeAttemptWindowMinutes int
	WebAuthnRPID             string
	WebAuthnRPOrigin         string
}

func SettingsFromConfig(app config.SystemSettings, frontendURL string) Settings {
	s := Settings{
		AuthEnabled:              app.AuthEnabled,
		EnrollmentCodeTTLHours:   app.EnrollmentCodeTTLHours,
		EnrollmentCodeLength:     app.EnrollmentCodeLength,
		RecoveryCodeLength:       app.RecoveryCodeLength,
		SessionTTLHours:          app.SessionTTLHours,
		MaxCodeAttempts:          app.MaxCodeAttempts,
		CodeAttemptWindowMinutes: app.CodeAttemptWindowMinutes,
		WebAuthnRPID:             app.WebAuthnRPID,
		WebAuthnRPOrigin:         app.WebAuthnRPOrigin,
	}
	if s.EnrollmentCodeTTLHours <= 0 {
		s.EnrollmentCodeTTLHours = 24
	}
	if s.EnrollmentCodeLength <= 0 {
		s.EnrollmentCodeLength = 8
	}
	if s.RecoveryCodeLength <= 0 {
		s.RecoveryCodeLength = 12
	}
	if s.SessionTTLHours <= 0 {
		s.SessionTTLHours = 168
	}
	if s.MaxCodeAttempts <= 0 {
		s.MaxCodeAttempts = 5
	}
	if s.CodeAttemptWindowMinutes <= 0 {
		s.CodeAttemptWindowMinutes = 15
	}
	if s.WebAuthnRPOrigin == "" {
		s.WebAuthnRPOrigin = frontendURL
	}
	if s.WebAuthnRPID == "" {
		s.WebAuthnRPID = rpIDFromOrigin(s.WebAuthnRPOrigin)
	}
	return s
}

func rpIDFromOrigin(origin string) string {
	// hostname only; localhost is valid for dev passkeys
	u, err := parseOrigin(origin)
	if err != nil {
		return "localhost"
	}
	return u
}

func parseOrigin(origin string) (string, error) {
	if origin == "" {
		return "localhost", nil
	}
	// lightweight parse without importing net/url in hot path tests
	for i := 0; i < len(origin); i++ {
		if origin[i] == '/' && i > 0 && origin[i-1] == '/' {
			host := origin[i+1:]
			for j, ch := range host {
				if ch == ':' || ch == '/' {
					return host[:j], nil
				}
			}
			return host, nil
		}
	}
	return origin, nil
}