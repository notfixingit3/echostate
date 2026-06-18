package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/models"
)

const (
	defaultBrowsePage  = 1
	defaultBrowseLimit = 20
	maxBrowseLimit     = 100
)

func (h *Handler) listTargets(c *gin.Context) {
	page, limit, ok := parseBrowsePagination(c)
	if !ok {
		return
	}
	q := c.Query("q")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	where := "WHERE 1=1"
	args := []any{}
	n := 1
	if q != "" {
		where += fmt.Sprintf(" AND (t.host ILIKE $%d OR $%d = ANY(t.tags))", n, n+1)
		args = append(args, "%"+q+"%", q)
		n += 2
	}

	var total int
	err := h.db.Pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM targets t %s
	`, where), args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count targets"})
		return
	}

	listArgs := make([]any, len(args))
	copy(listArgs, args)
	listArgs = append(listArgs, limit, (page-1)*limit)

	rows, err := h.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT t.id, t.host, t.tags, t.created_at,
			COALESCE((SELECT COUNT(*) FROM snapshots s WHERE s.target_id = t.id), 0) AS snapshot_count,
			(SELECT MAX(s.scanned_at) FROM snapshots s WHERE s.target_id = t.id) AS latest_snapshot_at,
			(SELECT s.raw_data->'asn'->>'asn' FROM snapshots s WHERE s.target_id = t.id ORDER BY s.scanned_at DESC LIMIT 1) AS latest_asn,
			(SELECT s.raw_data->'asn'->>'as_name' FROM snapshots s WHERE s.target_id = t.id ORDER BY s.scanned_at DESC LIMIT 1) AS latest_as_name,
			(SELECT s.raw_data->'web'->>'title' FROM snapshots s WHERE s.target_id = t.id ORDER BY s.scanned_at DESC LIMIT 1) AS latest_web_title
		FROM targets t
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, n, n+1), listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list targets"})
		return
	}
	defer rows.Close()

	targets := []models.TargetSummary{}
	for rows.Next() {
		var t models.TargetSummary
		var latestAt *time.Time
		err := rows.Scan(
			&t.ID, &t.Host, &t.Tags, &t.CreatedAt, &t.SnapshotCount, &latestAt,
			&t.LatestAsn, &t.LatestAsName, &t.LatestWebTitle,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan target"})
			return
		}
		t.LatestSnapshotAt = latestAt
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate targets"})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.TargetSummary]{
		Data:  targets,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (h *Handler) getTarget(c *gin.Context) {
	targetID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var detail models.TargetDetail
	var latestAt *time.Time
	err := h.db.Pool.QueryRow(ctx, `
		SELECT t.id, t.host, t.tags, t.created_at,
			COALESCE((SELECT COUNT(*) FROM snapshots s WHERE s.target_id = t.id), 0) AS snapshot_count,
			(SELECT MAX(s.scanned_at) FROM snapshots s WHERE s.target_id = t.id) AS latest_snapshot_at,
			(SELECT s.raw_data->'asn'->>'asn' FROM snapshots s WHERE s.target_id = t.id ORDER BY s.scanned_at DESC LIMIT 1) AS latest_asn,
			(SELECT s.raw_data->'asn'->>'as_name' FROM snapshots s WHERE s.target_id = t.id ORDER BY s.scanned_at DESC LIMIT 1) AS latest_as_name,
			(SELECT s.raw_data->'web'->>'title' FROM snapshots s WHERE s.target_id = t.id ORDER BY s.scanned_at DESC LIMIT 1) AS latest_web_title
		FROM targets t
		WHERE t.id = $1
	`, targetID).Scan(
		&detail.ID, &detail.Host, &detail.Tags, &detail.CreatedAt, &detail.SnapshotCount, &latestAt,
		&detail.LatestAsn, &detail.LatestAsName, &detail.LatestWebTitle,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch target"})
		return
	}
	detail.LatestSnapshotAt = latestAt

	if detail.SnapshotCount > 0 {
		var snapshot models.Snapshot
		var changeDetailsJSON []byte
		err := h.db.Pool.QueryRow(ctx, `
			SELECT s.id, s.target_id, s.scanned_at, s.last_seen, s.data_hash, s.raw_data, s.changes,
				COALESCE(s.change_details, '[]') AS change_details,
				COALESCE(s.client_ip, 'unknown') AS client_ip,
				s.pwhois_data, s.pwhois_looked_up_at, s.pwhois_origin_as, s.pwhois_org_name,
				s.pwhois_country_code, s.pwhois_city, s.pwhois_prefix
			FROM snapshots s
			WHERE s.target_id = $1
			ORDER BY s.scanned_at DESC
			LIMIT 1
		`, targetID).Scan(
			&snapshot.ID, &snapshot.TargetID, &snapshot.ScannedAt, &snapshot.LastSeen,
			&snapshot.DataHash, &snapshot.RawData, &snapshot.Changes, &changeDetailsJSON, &snapshot.ClientIP,
			&snapshot.PwhoisData, &snapshot.PwhoisLookedUp, &snapshot.PwhoisOriginAS,
			&snapshot.PwhoisOrgName, &snapshot.PwhoisCountry, &snapshot.PwhoisCity, &snapshot.PwhoisPrefix,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch latest snapshot"})
			return
		}
		snapshot.ChangeDetails, err = decodeChangeDetails(changeDetailsJSON)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode change details"})
			return
		}
		detail.LatestSnapshot = &snapshot
	}

	c.JSON(http.StatusOK, detail)
}

func (h *Handler) listTargetSnapshots(c *gin.Context) {
	targetID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	page, limit, ok := parseBrowsePagination(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	exists, err := h.targetExists(ctx, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify target"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}

	var total int
	err = h.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM snapshots WHERE target_id = $1
	`, targetID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count snapshots"})
		return
	}

	rows, err := h.db.Pool.Query(ctx, snapshotListSQL+`
		WHERE s.target_id = $1
		ORDER BY s.scanned_at DESC
		LIMIT $2 OFFSET $3
	`, targetID, limit, (page-1)*limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
		return
	}
	defer rows.Close()

	snapshots, err := scanSnapshotSummaries(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read snapshots"})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.SnapshotSummary]{
		Data:  snapshots,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (h *Handler) listSnapshots(c *gin.Context) {
	page, limit, ok := parseBrowsePagination(c)
	if !ok {
		return
	}

	where := "WHERE 1=1"
	args := []any{}
	n := 1

	if targetIDStr := c.Query("target_id"); targetIDStr != "" {
		targetID, err := uuid.Parse(targetIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_id"})
			return
		}
		where += fmt.Sprintf(" AND s.target_id = $%d", n)
		args = append(args, targetID)
		n++
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var total int
	err := h.db.Pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM snapshots s %s
	`, where), args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count snapshots"})
		return
	}

	listArgs := make([]any, len(args))
	copy(listArgs, args)
	listArgs = append(listArgs, limit, (page-1)*limit)

	rows, err := h.db.Pool.Query(ctx, snapshotListSQL+fmt.Sprintf(`
		%s
		ORDER BY s.scanned_at DESC
		LIMIT $%d OFFSET $%d
	`, where, n, n+1), listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snapshots"})
		return
	}
	defer rows.Close()

	snapshots, err := scanSnapshotSummaries(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read snapshots"})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.SnapshotSummary]{
		Data:  snapshots,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

const snapshotListSQL = `
	SELECT s.id, s.target_id, t.host, s.scanned_at, s.last_seen, s.data_hash, s.changes,
		COALESCE(s.client_ip, 'unknown') AS client_ip,
		s.pwhois_looked_up_at, s.pwhois_origin_as, s.pwhois_org_name,
		s.pwhois_country_code, s.pwhois_city, s.pwhois_prefix,
		s.raw_data->'asn'->>'asn' AS asn,
		s.raw_data->'asn'->>'as_name' AS as_name,
		s.raw_data->'web'->>'title' AS web_title,
		s.raw_data->'whois'->>'registrar' AS registrar,
		s.raw_data->'asn'->>'ip' AS resolved_ip
	FROM snapshots s
	JOIN targets t ON t.id = s.target_id
`

func scanSnapshotSummaries(rows pgx.Rows) ([]models.SnapshotSummary, error) {
	defer rows.Close()

	snapshots := []models.SnapshotSummary{}
	for rows.Next() {
		var s models.SnapshotSummary
		err := rows.Scan(
			&s.ID, &s.TargetID, &s.Host, &s.ScannedAt, &s.LastSeen, &s.DataHash, &s.Changes,
			&s.ClientIP,
			&s.PwhoisLookedUp, &s.PwhoisOriginAS, &s.PwhoisOrgName,
			&s.PwhoisCountry, &s.PwhoisCity, &s.PwhoisPrefix,
			&s.Asn, &s.AsName, &s.WebTitle, &s.Registrar, &s.ResolvedIP,
		)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (h *Handler) getSnapshot(c *gin.Context) {
	snapshotID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var snapshot models.Snapshot
	var changeDetailsJSON []byte
	err := h.db.Pool.QueryRow(ctx, `
		SELECT s.id, s.target_id, s.scanned_at, s.last_seen, s.data_hash, s.raw_data, s.changes,
			COALESCE(s.change_details, '[]') AS change_details,
			COALESCE(s.client_ip, 'unknown') AS client_ip,
			s.pwhois_data, s.pwhois_looked_up_at, s.pwhois_origin_as, s.pwhois_org_name,
			s.pwhois_country_code, s.pwhois_city, s.pwhois_prefix
		FROM snapshots s
		WHERE s.id = $1
	`, snapshotID).Scan(
		&snapshot.ID, &snapshot.TargetID, &snapshot.ScannedAt, &snapshot.LastSeen,
		&snapshot.DataHash, &snapshot.RawData, &snapshot.Changes, &changeDetailsJSON, &snapshot.ClientIP,
		&snapshot.PwhoisData, &snapshot.PwhoisLookedUp, &snapshot.PwhoisOriginAS,
		&snapshot.PwhoisOrgName, &snapshot.PwhoisCountry, &snapshot.PwhoisCity, &snapshot.PwhoisPrefix,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch snapshot"})
		return
	}
	snapshot.ChangeDetails, err = decodeChangeDetails(changeDetailsJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode change details"})
		return
	}

	c.JSON(http.StatusOK, snapshot)
}

func (h *Handler) listReports(c *gin.Context) {
	page, limit, ok := parseBrowsePagination(c)
	if !ok {
		return
	}

	where := "WHERE 1=1"
	args := []any{}
	n := 1

	if snapshotIDStr := c.Query("snapshot_id"); snapshotIDStr != "" {
		snapshotID, err := uuid.Parse(snapshotIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid snapshot_id"})
			return
		}
		where += fmt.Sprintf(" AND r.snapshot_id = $%d", n)
		args = append(args, snapshotID)
		n++
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := models.ReportStatus(statusStr)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		where += fmt.Sprintf(" AND r.status = $%d", n)
		args = append(args, statusStr)
		n++
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var total int
	err := h.db.Pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM reports r %s
	`, where), args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count reports"})
		return
	}

	listArgs := make([]any, len(args))
	copy(listArgs, args)
	listArgs = append(listArgs, limit, (page-1)*limit)

	rows, err := h.db.Pool.Query(ctx, fmt.Sprintf(`
		SELECT r.id, r.snapshot_id, COALESCE(t.host, 'unknown') AS host, r.status, r.created_at, r.completed_at
		FROM reports r
		JOIN snapshots s ON s.id = r.snapshot_id
		JOIN targets t ON t.id = s.target_id
		%s
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, n, n+1), listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}
	defer rows.Close()

	reports := []models.ReportSummary{}
	for rows.Next() {
		var r models.ReportSummary
		err := rows.Scan(&r.ID, &r.SnapshotID, &r.Host, &r.Status, &r.CreatedAt, &r.CompletedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan report"})
			return
		}
		reports = append(reports, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate reports"})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.ReportSummary]{
		Data:  reports,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (h *Handler) targetExists(ctx context.Context, targetID uuid.UUID) (bool, error) {
	var exists bool
	err := h.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM targets WHERE id = $1)
	`, targetID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check target existence: %w", err)
	}
	return exists, nil
}

func parseBrowsePagination(c *gin.Context) (page, limit int, ok bool) {
	pageStr := c.DefaultQuery("page", strconv.Itoa(defaultBrowsePage))
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defaultBrowseLimit))

	var err error
	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return 0, 0, false
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return 0, 0, false
	}

	if limit > maxBrowseLimit {
		limit = maxBrowseLimit
	}

	return page, limit, true
}

