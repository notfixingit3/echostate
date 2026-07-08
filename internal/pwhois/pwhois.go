package pwhois

import (
	"context"
	"fmt"
	"net"
	"time"

	pwhoislib "github.com/georgestarcher/pwhois"
	"github.com/notfixingit3/echostate/internal/config"
)

const (
	maxBatchSize = 100
	// pwhois.org provides the IP WHOIS lookup service used by this package.
	// Data is sourced from Regional Internet Registries (RIRs) and is subject to
	// their respective terms of use. See https://pwhois.org/ for details.
	defaultServer = "whois.pwhois.org"
	defaultPort   = 43
)

// PWHOISRecord mirrors the fields returned by the pwhois IP lookup response.
type PWHOISRecord struct {
	IP                  string    `json:"ip"`
	OriginAS            string    `json:"origin_as"`
	Prefix              string    `json:"prefix"`
	OrgName             string    `json:"org_name"`
	AsnOrgName          string    `json:"asn_org_name"`
	NetworkName         string    `json:"network_name"`
	City                string    `json:"city"`
	Region              string    `json:"region"`
	Country             string    `json:"country"`
	CountryCode         string    `json:"country_code"`
	Latitude            float64   `json:"latitude"`
	Longitude           float64   `json:"longitude"`
	AsnPath             string    `json:"asn_path"`
	CacheDate           time.Time `json:"cache_date"`
	RouteOriginatedDate time.Time `json:"route_originated_date"`
	RouteOriginatedTS   int64     `json:"route_originated_ts"`
}

// WhoisServer is the package-level shim for the pwhois server configuration.
// Tests may replace it with a mock instance.
var WhoisServer = defaultWhoisServer()

func defaultWhoisServer() *pwhoislib.WhoisServer {
	return &pwhoislib.WhoisServer{
		Server:       defaultServer,
		Port:         defaultPort,
		BatchMaxSize: maxBatchSize,
	}
}

// Lookup performs a pwhois lookup for up to maxBatchSize IP addresses.
// It establishes a fresh connection per call so the package-level WhoisServer
// shim can be replaced in tests without carrying connection state.
func Lookup(ctx context.Context, ips []string) ([]PWHOISRecord, error) {
	if len(ips) == 0 {
		return nil, nil
	}
	if len(ips) > maxBatchSize {
		ips = ips[:maxBatchSize]
	}

	valid := make([]string, 0, len(ips))
	seen := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		if net.ParseIP(ip) == nil {
			continue
		}
		valid = append(valid, ip)
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid IP addresses")
	}

	settings := config.GetSettings()
	serverStr := settings.PwhoisServer
	if serverStr == "" {
		serverStr = WhoisServer.Server
	}
	server := &pwhoislib.WhoisServer{
		Server:       serverStr,
		Port:         WhoisServer.Port,
		BatchMaxSize: maxBatchSize,
	}
	server.SetDefaultValues()

	if err := server.Connect(); err != nil {
		return nil, fmt.Errorf("connect to pwhois server: %w", err)
	}
	defer server.Connection.Close()

	query, err := server.FormatIpQuery(valid)
	if err != nil {
		return nil, fmt.Errorf("format pwhois query: %w", err)
	}

	ch := make(chan pwhoislib.IpLookupResponse, 1)
	go func() {
		server.LookupIP(query, ch)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Error != nil {
			return nil, res.Error
		}
		return recordsFromLib(res.Response), nil
	}
}

func recordsFromLib(recs []pwhoislib.WhoIs) []PWHOISRecord {
	out := make([]PWHOISRecord, len(recs))
	for i, r := range recs {
		out[i] = PWHOISRecord{
			IP:                  r.IP,
			OriginAS:            r.OriginAS,
			Prefix:              r.Prefix,
			OrgName:             r.OrgName,
			AsnOrgName:          r.AsnOrgName,
			NetworkName:         r.NetworkName,
			City:                r.City,
			Region:              r.Region,
			Country:             r.Country,
			CountryCode:         r.CountryCode,
			Latitude:            r.Latitude,
			Longitude:           r.Longitude,
			AsnPath:             r.AsnPath,
			CacheDate:           r.CacheDate,
			RouteOriginatedDate: r.RouteOriginatedDate,
			RouteOriginatedTS:   r.RouteOriginatedTS,
		}
	}
	return out
}
