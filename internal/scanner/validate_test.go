package scanner

import "testing"

func TestValidateScanTarget(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"bare domain", "example.com", "example.com", false},
		{"URL with path", "https://www.example.com/path", "example.com", false},
		{"IPv4", "8.8.8.8", "8.8.8.8", false},
		{"IPv6", "2001:db8::1", "2001:db8::1", false},
		{"empty", "", "", true},
		{"single label typo", "ipmcom", "", true},
		{"localhost", "localhost", "", true},
		{"invalid chars", "exam ple.com", "", true},
		{"short TLD", "example.c", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateScanTarget(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ValidateScanTarget(%q) expected error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateScanTarget(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ValidateScanTarget(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