func (h *Handler) updateTargetTags(c *gin.Context) {
	targetID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		Tags []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	_, err := h.db.Pool.Exec(ctx, `
		UPDATE targets SET tags = $1 WHERE id = $2
	`, req.Tags, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tags"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) getSnapshotDiff(c *gin.Context) {
	snapshotID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var currentRawData map[string]any
	var targetID uuid.UUID
	var scannedAt time.Time

	err := h.db.Pool.QueryRow(ctx, `SELECT target_id, scanned_at, raw_data FROM snapshots WHERE id = $1`, snapshotID).Scan(&targetID, &scannedAt, &currentRawData)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}

	var previousRawData map[string]any
	err = h.db.Pool.QueryRow(ctx, `
		SELECT raw_data FROM snapshots
		WHERE target_id = $1 AND scanned_at < $2
		ORDER BY scanned_at DESC LIMIT 1
	`, targetID, scannedAt).Scan(&previousRawData)
	
	if err != nil && err != pgx.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch previous snapshot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current": currentRawData,
		"previous": previousRawData,
	})
}

func (h *Handler) listTargetScreenshots(c *gin.Context) {
	targetID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	page, limit, ok := parseBrowsePagination(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	exists, err := h.targetExists(ctx, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify target"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}

	var total int
	err = h.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM snapshots s
		LEFT JOIN snapshot_blobs b
		  ON b.snapshot_id = s.id AND b.kind = 'screenshot_thumbnail'
		WHERE s.target_id = $1
		  AND (
		    b.id IS NOT NULL
		    OR s.raw_data->'screenshot'->>'thumbnail' IS NOT NULL
		    OR s.raw_data->'screenshot'->>'error' IS NOT NULL
		  )
	`, targetID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count screenshots"})
		return
	}

	rows, err := h.db.Pool.Query(ctx, `
		SELECT s.id, s.scanned_at,
			COALESCE(s.raw_data->'screenshot'->>'url', '') AS url,
			COALESCE(
				NULLIF(encode(b.data, 'base64'), ''),
				s.raw_data->'screenshot'->>'thumbnail',
				''
			) AS thumbnail,
			COALESCE(s.raw_data->'screenshot'->>'format', '') AS format,
			COALESCE(s.raw_data->'screenshot'->>'error', '') AS error,
			COALESCE(NULLIF(s.raw_data->'screenshot'->>'width', '')::int, 0) AS width,
			COALESCE(NULLIF(s.raw_data->'screenshot'->>'height', '')::int, 0) AS height
		FROM snapshots s
		LEFT JOIN snapshot_blobs b
		  ON b.snapshot_id = s.id AND b.kind = 'screenshot_thumbnail'
		WHERE s.target_id = $1
		  AND (
		    b.id IS NOT NULL
		    OR s.raw_data->'screenshot'->>'thumbnail' IS NOT NULL
		    OR s.raw_data->'screenshot'->>'error' IS NOT NULL
		  )
		ORDER BY s.scanned_at DESC
		LIMIT $2 OFFSET $3
	`, targetID, limit, (page-1)*limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list screenshots"})
		return
	}
	defer rows.Close()

	entries := []models.ScreenshotEntry{}
	for rows.Next() {
		var entry models.ScreenshotEntry
		if err := rows.Scan(
			&entry.SnapshotID,
			&entry.ScannedAt,
			&entry.URL,
			&entry.Thumbnail,
			&entry.Format,
			&entry.Error,
			&entry.Width,
			&entry.Height,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read screenshots"})
			return
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate screenshots"})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.ScreenshotEntry]{
		Data:  entries,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

