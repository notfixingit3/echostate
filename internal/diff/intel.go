package diff

import (
	"fmt"
	"sort"
	"strings"
)

func diffDNSIntel(previous, current map[string]any) []Entry {
	prevDNS, _ := previous["dns"].(map[string]any)
	curDNS, _ := current["dns"].(map[string]any)
	if curDNS == nil {
		return nil
	}

	var entries []Entry
	entries = append(entries, diffCAARecords(prevDNS, curDNS)...)
	entries = append(entries, diffDNSSEC(prevDNS, curDNS)...)
	entries = append(entries, diffMTASTS(prevDNS, curDNS)...)
	entries = append(entries, diffTLSRPT(prevDNS, curDNS)...)
	entries = append(entries, diffBIMI(prevDNS, curDNS)...)
	entries = append(entries, diffPTR(prevDNS, curDNS)...)
	return entries
}

func diffCrawlIntel(previous, current map[string]any) []Entry {
	prevCrawl, _ := previous["crawl"].(map[string]any)
	curCrawl, _ := current["crawl"].(map[string]any)
	if curCrawl == nil {
		return nil
	}

	var entries []Entry
	entries = append(entries, diffSecurityTxt(prevCrawl, curCrawl)...)
	entries = append(entries, diffOptionalCrawlFile(prevCrawl, curCrawl, "humans_txt", "humans_txt", "humans.txt")...)
	entries = append(entries, diffOptionalCrawlFile(prevCrawl, curCrawl, "ads_txt", "ads_txt", "ads.txt")...)
	return entries
}

func diffCAARecords(prevDNS, curDNS map[string]any) []Entry {
	prevSet := caaRecordSet(prevDNS)
	curSet := caaRecordSet(curDNS)
	return diffStringSetChanges(prevSet, curSet, "caa_added", "caa_removed", "dns.CAA", "CAA record")
}

func caaRecordSet(dns map[string]any) map[string]struct{} {
	set := make(map[string]struct{})
	if dns == nil {
		return set
	}
	for _, item := range mapAnySlice(dns["CAA"]) {
		record := stringVal(item, "record")
		if record == "" {
			record = fmt.Sprintf("%s %s", stringVal(item, "tag"), stringVal(item, "value"))
		}
		if record != "" {
			set[record] = struct{}{}
		}
	}
	return set
}

func diffDNSSEC(prevDNS, curDNS map[string]any) []Entry {
	prevStatus := dnssecStatus(prevDNS)
	curStatus := dnssecStatus(curDNS)
	if curStatus == "" || curStatus == prevStatus {
		return nil
	}

	severity := SeverityInfo
	if prevStatus == "signed" && curStatus == "unsigned" {
		severity = SeverityWarning
	}
	return []Entry{{
		Type:     "dnssec_status",
		Severity: severity,
		Field:    "dns.DNSSEC.status",
		Summary:  fmt.Sprintf("DNSSEC status changed: %s → %s", emptyDash(prevStatus), curStatus),
	}}
}

func dnssecStatus(dns map[string]any) string {
	if dns == nil {
		return ""
	}
	block, _ := dns["DNSSEC"].(map[string]any)
	return strings.ToLower(stringVal(block, "status"))
}

func diffMTASTS(prevDNS, curDNS map[string]any) []Entry {
	prevMode := mtaStsMode(prevDNS)
	curMode := mtaStsMode(curDNS)
	if curMode == "" {
		return nil
	}
	if prevMode == "" {
		return []Entry{{
			Type:     "mta_sts_added",
			Severity: SeverityInfo,
			Field:    "dns.MTA_STS.mode",
			Summary:  fmt.Sprintf("MTA-STS published (mode: %s)", curMode),
		}}
	}
	if curMode == prevMode {
		return nil
	}

	severity := SeverityInfo
	if prevMode == "enforce" && curMode != "enforce" {
		severity = SeverityWarning
	}
	return []Entry{{
		Type:     "mta_sts_mode",
		Severity: severity,
		Field:    "dns.MTA_STS.mode",
		Summary:  fmt.Sprintf("MTA-STS mode changed: %s → %s", prevMode, curMode),
	}}
}

