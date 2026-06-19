package pdf

import (
	"time"

	"github.com/notfixingit3/echostate/internal/models"
)

// PWhoisInfo is passive WHOIS enrichment stored on the snapshot row.
type PWhoisInfo struct {
	OriginAS    string
	OrgName     string
	CountryCode string
	City        string
	Prefix      string
	LookedUpAt  *time.Time
}

// HasData reports whether any pWhois field is populated.
func (p *PWhoisInfo) HasData() bool {
	if p == nil {
		return false
	}
	return p.OriginAS != "" || p.OrgName != "" || p.CountryCode != "" || p.City != "" || p.Prefix != ""
}

// ReportData is the full payload used to render a snapshot PDF.
type ReportData struct {
	Result         *models.ScanResult
	Changes        []string
	ChangeDetails  []models.ChangeDetail
	ScreenshotJPEG []byte
	ClientIP       string
	PWhois         *PWhoisInfo
}