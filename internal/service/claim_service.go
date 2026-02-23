package service

import (
	"context"

	"github.com/workshop1/otscan/internal/config"
	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/store"
)

type ClaimService struct {
	db        *store.DB
	rpcClient *rpc.Client
	cfg       *config.Config
}

func NewClaimService(db *store.DB, rpcClient *rpc.Client, cfg *config.Config) *ClaimService {
	return &ClaimService{db: db, rpcClient: rpcClient, cfg: cfg}
}

// SearchClaims searches DB first, then falls back to RPC.
func (s *ClaimService) SearchClaims(ctx context.Context, field, value string, offset, limit int) (*rpc.ListResult, error) {
	claims, total, err := s.db.SearchClaims(ctx, field, value, offset, limit)
	if err == nil && total > 0 {
		items := make([]rpc.ClaimRecordResult, len(claims))
		for i, c := range claims {
			items[i] = rpc.ClaimRecordResult{
				RUID:         c.RUID,
				Claimant:     c.Claimant,
				SubmitBlock:  c.SubmitBlock,
				SubmitTime:   c.SubmitTime,
				Published:    c.Published,
				AUID:         c.AUID,
				PUID:         c.PUID,
				PublishBlock: c.PublishBlock,
				PublishTime:  c.PublishTime,
			}
		}
		return &rpc.ListResult{
			Items:      items,
			TotalCount: uint32(total),
			Offset:     uint32(offset),
			Limit:      uint32(limit),
		}, nil
	}

	// RPC fallback
	url := s.pickNode()
	switch field {
	case "claimant":
		return s.rpcClient.ListByClaimant(ctx, url, value, uint32(offset), uint32(limit))
	case "auid":
		return s.rpcClient.ListByAuid(ctx, url, value, uint32(offset), uint32(limit))
	case "puid":
		return s.rpcClient.ListByPuid(ctx, url, value, uint32(offset), uint32(limit))
	}
	return nil, nil
}

// ListConflicts returns AUIDs with conflicting claims.
func (s *ClaimService) ListConflicts(ctx context.Context, offset, limit int) ([]store.ConflictSummary, int, error) {
	return s.db.ListConflicts(ctx, offset, limit)
}

// GetClaimStats returns aggregate claim statistics.
func (s *ClaimService) GetClaimStats(ctx context.Context) (*store.ClaimStats, error) {
	return s.db.GetClaimStats(ctx)
}

func (s *ClaimService) pickNode() string {
	if len(s.cfg.Nodes) > 0 {
		return s.cfg.Nodes[0].RPCURL
	}
	return ""
}
