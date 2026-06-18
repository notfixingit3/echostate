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

func syncBGPGraph(ctx context.Context, tx neo4j.ManagedTransaction, targetID string, rawData map[string]any) error {
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
		MERGE (t)-[:IN_PREFIX]->(p)
	`, map[string]any{
		"prefix":      prefix,
		"target_id":   targetID,
		"hijack_risk": hijackRisk,
		"rpki_status": rpkiStatus,
	})
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
			SET r.matches_origin = true, r.risk = $risk, r.rpki_status = $rpki_status
		`, map[string]any{
			"asn":         originASN,
			"prefix":      prefix,
			"risk":        originRisk,
			"rpki_status": rpkiStatus,
		})
		if err != nil {
			return err
		}
	}

	if routing == nil {
		return nil
	}

	origins, ok := routing["visible_origins"].([]any)
	if !ok {
		return syncASPaths(ctx, tx, targetID, prefix, routing)
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
			SET r.matches_origin = $matches_origin, r.risk = $risk, r.rpki_status = $rpki_status
		`, map[string]any{
			"asn":            asn,
			"prefix":         prefix,
			"matches_origin": matchesOrigin,
			"risk":           edgeRisk,
			"rpki_status":    asnRPki,
		})
		if err != nil {
			return err
		}
	}

	return syncASPaths(ctx, tx, targetID, prefix, routing)
}

func syncASPaths(ctx context.Context, tx neo4j.ManagedTransaction, targetID, prefix string, routing map[string]any) error {
	_, err := tx.Run(ctx, `
		MATCH (t:Target {id: $target_id})-[r:HAS_AS_PATH]->()
		DELETE r
	`, map[string]any{"target_id": targetID})
	if err != nil {
		return err
	}

	_, err = tx.Run(ctx, `
		MATCH ()-[r:AS_PATH_NEXT {target_id: $target_id}]->()
		DELETE r
	`, map[string]any{"target_id": targetID})
	if err != nil {
		return err
	}

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

			_, err = tx.Run(ctx, `
				MERGE (a:ASN {number: $asn})
			`, map[string]any{"asn": asn})
			if err != nil {
				return err
			}

			if i == 0 {
				_, err = tx.Run(ctx, `
					MATCH (t:Target {id: $target_id})
					MATCH (a:ASN {number: $asn})
					MERGE (t)-[:HAS_AS_PATH {path_index: $path_index, position: 0}]->(a)
				`, map[string]any{
					"target_id":  targetID,
					"asn":        asn,
					"path_index": pathIndex,
				})
				if err != nil {
					return err
				}
			}

			if i > 0 {
				prevASN := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(asns[i-1])), "AS")
				_, err = tx.Run(ctx, `
					MATCH (prev:ASN {number: $prev_asn})
					MATCH (next:ASN {number: $asn})
					MERGE (prev)-[:AS_PATH_NEXT {target_id: $target_id, path_index: $path_index, position: $position}]->(next)
				`, map[string]any{
					"prev_asn":   prevASN,
					"asn":        asn,
					"target_id":  targetID,
					"path_index": pathIndex,
					"position":   i,
				})
				if err != nil {
					return err
				}
			}
		}

		lastASN := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(asns[len(asns)-1])), "AS")
		if prefix != "" && lastASN != "" {
			_, err = tx.Run(ctx, `
				MATCH (a:ASN {number: $asn})
				MATCH (p:Prefix {cidr: $prefix})
				MERGE (a)-[:PATH_TO_PREFIX {target_id: $target_id, path_index: $path_index}]->(p)
			`, map[string]any{
				"asn":        lastASN,
				"prefix":     prefix,
				"target_id":  targetID,
				"path_index": pathIndex,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func syncCTGraph(ctx context.Context, tx neo4j.ManagedTransaction, targetID string, rawData map[string]any) error {
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

	_, err := tx.Run(ctx, `
		MATCH (t:Target {id: $target_id})-[r:DISCOVERED_VIA_CT]->()
		DELETE r
	`, map[string]any{"target_id": targetID})
	if err != nil {
		return err
	}

	count := 0
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
			MERGE (t)-[:DISCOVERED_VIA_CT]->(s)
		`, map[string]any{
			"host":      host,
			"target_id": targetID,
		})
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

