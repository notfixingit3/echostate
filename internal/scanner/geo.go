package scanner

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var lookupGeoFunc = lookupGeo

type geoResult struct {
	Country   string  `json:"country,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

func lookupGeo(ctx context.Context, ip string) (geoResult, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return geoResult{}, fmt.Errorf("empty ip")
	}
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsPrivate() || parsed.IsLoopback() || parsed.IsLinkLocalUnicast() {
		return geoResult{}, fmt.Errorf("non-public ip %s", ip)
	}

	payload, err := ripeStat(ctx, "geoloc", ip)
	if err != nil {
		return geoResult{}, err
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		return geoResult{}, fmt.Errorf("missing geoloc data")
	}

	located, ok := data["located_resources"].([]any)
	if !ok || len(located) == 0 {
		return geoResult{}, fmt.Errorf("no geoloc results")
	}

	first, ok := located[0].(map[string]any)
	if !ok {
		return geoResult{}, fmt.Errorf("invalid geoloc entry")
	}

	locations, ok := first["locations"].([]any)
	if !ok || len(locations) == 0 {
		return geoResult{}, fmt.Errorf("no geoloc locations")
	}

	loc, ok := locations[0].(map[string]any)
	if !ok {
		return geoResult{}, fmt.Errorf("invalid geoloc location")
	}

	result := geoResult{
		Country: stringProp(loc, "country"),
		City:    stringProp(loc, "city"),
	}
	if lat, ok := loc["latitude"].(float64); ok {
		result.Latitude = lat
	}
	if lon, ok := loc["longitude"].(float64); ok {
		result.Longitude = lon
	}
	if result.Latitude == 0 && result.Longitude == 0 {
		return geoResult{}, fmt.Errorf("missing coordinates")
	}
	return result, nil
}

func enrichHopGeo(ctx context.Context, hops []map[string]any) {
	cache := make(map[string]map[string]any)
	for _, hop := range hops {
		if hop["timeout"] == true {
			continue
		}
		ip := stringProp(hop, "ip")
		if ip == "" {
			continue
		}
		if cached, ok := cache[ip]; ok {
			applyGeoToHop(hop, cached)
			continue
		}

		geo, err := lookupGeoFunc(ctx, ip)
		if err != nil {
			continue
		}

		cached := map[string]any{}
		if geo.Country != "" {
			cached["country"] = geo.Country
		}
		if geo.City != "" {
			cached["city"] = geo.City
		}
		if geo.Latitude != 0 {
			cached["latitude"] = geo.Latitude
		}
		if geo.Longitude != 0 {
			cached["longitude"] = geo.Longitude
		}
		cache[ip] = cached
		applyGeoToHop(hop, cached)
	}
}

func applyGeoToHop(hop, geo map[string]any) {
	for _, key := range []string{"country", "city", "latitude", "longitude"} {
		if geo[key] != nil {
			hop[key] = geo[key]
		}
	}
}

func stringProp(props map[string]any, key string) string {
	if props == nil {
		return ""
	}
	value, ok := props[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func intProp(props map[string]any, key string) int {
	if props == nil {
		return 0
	}
	value, ok := props[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" {
			return 0
		}
		parsed, err := strconv.Atoi(text)
		if err != nil {
			return 0
		}
		return parsed
	}
}