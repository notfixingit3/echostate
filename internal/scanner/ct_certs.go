package scanner

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const ctMaxCertificates = 50

type ctCertRecord struct {
	NameValue    string `json:"name_value"`
	IssuerName   string `json:"issuer_name"`
	NotBefore    string `json:"not_before"`
	NotAfter     string `json:"not_after"`
	SerialNumber string `json:"serial_number"`
	CommonName   string `json:"common_name"`
	ID           int64  `json:"id"`
}

func parseCTCertificates(body []byte) ([]map[string]any, error) {
	var records []ctCertRecord
	if err := json.Unmarshal(body, &records); err != nil {
		return nil, fmt.Errorf("parse crt.sh JSON: %w", err)
	}

	seen := make(map[string]struct{})
	var certs []map[string]any

	for _, record := range records {
		serial := strings.TrimSpace(record.SerialNumber)
		if serial == "" {
			serial = fmt.Sprintf("id:%d", record.ID)
		}
		if _, ok := seen[serial]; ok {
			continue
		}
		seen[serial] = struct{}{}

		names := collectCTNames(record.NameValue, record.CommonName)
		certs = append(certs, map[string]any{
			"issuer":        strings.TrimSpace(record.IssuerName),
			"serial":        serial,
			"not_before":    strings.TrimSpace(record.NotBefore),
			"not_after":     strings.TrimSpace(record.NotAfter),
			"common_name":   strings.TrimSpace(record.CommonName),
			"sans":          names,
			"crt_sh_id":     record.ID,
		})

		if len(certs) >= ctMaxCertificates {
			break
		}
	}

	sort.Slice(certs, func(i, j int) bool {
		return fmt.Sprint(certs[i]["not_after"]) > fmt.Sprint(certs[j]["not_after"])
	})
	return certs, nil
}

func collectCTNames(nameValue, commonName string) []string {
	seen := make(map[string]struct{})
	var names []string
	add := func(name string) {
		name = normalizeCTName(name)
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	add(commonName)
	for _, raw := range strings.FieldsFunc(nameValue, func(r rune) bool {
		return r == '\n' || r == '\r'
	}) {
		add(raw)
	}
	sort.Strings(names)
	return names
}