func mtaStsMode(dns map[string]any) string {
	if dns == nil {
		return ""
	}
	block, _ := dns["MTA_STS"].(map[string]any)
	return strings.ToLower(stringVal(block, "mode"))
}

func diffTLSRPT(prevDNS, curDNS map[string]any) []Entry {
	prevRua := tlsRptRua(prevDNS)
	curRua := tlsRptRua(curDNS)
	if curRua == "" {
		if prevRua != "" {
			return []Entry{{
				Type:     "tls_rpt_removed",
				Severity: SeverityInfo,
				Field:    "dns.TLS_RPT",
				Summary:  "TLS-RPT reporting removed",
			}}
		}
		return nil
	}
	if prevRua == "" {
		return []Entry{{
			Type:     "tls_rpt_added",
			Severity: SeverityInfo,
			Field:    "dns.TLS_RPT",
			Summary:  fmt.Sprintf("TLS-RPT reporting added: %s", curRua),
		}}
	}
	if curRua != prevRua {
		return []Entry{{
			Type:     "tls_rpt_changed",
			Severity: SeverityInfo,
			Field:    "dns.TLS_RPT.rua",
			Summary:  fmt.Sprintf("TLS-RPT reporting changed: %s → %s", prevRua, curRua),
		}}
	}
	return nil
}

func tlsRptRua(dns map[string]any) string {
	if dns == nil {
		return ""
	}
	block, _ := dns["TLS_RPT"].(map[string]any)
	return stringVal(block, "rua")
}

func diffBIMI(prevDNS, curDNS map[string]any) []Entry {
	prevRecord := bimiRecord(prevDNS)
	curRecord := bimiRecord(curDNS)
	if curRecord == "" {
		if prevRecord != "" {
			return []Entry{{
				Type:     "bimi_removed",
				Severity: SeverityInfo,
				Field:    "dns.BIMI",
				Summary:  "BIMI record removed",
			}}
		}
		return nil
	}
	if prevRecord == "" {
		return []Entry{{
			Type:     "bimi_added",
			Severity: SeverityInfo,
			Field:    "dns.BIMI",
			Summary:  "BIMI record published",
		}}
	}
	if curRecord != prevRecord {
		return []Entry{{
			Type:     "bimi_changed",
			Severity: SeverityInfo,
			Field:    "dns.BIMI.record",
			Summary:  "BIMI record changed",
		}}
	}
	return nil
}

func bimiRecord(dns map[string]any) string {
	if dns == nil {
		return ""
	}
	block, _ := dns["BIMI"].(map[string]any)
	return stringVal(block, "record")
}

func diffPTR(prevDNS, curDNS map[string]any) []Entry {
	prevPTR := stringVal(prevDNS, "PTR")
	curPTR := stringVal(curDNS, "PTR")
	if curPTR == "" || curPTR == prevPTR {
		return nil
	}
	return []Entry{{
		Type:     "dns_ptr",
		Severity: SeverityInfo,
		Field:    "dns.PTR",
		Summary:  fmt.Sprintf("PTR record changed: %s → %s", emptyDash(prevPTR), curPTR),
	}}
}

