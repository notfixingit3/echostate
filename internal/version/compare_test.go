package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"0.0.1-beta.11", "0.0.1-beta.12", -1},
		{"0.0.1-beta.12", "0.0.1-beta.12", 0},
		{"0.0.1-beta.13", "0.0.1-beta.12", 1},
		{"v0.0.1-beta.12", "0.0.1-beta.12", 0},
		{"0.0.2", "0.0.1-beta.99", 1},
		{"0.0.1", "0.0.1-beta.1", 1},
		{"0.0.1-beta.2", "0.0.1-beta.10", -1},
	}

	for _, tc := range tests {
		require.Equal(t, tc.want, Compare(tc.a, tc.b), "%s vs %s", tc.a, tc.b)
	}
}

func TestIsPrerelease(t *testing.T) {
	require.True(t, IsPrerelease("0.0.1-beta.12"))
	require.False(t, IsPrerelease("0.0.1"))
}