func syncCertSANGraph(ctx context.Context, tx neo4j.ManagedTransaction, targetID string, rawData map[string]any) error {
	tlsMap, ok := rawData["tls"].(map[string]any)
	if !ok {
		return nil
	}

	sansAny, ok := tlsMap["dns_names"].([]any)
	if !ok || len(sansAny) == 0 {
		return nil
	}

	_, err := tx.Run(ctx, `
		MATCH (t:Target {id: $target_id})-[r:HAS_SAN]->()
		DELETE r
	`, map[string]any{"target_id": targetID})
	if err != nil {
		return err
	}

	seen := make(map[string]struct{})
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
			MERGE (t)-[:HAS_SAN]->(s)
		`, map[string]any{
			"name":      name,
			"target_id": targetID,
		})
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

func syncDNSGraph(ctx context.Context, tx neo4j.ManagedTransaction, targetID string, rawData map[string]any) error {
	dnsMap, ok := rawData["dns"].(map[string]any)
	if !ok {
		return nil
	}

	_, err := tx.Run(ctx, `
		MATCH (t:Target {id: $target_id})-[r:USES_NS|USES_MX|ALIASES_TO]->()
		DELETE r
	`, map[string]any{"target_id": targetID})
	if err != nil {
		return err
	}

	if nsAny, ok := dnsMap["NS"].([]any); ok {
		for _, item := range nsAny {
			host := normalizeDNSHost(item)
			if host == "" {
				continue
			}
			if err = mergeDNSRelation(ctx, tx, targetID, host, "USES_NS"); err != nil {
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
			if err = mergeDNSRelation(ctx, tx, targetID, host, "USES_MX"); err != nil {
				return err
			}
		}
	}

	if cname, ok := dnsMap["CNAME"].(string); ok {
		host := normalizeDNSHost(cname)
		if host != "" {
			if err = mergeDNSRelation(ctx, tx, targetID, host, "ALIASES_TO"); err != nil {
				return err
			}
		}
	}

	return nil
}

func mergeDNSRelation(ctx context.Context, tx neo4j.ManagedTransaction, targetID, host, relType string) error {
	_, err := tx.Run(ctx, fmt.Sprintf(`
		MERGE (d:DNSHost {host: $host})
		WITH d
		MATCH (t:Target {id: $target_id})
		MERGE (t)-[:%s]->(d)
	`, relType), map[string]any{
		"host":      host,
		"target_id": targetID,
	})
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

func syncTracerouteGraph(ctx context.Context, tx neo4j.ManagedTransaction, targetID string, rawData map[string]any) error {
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

	_, err := tx.Run(ctx, `
		MATCH (t:Target {id: $target_id})-[r:TRACEROUTE_HOP]->()
		DELETE r
	`, map[string]any{"target_id": targetID})
	if err != nil {
		return err
	}

	_, err = tx.Run(ctx, `
		MATCH (h:Hop {target_id: $target_id})
		DETACH DELETE h
	`, map[string]any{"target_id": targetID})
	if err != nil {
		return err
	}

	var resolvedIP string
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
				MERGE (h:Hop {target_id: $target_id, hop: $hop, vantage: $vantage})
				SET h.ip = $ip, h.timeout = $timeout, h.rtt_ms = $rtt_ms,
				    h.country = $country, h.city = $city,
				    h.latitude = $latitude, h.longitude = $longitude,
				    h.vantage_label = $vantage_label
				WITH h
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[:TRACEROUTE_HOP {order: $hop, vantage: $vantage}]->(h)
			`, map[string]any{
				"target_id":     targetID,
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
			})
			if err != nil {
				return err
			}

			if prevHop > 0 {
				_, err = tx.Run(ctx, `
					MATCH (prev:Hop {target_id: $target_id, hop: $prev_hop, vantage: $vantage})
					MATCH (next:Hop {target_id: $target_id, hop: $hop, vantage: $vantage})
					MERGE (prev)-[:NEXT_HOP]->(next)
				`, map[string]any{
					"target_id": targetID,
					"prev_hop":  prevHop,
					"hop":       hopNum,
					"vantage":   vantageID,
				})
				if err != nil {
					return err
				}
			}

			if ip != "" && !timeout {
				_, err = tx.Run(ctx, `
					MERGE (s:SharedHop {ip: $ip})
					WITH s
					MATCH (h:Hop {target_id: $target_id, hop: $hop, vantage: $vantage})
					MERGE (h)-[:SHARED_AT]->(s)
				`, map[string]any{
					"ip":        ip,
					"target_id": targetID,
					"hop":       hopNum,
					"vantage":   vantageID,
				})
				if err != nil {
					return err
				}
			}

			if ip != "" && resolvedIP != "" && ip == resolvedIP {
				_, err = tx.Run(ctx, `
					MERGE (i:IP {address: $ip})
					WITH i
					MATCH (h:Hop {target_id: $target_id, hop: $hop, vantage: $vantage})
					MERGE (h)-[:REACHES]->(i)
				`, map[string]any{
					"ip":        ip,
					"target_id": targetID,
					"hop":       hopNum,
					"vantage":   vantageID,
				})
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

func syncPeeringGraph(ctx context.Context, tx neo4j.ManagedTransaction, targetID string, rawData map[string]any) error {
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

	_, err := tx.Run(ctx, `
		MATCH (a:ASN {number: $asn})-[r:PRESENT_AT_IX]->()
		DELETE r
	`, map[string]any{"asn": asn})
	if err != nil {
		return err
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

		_, err = tx.Run(ctx, `
			MERGE (x:IX {id: $ix_id})
			SET x.name = $name, x.country = $country, x.city = $city
			WITH x
			MERGE (a:ASN {number: $asn})
			MERGE (a)-[r:PRESENT_AT_IX]->(x)
			SET r.speed = $speed
		`, map[string]any{
			"ix_id":   ixID,
			"asn":     asn,
			"name":    ixName,
			"country": country,
			"city":    city,
			"speed":   speed,
		})
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