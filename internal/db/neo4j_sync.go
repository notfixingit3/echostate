package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	graphCTSubdomainLimit = 100
	graphCertSANLimit     = 50
)

func syncBGPGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	asnMap, ok := rawData["asn"].(map[string]any)
	if !ok {
		return nil
	}

	prefix := stringProp(asnMap, "prefix")
	if prefix == "" {
		return nil
	}

	routing, _ := asnMap["routing"].(map[string]any)
	hijackRisk := stringProp(routing, "hijack_risk")
	rpkiStatus := overallRPKIStatus(routing)
	originASN := normalizeASNValue(asnMap["asn"])

	_, err := tx.Run(ctx, `
		MERGE (p:Prefix {cidr: $prefix})
		SET p.hijack_risk = $hijack_risk, p.rpki_status = $rpki_status
		WITH p
		MATCH (t:Target {id: $target_id})
		MERGE (t)-[r:IN_PREFIX]->(p)
		SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
	`, meta.with(map[string]any{
		"prefix":      prefix,
		"hijack_risk": hijackRisk,
		"rpki_status": rpkiStatus,
	}))
	if err != nil {
		return err
	}

	if originASN != "" {
		originRisk := originRouteRisk(hijackRisk, rpkiStatus)
		_, err = tx.Run(ctx, `
			MERGE (a:ASN {number: $asn})
			WITH a
			MATCH (p:Prefix {cidr: $prefix})
			MERGE (a)-[r:ANNOUNCES]->(p)
			SET r.matches_origin = true, r.risk = $risk, r.rpki_status = $rpki_status,
			    r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at, r.target_id = $target_id
		`, meta.with(map[string]any{
			"asn":         originASN,
			"prefix":      prefix,
			"risk":        originRisk,
			"rpki_status": rpkiStatus,
		}))
		if err != nil {
			return err
		}
	}

	if routing == nil {
		return nil
	}

	origins, ok := routing["visible_origins"].([]any)
	if !ok {
		return syncASPaths(ctx, tx, meta, prefix, routing)
	}

	for _, origin := range origins {
		asn := normalizeASNValue(origin)
		if asn == "" {
			continue
		}
		matchesOrigin := originASN != "" && asn == originASN
		asnRPki := rpkiStatusForASN(routing, asn)
		edgeRisk := visibleOriginRisk(hijackRisk, matchesOrigin, asnRPki)
		_, err = tx.Run(ctx, `
			MERGE (a:ASN {number: $asn})
			WITH a
			MATCH (p:Prefix {cidr: $prefix})
			MERGE (a)-[r:VISIBLE_ORIGIN]->(p)
			SET r.matches_origin = $matches_origin, r.risk = $risk, r.rpki_status = $rpki_status,
			    r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at, r.target_id = $target_id
		`, meta.with(map[string]any{
			"asn":            asn,
			"prefix":         prefix,
			"matches_origin": matchesOrigin,
			"risk":           edgeRisk,
			"rpki_status":    asnRPki,
		}))
		if err != nil {
			return err
		}
	}

	return syncASPaths(ctx, tx, meta, prefix, routing)
}

