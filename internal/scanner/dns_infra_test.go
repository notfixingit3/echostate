package scanner

import "testing"

func TestDetectCDNFromHostname(t *testing.T) {
	t.Parallel()

	tests := []struct {
		host string
		want string
	}{
		{"cdn.example.cloudflare.net.", "Cloudflare"},
		{"dualstack.example.cloudfront.net", "Amazon CloudFront"},
		{"example.fastly.net", "Fastly"},
		{"example.edgekey.net", "Akamai"},
		{"origin.example.com", ""},
	}

	for _, tt := range tests {
		if got := detectCDNFromHostname(tt.host); got != tt.want {
			t.Errorf("detectCDNFromHostname(%q) = %q, want %q", tt.host, got, tt.want)
		}
	}
}

func TestDetectMailProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mx   []string
		want string
	}{
		{[]string{"10 aspmx.l.google.com"}, "Google Workspace"},
		{[]string{"0 example-com.mail.protection.outlook.com"}, "Microsoft 365"},
		{[]string{"10 mx.zoho.com"}, "Zoho Mail"},
		{[]string{"10 mail.example.com"}, ""},
	}

	for _, tt := range tests {
		if got := detectMailProvider(tt.mx); got != tt.want {
			t.Errorf("detectMailProvider(%v) = %q, want %q", tt.mx, got, tt.want)
		}
	}
}

func TestEnrichInfraLabels(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"CNAME": "target.cdn.cloudflare.net.",
		"MX":    []string{"10 aspmx.l.google.com"},
	}
	enrichInfraLabels(data)

	labels, ok := data["INFRA_LABELS"].(map[string]any)
	if !ok {
		t.Fatalf("INFRA_LABELS = %#v", data["INFRA_LABELS"])
	}
	if labels["cdn_provider"] != "Cloudflare" {
		t.Fatalf("cdn_provider = %v", labels["cdn_provider"])
	}
	if labels["mail_provider"] != "Google Workspace" {
		t.Fatalf("mail_provider = %v", labels["mail_provider"])
	}
}