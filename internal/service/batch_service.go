package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/workshop1/otscan/internal/config"
	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/store"
)

type BatchService struct {
	db        *store.DB
	rpcClient *rpc.Client
	cfg       *config.Config
}

func NewBatchService(db *store.DB, rpcClient *rpc.Client, cfg *config.Config) *BatchService {
	return &BatchService{db: db, rpcClient: rpcClient, cfg: cfg}
}

type BatchListResponse struct {
	Batches []store.BatchRecord `json:"batches"`
	Total   int                 `json:"total"`
	Page    int                 `json:"page"`
	PageSize int               `json:"pageSize"`
}

// ListBatches returns paginated batches from DB, falls back to RPC.
// filter can be: "non-empty", "empty", "pending".
func (s *BatchService) ListBatches(ctx context.Context, status string, page, pageSize int, filter ...string) (*BatchListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	f := ""
	if len(filter) > 0 {
		f = filter[0]
	}

	batches, total, err := s.db.ListBatches(ctx, status, pageSize, offset, filter...)
	if err == nil && (total > 0 || f != "") {
		// When a filter is active, 0 results is a valid outcome (e.g. no empty batches).
		s.enrichAnchoredByName(ctx, batches)
		return &BatchListResponse{Batches: batches, Total: total, Page: page, PageSize: pageSize}, nil
	}

	// RPC fallback (only for unfiltered queries)
	return s.listBatchesFromRPC(ctx, status)
}

// GetBatch returns a batch from DB, falls back to RPC.
func (s *BatchService) GetBatch(ctx context.Context, id string) (*rpc.BatchResult, error) {
	// Try DB first
	dbBatch, err := s.db.GetBatch(ctx, id)
	if err == nil {
		result := &rpc.BatchResult{
			BatchID:        dbBatch.BatchID,
			OnChainID:      uint64(dbBatch.OnChainID),
			StartBlock:     dbBatch.StartBlock,
			EndBlock:       dbBatch.EndBlock,
			RootHash:       dbBatch.RootHash,
			OTSDigest:      dbBatch.OTSDigest,
			RUIDCount:      uint32(dbBatch.RUIDCount),
			TriggerType:    dbBatch.TriggerType,
			Status:         dbBatch.Status,
			BTCTxID:        dbBatch.BTCTxID,
			BTCBlockHeight: dbBatch.BTCBlockHeight,
			BTCTimestamp:   dbBatch.BTCTimestamp,
			CreatedAt:      dbBatch.CreatedAt,
			AnchoredBy:     dbBatch.AnchoredBy,
			AnchorBlock:    dbBatch.AnchorBlock,
			AnchoredByName: s.lookupAnchoredByName(ctx, dbBatch.AnchoredBy),
		}
		return result, nil
	}

	// RPC fallback
	return s.rpcClient.GetBatch(ctx, s.pickNode(), id)
}

// GetBatchRUIDs returns RUIDs from DB, falls back to RPC.
func (s *BatchService) GetBatchRUIDs(ctx context.Context, batchID string, offset, limit int) (*rpc.RUIDsResult, error) {
	ruids, total, err := s.db.GetBatchRUIDs(ctx, batchID, offset, limit)
	if err == nil && total > 0 {
		return &rpc.RUIDsResult{
			BatchID: batchID,
			Total:   uint32(total),
			Offset:  uint32(offset),
			Limit:   uint32(limit),
			RUIDs:   ruids,
		}, nil
	}

	// RPC fallback
	return s.rpcClient.GetRUIDs(ctx, s.pickNode(), batchID, uint32(offset), uint32(limit))
}

// enrichAnchoredByName populates AnchoredByName for batches that have AnchoredBy set.
func (s *BatchService) enrichAnchoredByName(ctx context.Context, batches []store.BatchRecord) {
	addrMap, err := s.db.GetAddressNodeMap(ctx)
	if err != nil || len(addrMap) == 0 {
		return
	}
	for i := range batches {
		if batches[i].AnchoredBy != "" {
			if name, ok := addrMap[strings.ToLower(batches[i].AnchoredBy)]; ok {
				batches[i].AnchoredByName = name
			}
		}
	}
}

// lookupAnchoredByName returns the node name for an address, or empty string.
func (s *BatchService) lookupAnchoredByName(ctx context.Context, addr string) string {
	if addr == "" {
		return ""
	}
	addrMap, err := s.db.GetAddressNodeMap(ctx)
	if err != nil {
		return ""
	}
	return addrMap[strings.ToLower(addr)]
}

func (s *BatchService) pickNode() string {
	if len(s.cfg.Nodes) > 0 {
		return s.cfg.Nodes[0].RPCURL
	}
	return ""
}

func (s *BatchService) listBatchesFromRPC(ctx context.Context, status string) (*BatchListResponse, error) {
	url := s.pickNode()
	st, err := s.rpcClient.OTSStatus(ctx, url)
	if err != nil {
		return nil, err
	}
	total := 0
	if st.Stats != nil {
		total = st.Stats.TotalBatches
		if total == 0 {
			total = st.Stats.Total
		}
	}

	var batches []store.BatchRecord
	for i := total; i >= 1; i-- {
		b, err := s.rpcClient.GetBatch(ctx, url, fmt.Sprintf("%d", i))
		if err != nil {
			continue
		}
		if status != "" && b.Status != status {
			continue
		}
		batches = append(batches, store.BatchRecord{
			BatchID:        b.BatchID,
			OnChainID:      int64(b.OnChainID),
			StartBlock:     b.StartBlock,
			EndBlock:       b.EndBlock,
			RootHash:       b.RootHash,
			RUIDCount:      int(b.RUIDCount),
			TriggerType:    b.TriggerType,
			Status:         b.Status,
			BTCTxID:        b.BTCTxID,
			BTCBlockHeight: b.BTCBlockHeight,
			BTCTimestamp:   b.BTCTimestamp,
			CreatedAt:      b.CreatedAt,
			AnchoredBy:     b.AnchoredBy,
			AnchorBlock:    b.AnchorBlock,
		})
	}
	return &BatchListResponse{Batches: batches, Total: len(batches), Page: 1, PageSize: len(batches)}, nil
}