func syncASPaths(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, prefix string, routing map[string]any) error {
	pathsAny, ok := routing["as_paths"].([]any)
	if !ok || len(pathsAny) == 0 {
		return nil
	}

	for pathIndex, pathAny := range pathsAny {
		pathStr, ok := pathAny.(string)
		if !ok || strings.TrimSpace(pathStr) == "" {
			continue
		}
		asns := strings.Fields(strings.TrimSpace(pathStr))
		if len(asns) == 0 {
			continue
		}

		for i, asn := range asns {
			asn = strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(asn)), "AS")
			if asn == "" {
				continue
			}

			_, err := tx.Run(ctx, `
				MERGE (a:ASN {number: $asn})
			`, map[string]any{"asn": asn})
			if err != nil {
				return err
			}

			if i == 0 {
				_, err = tx.Run(ctx, `
					MATCH (t:Target {id: $target_id})
					MATCH (a:ASN {number: $asn})
					MERGE (t)-[r:HAS_AS_PATH {path_index: $path_index, position: 0}]->(a)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`, meta.with(map[string]any{
					"asn":        asn,
					"path_index": pathIndex,
				}))
				if err != nil {
					return err
				}
			}

			if i > 0 {
				prevASN := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(asns[i-1])), "AS")
				_, err = tx.Run(ctx, `
					MATCH (prev:ASN {number: $prev_asn})
					MATCH (next:ASN {number: $asn})
					MERGE (prev)-[r:AS_PATH_NEXT {target_id: $target_id, path_index: $path_index, position: $position}]->(next)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`, meta.with(map[string]any{
					"prev_asn":   prevASN,
					"asn":        asn,
					"path_index": pathIndex,
					"position":   i,
				}))
				if err != nil {
					return err
				}
			}
		}

		lastASN := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(asns[len(asns)-1])), "AS")
		if prefix != "" && lastASN != "" {
			_, err := tx.Run(ctx, `
				MATCH (a:ASN {number: $asn})
				MATCH (p:Prefix {cidr: $prefix})
				MERGE (a)-[r:PATH_TO_PREFIX {target_id: $target_id, path_index: $path_index}]->(p)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"asn":        lastASN,
				"prefix":     prefix,
				"path_index": pathIndex,
			}))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func syncCTGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	targetID := meta.TargetID
	ctMap, ok := rawData["ct"].(map[string]any)
	if !ok {
		return nil
	}
	if _, skipped := ctMap["skipped"]; skipped {
		return nil
	}

	subdomainsAny, ok := ctMap["subdomains"].([]any)
	if !ok || len(subdomainsAny) == 0 {
		return nil
	}

	count := 0
	var err error
	for _, subAny := range subdomainsAny {
		if count >= graphCTSubdomainLimit {
			break
		}
		host, ok := subAny.(string)
		if !ok {
			continue
		}
		host = strings.TrimSpace(strings.TrimSuffix(host, "."))
		if host == "" {
			continue
		}

		_, err = tx.Run(ctx, `
			MERGE (s:Subdomain {host: $host})
			WITH s
			MATCH (t:Target {id: $target_id})
			MERGE (t)-[r:DISCOVERED_VIA_CT]->(s)
			SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
		`, meta.with(map[string]any{"host": host}))
		if err != nil {
			return err
		}

		_, err = tx.Run(ctx, `
			MATCH (s:Subdomain {host: $host})
			MATCH (scanned:Target {host: $host})
			WHERE scanned.id <> $target_id
			MERGE (s)-[:SCANNED_AS]->(scanned)
		`, map[string]any{
			"host":      host,
			"target_id": targetID,
		})
		if err != nil {
			return err
		}

		count++
	}

	return nil
}

func syncCertSANGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	targetID := meta.TargetID
	tlsMap, ok := rawData["tls"].(map[string]any)
	if !ok {
		return nil
	}

	sansAny, ok := tlsMap["dns_names"].([]any)
	if !ok || len(sansAny) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var err error
	count := 0
	for _, sanAny := range sansAny {
		if count >= graphCertSANLimit {
			break
		}
		name := normalizeSAN(fmt.Sprint(sanAny))
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}

		_, err = tx.Run(ctx, `
			MERGE (s:CertSAN {name: $name})
			WITH s
			MATCH (t:Target {id: $target_id})
			MERGE (t)-[r:HAS_SAN]->(s)
			SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
		`, meta.with(map[string]any{"name": name}))
		if err != nil {
			return err
		}

		_, err = tx.Run(ctx, `
			MATCH (s:CertSAN {name: $name})
			MATCH (scanned:Target)
			WHERE toLower(scanned.host) = $host AND scanned.id <> $target_id
			MERGE (s)-[:SCANNED_AS]->(scanned)
		`, map[string]any{
			"name":      name,
			"host":      name,
			"target_id": targetID,
		})
		if err != nil {
			return err
		}

		count++
	}

	return nil
}

func normalizeSAN(name string) string {
	name = strings.TrimSpace(strings.TrimSuffix(name, "."))
	return strings.ToLower(name)
}

func syncDNSGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	dnsMap, ok := rawData["dns"].(map[string]any)
	if !ok {
		return nil
	}

	var err error
	if nsAny, ok := dnsMap["NS"].([]any); ok {
		for _, item := range nsAny {
			host := normalizeDNSHost(item)
			if host == "" {
				continue
			}
			if err = mergeDNSRelation(ctx, tx, meta, host, "USES_NS"); err != nil {
				return err
			}
		}
	}

	if mxAny, ok := dnsMap["MX"].([]any); ok {
		for _, item := range mxAny {
			host := parseMXHost(item)
			if host == "" {
				continue
			}
			if err = mergeDNSRelation(ctx, tx, meta, host, "USES_MX"); err != nil {
				return err
			}
		}
	}

	if cname, ok := dnsMap["CNAME"].(string); ok {
		host := normalizeDNSHost(cname)
		if host != "" {
			if err = mergeDNSRelation(ctx, tx, meta, host, "ALIASES_TO"); err != nil {
				return err
			}
		}
	}

	if parsed, ok := dnsMap["DMARC_PARSED"].(map[string]any); ok {
		policy := strings.TrimSpace(fmt.Sprint(parsed["policy"]))
		if policy != "" {
			_, err = tx.Run(ctx, `
				MERGE (d:DMARCPolicy {policy: $policy})
				SET d.subdomain_policy = $subdomain_policy,
					d.record = $record
				WITH d
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[r:HAS_DMARC]->(d)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"policy":           policy,
				"subdomain_policy": strings.TrimSpace(fmt.Sprint(parsed["subdomain_policy"])),
				"record":           strings.TrimSpace(fmt.Sprint(parsed["record"])),
			}))
			if err != nil {
				return err
			}
		}
	}

	if soa, ok := dnsMap["SOA"].(map[string]any); ok {
		zone := strings.TrimSpace(fmt.Sprint(soa["zone"]))
		if zone == "" {
			zone = strings.TrimSpace(fmt.Sprint(soa["mname"]))
		}
		if zone != "" {
			_, err = tx.Run(ctx, `
				MERGE (z:SOAZone {zone: $zone})
				SET z.mname = $mname,
					z.rname = $rname,
					z.serial = $serial,
					z.record = $record
				WITH z
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[r:HAS_SOA]->(z)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"zone":   zone,
				"mname":  strings.TrimSpace(fmt.Sprint(soa["mname"])),
				"rname":  strings.TrimSpace(fmt.Sprint(soa["rname"])),
				"serial": intProp(soa, "serial"),
				"record": strings.TrimSpace(fmt.Sprint(soa["record"])),
			}))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func mergeDNSRelation(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, host, relType string) error {
	_, err := tx.Run(ctx, fmt.Sprintf(`
		MERGE (d:DNSHost {host: $host})
		WITH d
		MATCH (t:Target {id: $target_id})
		MERGE (t)-[r:%s]->(d)
		SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
	`, relType), meta.with(map[string]any{"host": host}))
	return err
}

func normalizeDNSHost(value any) string {
	host := strings.TrimSpace(fmt.Sprint(value))
	host = strings.TrimSuffix(host, ".")
	return host
}

func parseMXHost(value any) string {
	record := strings.TrimSpace(fmt.Sprint(value))
	parts := strings.Fields(record)
	if len(parts) < 2 {
		return ""
	}
	return normalizeDNSHost(parts[1])
}

func syncTracerouteGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	trMap, ok := rawData["traceroute"].(map[string]any)
	if !ok {
		return nil
	}
	if _, skipped := trMap["skipped"]; skipped {
		return nil
	}

	vantages := tracerouteVantagesFromRaw(trMap)
	if len(vantages) == 0 {
		return nil
	}

	var resolvedIP string
	var err error
	if asnMap, ok := rawData["asn"].(map[string]any); ok {
		resolvedIP = stringProp(asnMap, "ip")
	}

	for _, vantage := range vantages {
		vantageID := stringProp(vantage, "id")
		if vantageID == "" {
			vantageID = "local"
		}
		vantageLabel := stringProp(vantage, "label")
		hopsAny, ok := vantage["hops"].([]any)
		if !ok || len(hopsAny) == 0 {
			continue
		}

		var prevHop int
		for _, hopAny := range hopsAny {
			hopMap, ok := hopAny.(map[string]any)
			if !ok {
				continue
			}

			hopNum := intProp(hopMap, "hop")
			if hopNum <= 0 {
				continue
			}

			ip := stringProp(hopMap, "ip")
			timeout := boolProp(hopMap, "timeout")
			rttMs := floatProp(hopMap, "rtt_ms")
			country := stringProp(hopMap, "country")
			city := stringProp(hopMap, "city")
			latitude := floatProp(hopMap, "latitude")
			longitude := floatProp(hopMap, "longitude")

			_, err = tx.Run(ctx, `
				MERGE (h:Hop {target_id: $target_id, hop: $hop, vantage: $vantage, snapshot_id: $snapshot_id})
				SET h.ip = $ip, h.timeout = $timeout, h.rtt_ms = $rtt_ms,
				    h.country = $country, h.city = $city,
				    h.latitude = $latitude, h.longitude = $longitude,
				    h.vantage_label = $vantage_label, h.scanned_at = $scanned_at
				WITH h
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[r:TRACEROUTE_HOP {order: $hop, vantage: $vantage}]->(h)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"hop":           hopNum,
				"vantage":       vantageID,
				"vantage_label": vantageLabel,
				"ip":            ip,
				"timeout":       timeout,
				"rtt_ms":        rttMs,
				"country":       country,
				"city":          city,
				"latitude":      latitude,
				"longitude":     longitude,
			}))
			if err != nil {
				return err
			}

			if prevHop > 0 {
				_, err = tx.Run(ctx, `
					MATCH (prev:Hop {target_id: $target_id, hop: $prev_hop, vantage: $vantage, snapshot_id: $snapshot_id})
					MATCH (next:Hop {target_id: $target_id, hop: $hop, vantage: $vantage, snapshot_id: $snapshot_id})
					MERGE (prev)-[r:NEXT_HOP]->(next)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`, meta.with(map[string]any{
					"prev_hop": prevHop,
					"hop":      hopNum,
					"vantage":  vantageID,
				}))
				if err != nil {
					return err
				}
			}

			if ip != "" && !timeout {
				_, err = tx.Run(ctx, `
					MERGE (s:SharedHop {ip: $ip})
					WITH s
					MATCH (h:Hop {target_id: $target_id, hop: $hop, vantage: $vantage, snapshot_id: $snapshot_id})
					MERGE (h)-[r:SHARED_AT]->(s)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`, meta.with(map[string]any{
					"ip":      ip,
					"hop":     hopNum,
					"vantage": vantageID,
				}))
				if err != nil {
					return err
				}
			}

			if ip != "" && resolvedIP != "" && ip == resolvedIP {
				_, err = tx.Run(ctx, `
					MERGE (i:IP {address: $ip})
					WITH i
					MATCH (h:Hop {target_id: $target_id, hop: $hop, vantage: $vantage, snapshot_id: $snapshot_id})
					MERGE (h)-[r:REACHES]->(i)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`, meta.with(map[string]any{
					"ip":      ip,
					"hop":     hopNum,
					"vantage": vantageID,
				}))
				if err != nil {
					return err
				}
			}

			prevHop = hopNum
		}
	}

	return nil
}

