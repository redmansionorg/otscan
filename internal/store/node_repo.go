package store

import (
	"context"
	"encoding/json"
	"time"
)

type NodeHistoryRecord struct {
	BlockNumber  uint64    `json:"blockNumber"`
	PendingCount int       `json:"pendingCount"`
	Status       string    `json:"status"`
	RecordedAt   time.Time `json:"recordedAt"`
}

type NodeRecord struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	RPCURL   string `json:"rpcUrl"`
	IsActive bool   `json:"isActive"`
}

type NodeStatusRecord struct {
	NodeID             int                    `json:"nodeId"`
	NodeName           string                 `json:"nodeName"`
	RPCURL             string                 `json:"rpcUrl"`
	Status             string                 `json:"status"`
	BlockNumber        uint64                 `json:"blockNumber"`
	OTSMode            string                 `json:"otsMode"`
	PendingCount       int                    `json:"pendingCount"`
	TotalCreated       int                    `json:"totalCreated"`
	TotalConfirmed     int                    `json:"totalConfirmed"`
	LastProcessedBlock uint64                 `json:"lastProcessedBlock"`
	Components         map[string]interface{} `json:"components,omitempty"`
	LastAnchor         *string                `json:"lastAnchor,omitempty"`
	Coinbase           string                 `json:"coinbase,omitempty"`
	PolledAt           time.Time              `json:"polledAt"`
}

// UpsertNode creates or updates a node record.
func (db *DB) UpsertNode(ctx context.Context, name, rpcURL string) (int, error) {
	var id int
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO nodes (name, rpc_url) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET rpc_url = $2, is_active = true
		RETURNING id
	`, name, rpcURL).Scan(&id)
	return id, err
}

// UpsertNodeStatus inserts or updates current node status.
func (db *DB) UpsertNodeStatus(ctx context.Context, nodeID int, status string, blockNumber uint64,
	otsMode string, pendingCount, totalCreated, totalConfirmed int,
	lastProcessedBlock uint64, components map[string]interface{}, lastAnchor *string, coinbase string) error {

	compJSON, _ := json.Marshal(components)

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO node_status (node_id, status, block_number, ots_mode, pending_count,
			total_created, total_confirmed, last_processed_block, components, last_anchor, coinbase, polled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (node_id) DO UPDATE SET
			status = $2, block_number = $3, ots_mode = $4, pending_count = $5,
			total_created = $6, total_confirmed = $7, last_processed_block = $8,
			components = $9, last_anchor = $10, coinbase = $11, polled_at = NOW()
	`, nodeID, status, blockNumber, otsMode, pendingCount, totalCreated, totalConfirmed,
		lastProcessedBlock, compJSON, lastAnchor, coinbase)
	return err
}

// InsertNodeStatusHistory adds a history record for trend tracking.
func (db *DB) InsertNodeStatusHistory(ctx context.Context, nodeID int, blockNumber uint64, pendingCount int, status string) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO node_status_history (node_id, block_number, pending_count, status)
		VALUES ($1, $2, $3, $4)
	`, nodeID, blockNumber, pendingCount, status)
	return err
}

// GetAllNodeStatus returns current status for all active nodes.
func (db *DB) GetAllNodeStatus(ctx context.Context) ([]NodeStatusRecord, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT n.id, n.name, n.rpc_url,
			COALESCE(ns.status, 'unknown'), COALESCE(ns.block_number, 0),
			COALESCE(ns.ots_mode, ''), COALESCE(ns.pending_count, 0),
			COALESCE(ns.total_created, 0), COALESCE(ns.total_confirmed, 0),
			COALESCE(ns.last_processed_block, 0), ns.components, ns.last_anchor,
			COALESCE(ns.coinbase, ''), COALESCE(ns.polled_at, NOW())
		FROM nodes n
		LEFT JOIN node_status ns ON ns.node_id = n.id
		WHERE n.is_active = true
		ORDER BY n.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []NodeStatusRecord
	for rows.Next() {
		var r NodeStatusRecord
		var compJSON []byte
		err := rows.Scan(&r.NodeID, &r.NodeName, &r.RPCURL,
			&r.Status, &r.BlockNumber, &r.OTSMode, &r.PendingCount,
			&r.TotalCreated, &r.TotalConfirmed, &r.LastProcessedBlock,
			&compJSON, &r.LastAnchor, &r.Coinbase, &r.PolledAt)
		if err != nil {
			return nil, err
		}
		if compJSON != nil {
			json.Unmarshal(compJSON, &r.Components)
		}
		results = append(results, r)
	}
	return results, nil
}

// GetNodeStatus returns status for a specific node by name.
func (db *DB) GetNodeStatus(ctx context.Context, name string) (*NodeStatusRecord, error) {
	r := &NodeStatusRecord{}
	var compJSON []byte
	err := db.Pool.QueryRow(ctx, `
		SELECT n.id, n.name, n.rpc_url,
			COALESCE(ns.status, 'unknown'), COALESCE(ns.block_number, 0),
			COALESCE(ns.ots_mode, ''), COALESCE(ns.pending_count, 0),
			COALESCE(ns.total_created, 0), COALESCE(ns.total_confirmed, 0),
			COALESCE(ns.last_processed_block, 0), ns.components, ns.last_anchor,
			COALESCE(ns.coinbase, ''), COALESCE(ns.polled_at, NOW())
		FROM nodes n
		LEFT JOIN node_status ns ON ns.node_id = n.id
		WHERE n.name = $1 AND n.is_active = true
	`, name).Scan(&r.NodeID, &r.NodeName, &r.RPCURL,
		&r.Status, &r.BlockNumber, &r.OTSMode, &r.PendingCount,
		&r.TotalCreated, &r.TotalConfirmed, &r.LastProcessedBlock,
		&compJSON, &r.LastAnchor, &r.Coinbase, &r.PolledAt)
	if err != nil {
		return nil, err
	}
	if compJSON != nil {
		json.Unmarshal(compJSON, &r.Components)
	}
	return r, nil
}

// GetAddressNodeMap returns a mapping from coinbase address (lowercased) to node name.
func (db *DB) GetAddressNodeMap(ctx context.Context) (map[string]string, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT n.name, COALESCE(LOWER(ns.coinbase), '')
		FROM nodes n
		LEFT JOIN node_status ns ON ns.node_id = n.id
		WHERE n.is_active = true AND ns.coinbase IS NOT NULL AND ns.coinbase != ''
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var name, addr string
		if err := rows.Scan(&name, &addr); err == nil && addr != "" {
			m[addr] = name
		}
	}
	return m, nil
}

// GetNodeHistory returns recent status history for a node (for trend charts).
func (db *DB) GetNodeHistory(ctx context.Context, name string, limit int) ([]NodeHistoryRecord, error) {
	if limit <= 0 || limit > 1000 {
		limit = 360 // ~1 hour at 10s interval
	}
	rows, err := db.Pool.Query(ctx, `
		SELECT nsh.block_number, nsh.pending_count, nsh.status, nsh.recorded_at
		FROM node_status_history nsh
		JOIN nodes n ON n.id = nsh.node_id
		WHERE n.name = $1
		ORDER BY nsh.recorded_at DESC
		LIMIT $2
	`, name, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []NodeHistoryRecord
	for rows.Next() {
		var r NodeHistoryRecord
		rows.Scan(&r.BlockNumber, &r.PendingCount, &r.Status, &r.RecordedAt)
		results = append(results, r)
	}
	// Reverse to chronological order
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results, nil
}
