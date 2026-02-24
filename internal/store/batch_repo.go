package store

import (
	"context"
	"fmt"
	"time"
)

type BatchRecord struct {
	ID             int       `json:"id"`
	BatchID        string    `json:"batchID"`
	OnChainID      int64     `json:"onChainID"`
	StartBlock     uint64    `json:"startBlock"`
	EndBlock       uint64    `json:"endBlock"`
	RootHash       string    `json:"rootHash"`
	OTSDigest      string    `json:"otsDigest,omitempty"`
	RUIDCount      int       `json:"ruidCount"`
	TriggerType    string    `json:"triggerType"`
	Status         string    `json:"status"`
	BTCTxID        string    `json:"btcTxID,omitempty"`
	BTCBlockHeight uint64    `json:"btcBlockHeight"`
	BTCTimestamp   uint64    `json:"btcTimestamp"`
	CreatedAt      int64     `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	AnchoredBy     string    `json:"anchoredBy,omitempty"`
	AnchorBlock    uint64    `json:"anchorBlock"`
	AnchoredByName string    `json:"anchoredByName,omitempty"` // not persisted, populated by service layer
}

// UpsertBatch inserts or updates a batch record.
func (db *DB) UpsertBatch(ctx context.Context, b *BatchRecord) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO batches (batch_id, on_chain_id, start_block, end_block, root_hash,
			ots_digest, ruid_count, trigger_type, status, btc_tx_id,
			btc_block_height, btc_timestamp, created_at, anchored_by, anchor_block, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
		ON CONFLICT (batch_id) DO UPDATE SET
			on_chain_id = COALESCE(NULLIF($2, 0), batches.on_chain_id),
			status = $9,
			btc_tx_id = COALESCE(NULLIF($10, ''), batches.btc_tx_id),
			btc_block_height = CASE WHEN $11 > 0 THEN $11 ELSE batches.btc_block_height END,
			btc_timestamp = CASE WHEN $12 > 0 THEN $12 ELSE batches.btc_timestamp END,
			ots_digest = COALESCE(NULLIF($6, ''), batches.ots_digest),
			anchored_by = COALESCE(NULLIF($14, ''), batches.anchored_by),
			anchor_block = CASE WHEN $15 > 0 THEN $15 ELSE batches.anchor_block END,
			updated_at = NOW()
	`, b.BatchID, b.OnChainID, b.StartBlock, b.EndBlock, b.RootHash,
		b.OTSDigest, b.RUIDCount, b.TriggerType, b.Status, b.BTCTxID,
		b.BTCBlockHeight, b.BTCTimestamp, b.CreatedAt, b.AnchoredBy, b.AnchorBlock)
	return err
}