func tracerouteVantagesFromRaw(trMap map[string]any) []map[string]any {
	if vantagesAny, ok := trMap["vantages"].([]any); ok && len(vantagesAny) > 0 {
		var vantages []map[string]any
		for _, item := range vantagesAny {
			if vantage, ok := item.(map[string]any); ok {
				vantages = append(vantages, vantage)
			}
		}
		if len(vantages) > 0 {
			return vantages
		}
	}

	hopsAny, ok := trMap["hops"].([]any)
	if !ok || len(hopsAny) == 0 {
		return nil
	}

	return []map[string]any{
		{
			"id":    "local",
			"label": "Local scanner",
			"hops":  hopsAny,
		},
	}
}

func syncPeeringGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	asnMap, ok := rawData["asn"].(map[string]any)
	if !ok {
		return nil
	}

	asn := normalizeASNValue(asnMap["asn"])
	if asn == "" {
		return nil
	}

	peering, ok := asnMap["peeringdb"].(map[string]any)
	if !ok {
		return nil
	}

	ixlanAny, ok := peering["ixlan"].([]any)
	if !ok || len(ixlanAny) == 0 {
		return nil
	}

	for _, item := range ixlanAny {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}

		ixID := intProp(row, "ix_id")
		if ixID <= 0 {
			continue
		}

		ixName := stringProp(row, "ix_name")
		country := stringProp(row, "country")
		city := stringProp(row, "city")
		speed := intProp(row, "speed")

		_, err := tx.Run(ctx, `
			MERGE (x:IX {id: $ix_id})
			SET x.name = $name, x.country = $country, x.city = $city
			WITH x
			MERGE (a:ASN {number: $asn})
			MERGE (a)-[r:PRESENT_AT_IX]->(x)
			SET r.speed = $speed, r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at, r.target_id = $target_id
		`, meta.with(map[string]any{
			"ix_id":   ixID,
			"asn":     asn,
			"name":    ixName,
			"country": country,
			"city":    city,
			"speed":   speed,
		}))
		if err != nil {
			return err
		}
	}

	return nil
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
		return 0
	}
}

