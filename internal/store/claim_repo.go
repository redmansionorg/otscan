package store

import "context"

type ClaimRecord struct {
	ID           int    `json:"id"`
	RUID         string `json:"ruid"`
	Claimant     string `json:"claimant,omitempty"`
	SubmitBlock  uint64 `json:"submitBlock"`
	SubmitTime   uint64 `json:"submitTime"`
	Published    bool   `json:"published"`
	AUID         string `json:"auid,omitempty"`
	PUID         string `json:"puid,omitempty"`
	PublishBlock uint64 `json:"publishBlock"`
	PublishTime  uint64 `json:"publishTime"`
	BatchID      string `json:"batchID,omitempty"`
}

// UpsertClaim inserts or updates a claim record.
func (db *DB) UpsertClaim(ctx context.Context, c *ClaimRecord) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO claims (ruid, claimant, submit_block, submit_time, published,
			auid, puid, publish_block, publish_time, batch_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ruid) DO UPDATE SET
			claimant = COALESCE(NULLIF($2, ''), claims.claimant),
			published = $5,
			auid = COALESCE(NULLIF($6, ''), claims.auid),
			puid = COALESCE(NULLIF($7, ''), claims.puid),
			publish_block = CASE WHEN $8 > 0 THEN $8 ELSE claims.publish_block END,
			publish_time = CASE WHEN $9 > 0 THEN $9 ELSE claims.publish_time END,
			batch_id = COALESCE(NULLIF($10, ''), claims.batch_id)
	`, c.RUID, c.Claimant, c.SubmitBlock, c.SubmitTime, c.Published,
		c.AUID, c.PUID, c.PublishBlock, c.PublishTime, c.BatchID)
	return err
}

// SearchClaims searches claims by claimant, auid, or puid.
func (db *DB) SearchClaims(ctx context.Context, field, value string, offset, limit int) ([]ClaimRecord, int, error) {
	var col string
	switch field {
	case "claimant":
		col = "claimant"
	case "auid":
		col = "auid"
	case "puid":
		col = "puid"
	default:
		col = "claimant"
	}

	var total int
	db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM claims WHERE "+col+" = $1", value).Scan(&total)

	rows, err := db.Pool.Query(ctx, `
		SELECT ruid, COALESCE(claimant,''), submit_block, submit_time, published,
			COALESCE(auid,''), COALESCE(puid,''), publish_block, publish_time, COALESCE(batch_id,'')
		FROM claims WHERE `+col+` = $1 ORDER BY id DESC LIMIT $2 OFFSET $3
	`, value, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []ClaimRecord
	for rows.Next() {
		var c ClaimRecord
		rows.Scan(&c.RUID, &c.Claimant, &c.SubmitBlock, &c.SubmitTime, &c.Published,
			&c.AUID, &c.PUID, &c.PublishBlock, &c.PublishTime, &c.BatchID)
		results = append(results, c)
	}
	return results, total, nil
}

// GetClaim returns a single claim by RUID.
func (db *DB) GetClaim(ctx context.Context, ruid string) (*ClaimRecord, error) {
	c := &ClaimRecord{}
	err := db.Pool.QueryRow(ctx, `
		SELECT ruid, COALESCE(claimant,''), submit_block, submit_time, published,
			COALESCE(auid,''), COALESCE(puid,''), publish_block, publish_time, COALESCE(batch_id,'')
		FROM claims WHERE ruid = $1
	`, ruid).Scan(&c.RUID, &c.Claimant, &c.SubmitBlock, &c.SubmitTime, &c.Published,
		&c.AUID, &c.PUID, &c.PublishBlock, &c.PublishTime, &c.BatchID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetClaimCount returns total number of claims.
func (db *DB) GetClaimCount(ctx context.Context) (int, error) {
	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM claims").Scan(&count)
	return count, err
}

// GetSyncMeta retrieves a value from the sync_meta table.
func (db *DB) GetSyncMeta(ctx context.Context, key string) (string, error) {
	var value string
	err := db.Pool.QueryRow(ctx, "SELECT value FROM sync_meta WHERE key = $1", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSyncMeta sets a value in the sync_meta table (upsert).
func (db *DB) SetSyncMeta(ctx context.Context, key, value string) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO sync_meta (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
	`, key, value)
	return err
}

// ConflictSummary represents a conflicting AUID with multiple PUIDs.
type ConflictSummary struct {
	AUID          string `json:"auid"`
	PUIDCount     int    `json:"puidCount"`
	ClaimCount    int    `json:"claimCount"`
	EarliestBlock uint64 `json:"earliestBlock"`
	LatestBlock   uint64 `json:"latestBlock"`
}

// ListConflicts returns AUIDs that have claims from multiple PUIDs.
func (db *DB) ListConflicts(ctx context.Context, offset, limit int) ([]ConflictSummary, int, error) {
	var total int
	db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT auid FROM claims
			WHERE published = true AND auid IS NOT NULL AND auid != ''
			GROUP BY auid HAVING COUNT(DISTINCT puid) > 1
		) sub
	`).Scan(&total)

	rows, err := db.Pool.Query(ctx, `
		SELECT auid,
			COUNT(DISTINCT puid) AS puid_count,
			COUNT(*) AS claim_count,
			MIN(submit_block) AS earliest_block,
			MAX(submit_block) AS latest_block
		FROM claims
		WHERE published = true AND auid IS NOT NULL AND auid != ''
		GROUP BY auid
		HAVING COUNT(DISTINCT puid) > 1
		ORDER BY earliest_block ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []ConflictSummary
	for rows.Next() {
		var s ConflictSummary
		rows.Scan(&s.AUID, &s.PUIDCount, &s.ClaimCount, &s.EarliestBlock, &s.LatestBlock)
		results = append(results, s)
	}
	return results, total, nil
}

// ClaimStats holds aggregate claim statistics.
type ClaimStats struct {
	TotalClaims    int `json:"totalClaims"`
	PublishedCount int `json:"publishedCount"`
	UniqueAUIDs    int `json:"uniqueAuids"`
	UniquePUIDs    int `json:"uniquePuids"`
	ConflictAUIDs  int `json:"conflictAuids"`
}

// GetClaimStats returns aggregate statistics about claims.
func (db *DB) GetClaimStats(ctx context.Context) (*ClaimStats, error) {
	stats := &ClaimStats{}
	db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM claims").Scan(&stats.TotalClaims)
	db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM claims WHERE published = true").Scan(&stats.PublishedCount)
	db.Pool.QueryRow(ctx, "SELECT COUNT(DISTINCT auid) FROM claims WHERE auid IS NOT NULL AND auid != ''").Scan(&stats.UniqueAUIDs)
	db.Pool.QueryRow(ctx, "SELECT COUNT(DISTINCT puid) FROM claims WHERE puid IS NOT NULL AND puid != ''").Scan(&stats.UniquePUIDs)
	db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT auid FROM claims
			WHERE published = true AND auid IS NOT NULL AND auid != ''
			GROUP BY auid HAVING COUNT(DISTINCT puid) > 1
		) sub
	`).Scan(&stats.ConflictAUIDs)
	return stats, nil
}
