package scanner

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/image/draw"
)

const (
	screenshotTimeout      = 22 * time.Second
	screenshotViewportW    = 1280
	screenshotViewportH    = 720
	screenshotThumbWidth   = 400
	screenshotJPEGQuality  = 72
	screenshotThumbQuality = 60
)

func newScreenshotGatherer(browserWSURL string) Gatherer {
	if browserWSURL == "" {
		browserWSURL = defaultBrowserWSURL
	}

	return func(ctx context.Context, host string) (string, map[string]any, error) {
		host = NormalizeHost(host)
		if host == "" {
			return "screenshot", map[string]any{"error": "empty host"}, errors.New("empty host")
		}

		ctx, cancel := context.WithTimeout(ctx, screenshotTimeout)
		defer cancel()

		resolvedWS := resolveBrowserWSURL(ctx, browserWSURL)
		allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, resolvedWS)
		defer allocCancel()

		capture, pageURL, err := captureScreenshot(allocCtx, "https://"+host)
		if err != nil {
			capture, pageURL, err = captureScreenshot(allocCtx, "http://"+host)
		}

		result := map[string]any{
			"captured_at": time.Now().UTC().Format(time.RFC3339),
			"url":         pageURL,
		}
		if err != nil {
			result["error"] = err.Error()
			return "screenshot", result, fmt.Errorf("screenshot gather failed for %s: %w", host, err)
		}

		result["width"] = capture.Width
		result["height"] = capture.Height
		result["format"] = "jpeg"
		result["thumbnail"] = capture.ThumbnailBase64
		return "screenshot", result, nil
	}
}

type screenshotCapture struct {
	Width           int
	Height          int
	ThumbnailBase64 string
}

func captureScreenshot(parent context.Context, pageURL string) (screenshotCapture, string, error) {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()

	var buf []byte
	var finalURL string

	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(screenshotViewportW, screenshotViewportH),
		chromedp.Navigate(pageURL),
		chromedp.Sleep(2 * time.Second),
		chromedp.Location(&finalURL),
		chromedp.CaptureScreenshot(&buf),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		return screenshotCapture{}, pageURL, err
	}
	if len(buf) == 0 {
		return screenshotCapture{}, finalURL, fmt.Errorf("empty screenshot buffer")
	}

	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return screenshotCapture{}, finalURL, fmt.Errorf("decode screenshot: %w", err)
	}

	bounds := img.Bounds()
	thumb, err := resizeJPEGThumbnail(img, screenshotThumbWidth, screenshotThumbQuality)
	if err != nil {
		return screenshotCapture{}, finalURL, err
	}

	return screenshotCapture{
		Width:           bounds.Dx(),
		Height:          bounds.Dy(),
		ThumbnailBase64: base64.StdEncoding.EncodeToString(thumb),
	}, finalURL, nil
}

func resizeJPEGThumbnail(src image.Image, maxWidth, quality int) ([]byte, error) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions")
	}

	targetWidth := width
	targetHeight := height
	if width > maxWidth {
		targetWidth = maxWidth
		targetHeight = maxWidth * height / width
		if targetHeight < 1 {
			targetHeight = 1
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
