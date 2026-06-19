package enrichment

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFriendlyError_StripsRequestURL(t *testing.T) {
	err := fmt.Errorf(`Get "https://web.archive.org/cdx/search/cdx?url=example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)`)
	assert.Equal(t, "request timed out", FriendlyError(err))
}

func TestFriendlyError_PreservesHTTPStatus(t *testing.T) {
	err := errors.New("wayback HTTP 503: service unavailable")
	assert.Equal(t, "wayback HTTP 503: service unavailable", FriendlyError(err))
}

func TestFriendlyError_StripsShodanAPIKeyFromURL(t *testing.T) {
	err := fmt.Errorf(`Get "https://api.shodan.io/shodan/host/1.2.3.4?key=supersecret": context deadline exceeded`)
	msg := FriendlyError(err)
	assert.Equal(t, "request timed out", msg)
	assert.NotContains(t, msg, "supersecret")
	assert.NotContains(t, msg, "key=")
}