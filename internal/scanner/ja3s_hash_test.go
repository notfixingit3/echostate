package scanner

import "testing"

func TestDigestJA3S(t *testing.T) {
	hash := digestJA3S(&serverHelloBasic{
		Vers:        771,
		CipherSuite: 49195,
		Extensions:  []uint16{0x0a0a, 11, 10},
	})
	if len(hash) != 32 {
		t.Fatalf("digestJA3S() len = %d, want 32", len(hash))
	}
}
