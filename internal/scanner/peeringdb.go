package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	peeringDBBaseURL  = "https://peeringdb.com/api"
	peeringDBMaxBytes = 512 * 1024
	peeringDBIXLimit  = 40
)

var (
	peeringDBBaseURLVar = peeringDBBaseURL
	fetchPeeringDBFunc  = fetchPeeringDB
)

func enrichPeeringDB(ctx context.Context, asn string) (map[string]any, error) {
	asn = strings.TrimSpace(strings.TrimPrefix(strings.ToUpper(asn), "AS"))
	if asn == "" {
		return nil, fmt.Errorf("empty asn")
	}
	return fetchPeeringDBFunc(ctx, asn)
}

func fetchPeeringDB(ctx context.Context, asn string) (map[string]any, error) {
	netBody, _, err := httpGet(ctx, peeringDBBaseURLVar+"/net?asn="+asn+"&depth=1", peeringDBMaxBytes)
	if err != nil {
		return nil, fmt.Errorf("peeringdb net lookup: %w", err)
	}

	var netPayload struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(netBody, &netPayload); err != nil {
		return nil, fmt.Errorf("peeringdb net parse: %w", err)
	}
	if len(netPayload.Data) == 0 {
		return nil, fmt.Errorf("peeringdb: no network for AS%s", asn)
	}

	net := netPayload.Data[0]
	netID := intProp(net, "id")
	result := map[string]any{
		"net_id": netID,
		"name":   stringProp(net, "name"),
		"asn":    asn,
	}

	ixByID := indexPeeringDBObjects(net["ix"])
	netixlan := collectPeeringDBObjects(net["netixlan"])
	if len(netixlan) == 0 {
		ixlanBody, _, err := httpGet(ctx, fmt.Sprintf("%s/netixlan?net_id=%d", peeringDBBaseURLVar, netID), peeringDBMaxBytes)
		if err == nil {
			var ixlanPayload struct {
				Data []map[string]any `json:"data"`
			}
			if err := json.Unmarshal(ixlanBody, &ixlanPayload); err == nil {
				netixlan = ixlanPayload.Data
			}
		}
	}

	var ixlan []map[string]any
	for _, row := range netixlan {
		if len(ixlan) >= peeringDBIXLimit {
			break
		}
		ixID := intProp(row, "ix_id")
		if ixID <= 0 {
			continue
		}

		ixName := stringProp(row, "name")
		country := ""
		city := ""
		if ix, ok := ixByID[ixID]; ok {
			if ixName == "" {
				ixName = stringProp(ix, "name")
			}
			country = stringProp(ix, "country")
			city = stringProp(ix, "city")
		}

		speed := intProp(row, "speed")
		entry := map[string]any{
			"ix_id":   ixID,
			"ix_name": ixName,
			"country": country,
			"city":    city,
			"speed":   speed,
		}
		if ipv4 := stringProp(row, "ipaddr4"); ipv4 != "" {
			entry["ipv4"] = ipv4
		}
		if ipv6 := stringProp(row, "ipaddr6"); ipv6 != "" {
			entry["ipv6"] = ipv6
		}
		ixlan = append(ixlan, entry)
	}

	result["ix_count"] = len(ixlan)
	result["ixlan"] = ixlan
	return result, nil
}

func indexPeeringDBObjects(value any) map[int]map[string]any {
	index := make(map[int]map[string]any)
	for _, item := range collectPeeringDBObjects(value) {
		id := intProp(item, "id")
		if id > 0 {
			index[id] = item
		}
	}
	return index
}

func collectPeeringDBObjects(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		var out []map[string]any
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				out = append(out, row)
			}
		}
		return out
	case []map[string]any:
		return typed
	case map[string]any:
		return []map[string]any{typed}
	default:
		return nil
	}
}
