package enrichment

import (
	"errors"
	"strings"
)

// FriendlyError returns a short, report-friendly message without raw request URLs.
func FriendlyError(err error) string {
	if err == nil {
		return ""
	}

	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "request failed"
	}

	if strings.HasPrefix(msg, `Get "`) || strings.HasPrefix(msg, `Post "`) || strings.HasPrefix(msg, `Head "`) {
		if end := strings.Index(msg, `": `); end > 0 {
			msg = strings.TrimSpace(msg[end+3:])
		}
	}

	switch {
	case strings.Contains(msg, "context deadline exceeded"),
		strings.Contains(msg, "Client.Timeout exceeded"),
		strings.Contains(msg, "i/o timeout"):
		return "request timed out"
	case strings.HasPrefix(msg, "wayback HTTP "):
		return msg
	case strings.HasPrefix(msg, "HTTP "):
		return msg
	default:
		return msg
	}
}

func friendlyServiceError(service string, err error) error {
	if err == nil {
		return nil
	}
	if msg := FriendlyError(err); msg != "" {
		return errors.New(service + ": " + msg)
	}
	return errors.New(service + ": request failed")
}