func diffSecurityTxt(prevCrawl, curCrawl map[string]any) []Entry {
	prevTxt, _ := prevCrawl["security_txt"].(map[string]any)
	curTxt, _ := curCrawl["security_txt"].(map[string]any)
	if curTxt == nil {
		if prevTxt != nil {
			return []Entry{{
				Type:     "security_txt_removed",
				Severity: SeverityWarning,
				Field:    "crawl.security_txt",
				Summary:  "security.txt removed or unreachable",
			}}
		}
		return nil
	}
	if prevTxt == nil {
		return []Entry{{
			Type:     "security_txt_added",
			Severity: SeverityInfo,
			Field:    "crawl.security_txt",
			Summary:  "security.txt published",
		}}
	}

	var entries []Entry
	prevContacts := toStringSet(stringSlice(prevTxt, "contacts"))
	for _, contact := range stringSlice(curTxt, "contacts") {
		if _, ok := prevContacts[contact]; !ok {
			entries = append(entries, Entry{
				Type:     "security_contact_added",
				Severity: SeverityInfo,
				Field:    "crawl.security_txt.contacts",
				Summary:  fmt.Sprintf("security.txt contact added: %s", contact),
				Detail:   contact,
			})
		}
	}

	prevExpires := stringVal(prevTxt, "expires")
	curExpires := stringVal(curTxt, "expires")
	if curExpires != "" && curExpires != prevExpires {
		entries = append(entries, Entry{
			Type:     "security_txt_expires",
			Severity: SeverityInfo,
			Field:    "crawl.security_txt.expires",
			Summary:  fmt.Sprintf("security.txt expiry changed: %s → %s", emptyDash(prevExpires), curExpires),
		})
	}
	return entries
}

func diffOptionalCrawlFile(prevCrawl, curCrawl map[string]any, key, typ, label string) []Entry {
	prevBlock, prevOK := prevCrawl[key].(map[string]any)
	curBlock, curOK := curCrawl[key].(map[string]any)
	if curOK && !prevOK {
		return []Entry{{
			Type:     typ + "_added",
			Severity: SeverityInfo,
			Field:    "crawl." + key,
			Summary:  fmt.Sprintf("%s published", label),
		}}
	}
	if !curOK && prevOK {
		return []Entry{{
			Type:     typ + "_removed",
			Severity: SeverityInfo,
			Field:    "crawl." + key,
			Summary:  fmt.Sprintf("%s removed or unreachable", label),
		}}
	}
	if curOK && prevOK {
		prevSource := stringVal(prevBlock, "source")
		curSource := stringVal(curBlock, "source")
		if prevSource != curSource && curSource != "" {
			return []Entry{{
				Type:     typ + "_changed",
				Severity: SeverityInfo,
				Field:    "crawl." + key,
				Summary:  fmt.Sprintf("%s source changed: %s → %s", label, emptyDash(prevSource), curSource),
			}}
		}
	}
	return nil
}

func diffRedirectChain(prevWeb, curWeb map[string]any) []Entry {
	prevFinal := redirectFinalURL(prevWeb)
	curFinal := redirectFinalURL(curWeb)
	prevHops := redirectHopCount(prevWeb)
	curHops := redirectHopCount(curWeb)

	if curFinal == "" && curHops == 0 {
		return nil
	}
	if prevFinal == curFinal && prevHops == curHops {
		return nil
	}
	if prevFinal == "" && curHops > 0 {
		return []Entry{{
			Type:     "redirect_chain_added",
			Severity: SeverityInfo,
			Field:    "web.redirect_chain",
			Summary:  fmt.Sprintf("Redirect chain detected (%d hop(s) → %s)", curHops, curFinal),
		}}
	}

	severity := SeverityInfo
	if curHops > prevHops+1 {
		severity = SeverityWarning
	}
	return []Entry{{
		Type:     "redirect_chain_changed",
		Severity: severity,
		Field:    "web.redirect_chain",
		Summary:  fmt.Sprintf("Redirect chain changed: %d hop(s) → %s (was %d hop(s) → %s)", curHops, curFinal, prevHops, emptyDash(prevFinal)),
	}}
}

func redirectFinalURL(web map[string]any) string {
	if web == nil {
		return ""
	}
	chain := mapAnySlice(web["redirect_chain"])
	if len(chain) == 0 {
		return ""
	}
	last := chain[len(chain)-1]
	return stringVal(last, "url")
}

