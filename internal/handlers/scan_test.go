package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpsertTargetStoresNormalizedHost(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	ctx := context.Background()
	rawHost := "https://www.Example.COM:8080/path"
	expected := "example.com"

	id, err := h.upsertTarget(ctx, rawHost)
	require.NoError(t, err)
	require.NotEqual(t, [16]byte{}, id)

	var normalized string
	err = d.Pool.QueryRow(ctx, `
		SELECT normalized_host FROM targets WHERE id = $1
	`, id).Scan(&normalized)
	require.NoError(t, err)
	require.Equal(t, expected, normalized)
}