func boolProp(props map[string]any, key string) bool {
	if props == nil {
		return false
	}
	value, ok := props[key]
	if !ok || value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

func overallRPKIStatus(routing map[string]any) string {
	if routing == nil {
		return ""
	}
	rpki, ok := routing["rpki"].(map[string]any)
	if !ok {
		return ""
	}
	validating, ok := rpki["validating_roas"].([]any)
	if !ok || len(validating) == 0 {
		return ""
	}
	seenValid := false
	seenInvalid := false
	for _, item := range validating {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		switch strings.ToLower(stringProp(row, "validity")) {
		case "valid":
			seenValid = true
		case "invalid":
			seenInvalid = true
		}
	}
	if seenInvalid {
		return "invalid"
	}
	if seenValid {
		return "valid"
	}
	return ""
}

func rpkiStatusForASN(routing map[string]any, asn string) string {
	if routing == nil {
		return ""
	}
	rpki, ok := routing["rpki"].(map[string]any)
	if !ok {
		return ""
	}
	validating, ok := rpki["validating_roas"].([]any)
	if !ok {
		return ""
	}
	asn = normalizeASNValue(asn)
	for _, item := range validating {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		origin := normalizeASNValue(row["origin"])
		if origin != asn {
			continue
		}
		return strings.ToLower(stringProp(row, "validity"))
	}
	return overallRPKIStatus(routing)
}

func originRouteRisk(hijackRisk, rpkiStatus string) string {
	if rpkiStatus == "invalid" || hijackRisk == "high" {
		return "high"
	}
	if hijackRisk == "medium" {
		return "medium"
	}
	if hijackRisk == "low" && (rpkiStatus == "valid" || rpkiStatus == "") {
		return "low"
	}
	if hijackRisk == "" {
		return "unknown"
	}
	return hijackRisk
}

func visibleOriginRisk(hijackRisk string, matchesOrigin bool, rpkiStatus string) string {
	if !matchesOrigin {
		return "high"
	}
	if rpkiStatus == "invalid" {
		return "high"
	}
	if rpkiStatus == "valid" {
		return "low"
	}
	return originRouteRisk(hijackRisk, rpkiStatus)
}

func normalizeASNValue(value any) string {
	asn := strings.TrimSpace(fmt.Sprint(value))
	return strings.TrimPrefix(strings.ToUpper(asn), "AS")
}

func floatProp(props map[string]any, key string) float64 {
	if props == nil {
		return 0
	}
	value, ok := props[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}
