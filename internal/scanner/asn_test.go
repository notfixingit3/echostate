package scanner

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestReverseIPForCymru(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want string
	}{
		{"google dns", "8.8.8.8", "8.8.8.8"},
		{"cloudflare dns", "1.1.1.1", "1.1.1.1"},
		{"team cymru example", "216.90.108.31", "31.108.90.216"},
		{"simple", "1.2.3.4", "4.3.2.1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := reverseIPForCymru(tc.ip)
			if got != tc.want {
				t.Errorf("reverseIPForCymru(%q) = %q, want %q", tc.ip, got, tc.want)
			}
		})
	}
}

func TestASNParse(t *testing.T) {
	t.Run("six field origin record", func(t *testing.T) {
		txt := "15169 | 8.8.8.0/24 | US | arin | 1992-12-01 | GOOGLE, US"
		got, err := parseCymruTXT(txt)
		if err != nil {
			t.Fatalf("parseCymruTXT error: %v", err)
		}
		want := map[string]any{
			"asn":       "15169",
			"prefix":    "8.8.8.0/24",
			"country":   "US",
			"registry":  "arin",
			"allocated": "1992-12-01",
			"as_name":   "GOOGLE, US",
		}
		assertMapEqual(t, got, want)
	})

	t.Run("five field origin record gets empty as_name", func(t *testing.T) {
		txt := "23028 | 216.90.108.0/24 | US | arin | 1998-09-25"
		got, err := parseCymruTXT(txt)
		if err != nil {
			t.Fatalf("parseCymruTXT error: %v", err)
		}
		if got["as_name"] != "" {
			t.Errorf("as_name = %q, want empty", got["as_name"])
		}
		if got["asn"] != "23028" {
			t.Errorf("asn = %q, want 23028", got["asn"])
		}
	})

	t.Run("asn name record", func(t *testing.T) {
		txt := "23028 | US | arin | 2002-01-04 | TEAM-CYMRU - Team Cymru Inc., US"
		got, err := parseCymruASNNameTXT(txt)
		if err != nil {
			t.Fatalf("parseCymruASNNameTXT error: %v", err)
		}
		want := "TEAM-CYMRU - Team Cymru Inc., US"
		if got != want {
			t.Errorf("parseCymruASNNameTXT = %q, want %q", got, want)
		}
	})

	t.Run("malformed record returns error", func(t *testing.T) {
		_, err := parseCymruTXT("not a valid record")
		if err == nil {
			t.Fatal("expected error for malformed record")
		}
	})

	t.Run("gatherASN with mocked DNS", func(t *testing.T) {
		origLookupIP := lookupIPAddrFunc
		origLookupTXT := lookupTXTFunc
		defer func() {
			lookupIPAddrFunc = origLookupIP
			lookupTXTFunc = origLookupTXT
		}()

		lookupIPAddrFunc = func(ctx context.Context, r *net.Resolver, host string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
		}

		lookupTXTFunc = func(ctx context.Context, r *net.Resolver, name string) ([]string, error) {
			switch name {
			case "8.8.8.8.origin.asn.cymru.com":
				return []string{"15169 | 8.8.8.0/24 | US | arin | 1992-12-01"}, nil
			case "AS15169.asn.cymru.com":
				return []string{"15169 | US | arin | 1992-12-01 | GOOGLE, US"}, nil
			}
			return nil, errors.New("unexpected query: " + name)
		}

		key, value, err := gatherASN(context.Background(), "dns.google")
		if err != nil {
			t.Fatalf("gatherASN error: %v", err)
		}
		if key != "asn" {
			t.Errorf("key = %q, want asn", key)
		}
		want := map[string]any{
			"asn":       "15169",
			"prefix":    "8.8.8.0/24",
			"country":   "US",
			"registry":  "arin",
			"allocated": "1992-12-01",
			"as_name":   "GOOGLE, US",
			"ip":        "8.8.8.8",
		}
		assertMapEqual(t, value, want)
	})

	t.Run("gatherASN with IP input skips resolution", func(t *testing.T) {
		origLookupIP := lookupIPAddrFunc
		origLookupTXT := lookupTXTFunc
		defer func() {
			lookupIPAddrFunc = origLookupIP
			lookupTXTFunc = origLookupTXT
		}()

		resolved := false
		lookupIPAddrFunc = func(ctx context.Context, r *net.Resolver, host string) ([]net.IPAddr, error) {
			resolved = true
			return nil, errors.New("should not be called")
		}

		lookupTXTFunc = func(ctx context.Context, r *net.Resolver, name string) ([]string, error) {
			if name == "8.8.8.8.origin.asn.cymru.com" {
				return []string{"15169 | 8.8.8.0/24 | US | arin | 1992-12-01 | GOOGLE, US"}, nil
			}
			return nil, errors.New("unexpected query: " + name)
		}

		_, value, err := gatherASN(context.Background(), "8.8.8.8")
		if err != nil {
			t.Fatalf("gatherASN error: %v", err)
		}
		if resolved {
			t.Error("expected IP input to skip DNS resolution")
		}
		if value["ip"] != "8.8.8.8" {
			t.Errorf("ip = %q, want 8.8.8.8", value["ip"])
		}
	})
}

func TestASNLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live ASN lookup in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), asnGatherTimeout)
	defer cancel()

	key, value, err := gatherASN(ctx, "one.one.one.one")
	if err != nil {
		t.Skipf("live ASN lookup unavailable, skipping: %v", err)
	}

	if key != "asn" {
		t.Errorf("key = %q, want asn", key)
	}

	for _, field := range []string{"asn", "prefix", "country", "registry", "allocated", "ip"} {
		v, ok := value[field]
		if !ok || v == "" {
			t.Errorf("expected non-empty %q in result", field)
		}
	}
}

func assertMapEqual(t *testing.T, got, want map[string]any) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("map length = %d, want %d", len(got), len(want))
	}
	for k, wantV := range want {
		gotV, ok := got[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if gotV != wantV {
			t.Errorf("%q = %v, want %v", k, gotV, wantV)
		}
	}
}