// GetBatch returns a batch by batchID or onChainID.
func (db *DB) GetBatch(ctx context.Context, idOrOnChainID string) (*BatchRecord, error) {
	b := &BatchRecord{}
	err := db.Pool.QueryRow(ctx, `
		SELECT batch_id, COALESCE(on_chain_id, 0), start_block, end_block, root_hash,
			COALESCE(ots_digest, ''), ruid_count, COALESCE(trigger_type, ''), status,
			COALESCE(btc_tx_id, ''), COALESCE(btc_block_height, 0), COALESCE(btc_timestamp, 0),
			COALESCE(created_at, 0), updated_at,
			COALESCE(anchored_by, ''), COALESCE(anchor_block, 0)
		FROM batches WHERE batch_id = $1 OR on_chain_id::text = $1
		LIMIT 1
	`, idOrOnChainID).Scan(&b.BatchID, &b.OnChainID, &b.StartBlock, &b.EndBlock, &b.RootHash,
		&b.OTSDigest, &b.RUIDCount, &b.TriggerType, &b.Status, &b.BTCTxID,
		&b.BTCBlockHeight, &b.BTCTimestamp, &b.CreatedAt, &b.UpdatedAt,
		&b.AnchoredBy, &b.AnchorBlock)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GetBatchCount returns total number of batches.
func (db *DB) GetBatchCount(ctx context.Context) (int, error) {
	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM batches").Scan(&count)
	return count, err
}

// GetPendingBatchCount returns number of pending batches.
func (db *DB) GetPendingBatchCount(ctx context.Context) (int, error) {
	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM batches WHERE status = 'pending' OR status = 'submitted' OR status = 'confirmed'").Scan(&count)
	return count, err
}

// ListBatches returns batches with optional status or filter, ordered by on_chain_id desc.
// filter can be: "non-empty" (ruidCount > 0), "empty" (ruidCount == 0), "pending" (status != 'anchored').
// status filters by exact status value.
func (db *DB) ListBatches(ctx context.Context, status string, limit, offset int, filter ...string) ([]BatchRecord, int, error) {
	var total int
	selectCols := `SELECT batch_id, COALESCE(on_chain_id, 0), start_block, end_block, root_hash,
		COALESCE(ots_digest, ''), ruid_count, COALESCE(trigger_type, ''), status,
		COALESCE(btc_tx_id, ''), COALESCE(btc_block_height, 0), COALESCE(btc_timestamp, 0),
		COALESCE(created_at, 0), updated_at,
		COALESCE(anchored_by, ''), COALESCE(anchor_block, 0)
		FROM batches`

	// Determine WHERE clause from filter or status
	var where string
	f := ""
	if len(filter) > 0 {
		f = filter[0]
	}
	switch f {
	case "non-empty":
		where = " WHERE ruid_count > 0"
	case "empty":
		where = " WHERE ruid_count = 0"
	case "pending":
		where = " WHERE status != 'anchored'"
	default:
		if status != "" {
			countQuery := "SELECT COUNT(*) FROM batches WHERE status = $1"
			listQuery := selectCols + " WHERE status = $1 ORDER BY on_chain_id DESC NULLS LAST LIMIT $2 OFFSET $3"
			db.Pool.QueryRow(ctx, countQuery, status).Scan(&total)
			rows, err := db.Pool.Query(ctx, listQuery, status, limit, offset)
			if err != nil {
				return nil, 0, err
			}
			defer rows.Close()
			return scanBatches(rows, total)
		}
	}

	countQuery := "SELECT COUNT(*) FROM batches" + where
	listQuery := selectCols + where + " ORDER BY on_chain_id DESC NULLS LAST LIMIT $1 OFFSET $2"
	db.Pool.QueryRow(ctx, countQuery).Scan(&total)
	rows, err := db.Pool.Query(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanBatches(rows, total)
}

func scanBatches(rows interface{ Next() bool; Scan(...interface{}) error }, total int) ([]BatchRecord, int, error) {
	var batches []BatchRecord
	for rows.Next() {
		var b BatchRecord
		err := rows.Scan(&b.BatchID, &b.OnChainID, &b.StartBlock, &b.EndBlock, &b.RootHash,
			&b.OTSDigest, &b.RUIDCount, &b.TriggerType, &b.Status, &b.BTCTxID,
			&b.BTCBlockHeight, &b.BTCTimestamp, &b.CreatedAt, &b.UpdatedAt,
			&b.AnchoredBy, &b.AnchorBlock)
		if err != nil {
			return nil, 0, err
		}
		batches = append(batches, b)
	}
	return batches, total, nil
}

// UpsertBatchRUIDs stores RUIDs associated with a batch.
func (db *DB) UpsertBatchRUIDs(ctx context.Context, batchID string, ruids []string) error {
	for _, ruid := range ruids {
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO batch_ruids (batch_id, ruid) VALUES ($1, $2)
			ON CONFLICT (batch_id, ruid) DO NOTHING
		`, batchID, ruid)
		if err != nil {
			return err
		}
	}
	return nil
}

// BatchWithRUIDProgress represents a batch and its RUID sync progress.
type BatchWithRUIDProgress struct {
	BatchID     string
	RUIDCount   int
	SyncedRUIDs int
}

// ListBatchesWithIncompleteRUIDs returns batches where synced RUID count < expected.
func (db *DB) ListBatchesWithIncompleteRUIDs(ctx context.Context) ([]BatchWithRUIDProgress, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT b.batch_id, b.ruid_count, COALESCE(r.synced, 0) as synced_ruids
		FROM batches b
		LEFT JOIN (
			SELECT batch_id, COUNT(*) as synced
			FROM batch_ruids
			GROUP BY batch_id
		) r ON b.batch_id = r.batch_id
		WHERE b.ruid_count > 0 AND COALESCE(r.synced, 0) < b.ruid_count
		ORDER BY b.on_chain_id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []BatchWithRUIDProgress
	for rows.Next() {
		var b BatchWithRUIDProgress
		if err := rows.Scan(&b.BatchID, &b.RUIDCount, &b.SyncedRUIDs); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

// BulkInsertBatchRUIDs inserts a batch of RUIDs efficiently using a single query.
func (db *DB) BulkInsertBatchRUIDs(ctx context.Context, batchID string, ruids []string) error {
	if len(ruids) == 0 {
		return nil
	}

	// Build a bulk INSERT with VALUES list
	const batchSize = 500
	for i := 0; i < len(ruids); i += batchSize {
		end := i + batchSize
		if end > len(ruids) {
			end = len(ruids)
		}
		chunk := ruids[i:end]

		query := "INSERT INTO batch_ruids (batch_id, ruid) VALUES "
		args := make([]interface{}, 0, len(chunk)*2)
		for j, ruid := range chunk {
			if j > 0 {
				query += ","
			}
			query += fmt.Sprintf("($%d,$%d)", j*2+1, j*2+2)
			args = append(args, batchID, ruid)
		}
		query += " ON CONFLICT (batch_id, ruid) DO NOTHING"

		if _, err := db.Pool.Exec(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

// GetTotalRUIDCount returns the total number of RUIDs across all batches.
func (db *DB) GetTotalRUIDCount(ctx context.Context) int {
	var count int
	db.Pool.QueryRow(ctx, "SELECT COALESCE(SUM(ruid_count), 0) FROM batches").Scan(&count)
	return count
}

// GetBatchRUIDs returns RUIDs for a batch.
func (db *DB) GetBatchRUIDs(ctx context.Context, batchID string, offset, limit int) ([]string, int, error) {
	var total int
	db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM batch_ruids WHERE batch_id = $1", batchID).Scan(&total)

	rows, err := db.Pool.Query(ctx, `
		SELECT ruid FROM batch_ruids WHERE batch_id = $1 ORDER BY id LIMIT $2 OFFSET $3
	`, batchID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ruids []string
	for rows.Next() {
		var r string
		rows.Scan(&r)
		ruids = append(ruids, r)
	}
	return ruids, total, nil
}