func redirectHopCount(web map[string]any) int {
	if web == nil {
		return 0
	}
	chain := mapAnySlice(web["redirect_chain"])
	if len(chain) <= 1 {
		return 0
	}
	return len(chain) - 1
}

func diffHSTSPreload(prevWeb, curWeb map[string]any) []Entry {
	prevStatus := hstsPreloadStatus(prevWeb)
	curStatus := hstsPreloadStatus(curWeb)
	if curStatus == "" || curStatus == prevStatus {
		return nil
	}
	return []Entry{{
		Type:     "hsts_preload_status",
		Severity: SeverityInfo,
		Field:    "web.hsts_preload.preload_status",
		Summary:  fmt.Sprintf("HSTS preload status changed: %s → %s", emptyDash(prevStatus), curStatus),
	}}
}

func hstsPreloadStatus(web map[string]any) string {
	if web == nil {
		return ""
	}
	block, _ := web["hsts_preload"].(map[string]any)
	return strings.ToLower(stringVal(block, "preload_status"))
}

func diffCookieNames(prevWeb, curWeb map[string]any) []Entry {
	prevSet := toStringSet(stringSlice(prevWeb, "cookie_names"))
	var added []string
	for _, name := range stringSlice(curWeb, "cookie_names") {
		if _, ok := prevSet[name]; !ok {
			added = append(added, name)
		}
	}
	sort.Strings(added)

	var entries []Entry
	for _, name := range added {
		entries = append(entries, Entry{
			Type:     "cookie_name_added",
			Severity: SeverityInfo,
			Field:    "web.cookie_names",
			Summary:  fmt.Sprintf("New cookie name: %s", name),
			Detail:   name,
		})
	}
	return entries
}

func diffSecurityHeaders(prevWeb, curWeb map[string]any) []Entry {
	prevHeaders, _ := prevWeb["security_headers"].(map[string]any)
	curHeaders, _ := curWeb["security_headers"].(map[string]any)
	if curHeaders == nil {
		return nil
	}

	watched := []struct {
		key  string
		typ  string
		name string
	}{
		{"content-security-policy", "csp_changed", "Content-Security-Policy"},
		{"strict-transport-security", "hsts_changed", "Strict-Transport-Security"},
		{"permissions-policy", "permissions_policy_changed", "Permissions-Policy"},
		{"referrer-policy", "referrer_policy_changed", "Referrer-Policy"},
	}

	var entries []Entry
	for _, header := range watched {
		prevVal := securityHeaderValue(prevHeaders, header.key)
		curVal := securityHeaderValue(curHeaders, header.key)
		if curVal == "" || curVal == prevVal {
			continue
		}
		severity := SeverityInfo
		if header.key == "content-security-policy" || header.key == "strict-transport-security" {
			if prevVal != "" {
				severity = SeverityWarning
			}
		}
		entries = append(entries, Entry{
			Type:     header.typ,
			Severity: severity,
			Field:    "web.security_headers." + header.key,
			Summary:  fmt.Sprintf("%s changed", header.name),
		})
	}
	return entries
}

func securityHeaderValue(headers map[string]any, key string) string {
	if headers == nil {
		return ""
	}
	if val, ok := headers[key]; ok {
		return strings.TrimSpace(fmt.Sprint(val))
	}
	lower := strings.ToLower(key)
	for k, v := range headers {
		if strings.ToLower(k) == lower {
			return strings.TrimSpace(fmt.Sprint(v))
		}
	}
	return ""
}

func diffCTCertificates(prevCT, curCT map[string]any) []Entry {
	prevSerials := ctCertificateSerials(prevCT)
	var entries []Entry
	for _, cert := range mapAnySlice(curCT["certificates"]) {
		serial := stringVal(cert, "serial")
		if serial == "" {
			continue
		}
		if _, ok := prevSerials[serial]; ok {
			continue
		}
		issuer := stringVal(cert, "issuer")
		summary := fmt.Sprintf("New CT certificate: serial %s", serial)
		if issuer != "" {
			summary = fmt.Sprintf("New CT certificate from %s (serial %s)", issuer, serial)
		}
		entries = append(entries, Entry{
			Type:     "new_ct_certificate",
			Severity: SeverityInfo,
			Field:    "ct.certificates",
			Summary:  summary,
			Detail:   serial,
		})
	}
	return entries
}

