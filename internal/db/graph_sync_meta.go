package db

type graphSyncMeta struct {
	TargetID   string
	SnapshotID string
	ScannedAt  int64
}

func (m graphSyncMeta) with(extra map[string]any) map[string]any {
	out := map[string]any{
		"target_id":   m.TargetID,
		"snapshot_id": m.SnapshotID,
		"scanned_at":  m.ScannedAt,
	}
	for key, value := range extra {
		out[key] = value
	}
	return out
}
