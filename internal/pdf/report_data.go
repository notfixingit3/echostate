package pdf

import "github.com/notfixingit3/echostate/internal/models"

// ReportData is the full payload used to render a snapshot PDF.
type ReportData struct {
	Result         *models.ScanResult
	Changes        []string
	ChangeDetails  []models.ChangeDetail
	ScreenshotJPEG []byte
}