func ctCertificateSerials(ct map[string]any) map[string]struct{} {
	set := make(map[string]struct{})
	if ct == nil {
		return set
	}
	for _, cert := range mapAnySlice(ct["certificates"]) {
		serial := stringVal(cert, "serial")
		if serial != "" {
			set[serial] = struct{}{}
		}
	}
	return set
}

func diffTLSExtended(prevTLS, curTLS map[string]any) []Entry {
	if curTLS == nil {
		return nil
	}

	var entries []Entry

	prevVersion := stringVal(prevTLS, "tls_version")
	curVersion := stringVal(curTLS, "tls_version")
	if curVersion != "" && curVersion != prevVersion {
		entries = append(entries, Entry{
			Type:     "tls_version",
			Severity: SeverityInfo,
			Field:    "tls.tls_version",
			Summary:  fmt.Sprintf("TLS version changed: %s → %s", emptyDash(prevVersion), curVersion),
		})
	}

	prevCipher := stringVal(prevTLS, "negotiated_cipher")
	curCipher := stringVal(curTLS, "negotiated_cipher")
	if curCipher != "" && curCipher != prevCipher {
		entries = append(entries, Entry{
			Type:     "tls_cipher",
			Severity: SeverityInfo,
			Field:    "tls.negotiated_cipher",
			Summary:  fmt.Sprintf("TLS cipher changed: %s → %s", emptyDash(prevCipher), curCipher),
		})
	}

	prevSerial := stringVal(prevTLS, "serial")
	curSerial := stringVal(curTLS, "serial")
	if curSerial != "" && curSerial != prevSerial && prevSerial != "" {
		entries = append(entries, Entry{
			Type:     "tls_serial",
			Severity: SeverityWarning,
			Field:    "tls.serial",
			Summary:  fmt.Sprintf("TLS certificate serial changed: %s → %s", prevSerial, curSerial),
		})
	}

	prevOCSP := boolVal(prevTLS, "ocsp_stapled")
	curOCSP := boolVal(curTLS, "ocsp_stapled")
	if curOCSP != prevOCSP {
		severity := SeverityInfo
		if prevOCSP && !curOCSP {
			severity = SeverityWarning
		}
		state := "disabled"
		if curOCSP {
			state = "enabled"
		}
		entries = append(entries, Entry{
			Type:     "tls_ocsp_stapling",
			Severity: severity,
			Field:    "tls.ocsp_stapled",
			Summary:  fmt.Sprintf("OCSP stapling %s", state),
		})
	}

	return entries
}

func diffStringSetChanges(prevSet, curSet map[string]struct{}, addedType, removedType, field, label string) []Entry {
	var added, removed []string
	for item := range curSet {
		if _, ok := prevSet[item]; !ok {
			added = append(added, item)
		}
	}
	for item := range prevSet {
		if _, ok := curSet[item]; !ok {
			removed = append(removed, item)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)

	var entries []Entry
	for _, item := range added {
		entries = append(entries, Entry{
			Type: addedType, Severity: SeverityInfo, Field: field,
			Summary: fmt.Sprintf("%s added: %s", label, item), Detail: item,
		})
	}
	for _, item := range removed {
		entries = append(entries, Entry{
			Type: removedType, Severity: SeverityInfo, Field: field,
			Summary: fmt.Sprintf("%s removed: %s", label, item), Detail: item,
		})
	}
	return entries
}

func mapAnySlice(raw any) []map[string]any {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	var out []map[string]any
	for _, item := range list {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}
