package scanner

import "testing"

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"full URL with www", "https://www.example.com/path?query=1#frag", "example.com"},
		{"http with port and uppercase", "http://Example.COM:8080/", "example.com"},
		{"www prefix only", "www.example.com", "example.com"},
		{"bare IP", "192.168.1.1", "192.168.1.1"},
		{"bare domain", "example.com", "example.com"},
		{"empty string", "", ""},
		{"scheme with userinfo", "https://user:pass@sub.example.com:9000/x", "sub.example.com"},
		{"uppercase www", "HTTPS://WWW.EXAMPLE.COM", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeHost(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeHost(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
