package service

import (
	"context"
	"fmt"

	"github.com/workshop1/otscan/internal/cache"
	"github.com/workshop1/otscan/internal/config"
	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/store"
)

type NodeService struct {
	db        *store.DB
	cache     *cache.Cache
	rpcClient *rpc.Client
	cfg       *config.Config
}

func NewNodeService(db *store.DB, c *cache.Cache, rpcClient *rpc.Client, cfg *config.Config) *NodeService {
	return &NodeService{db: db, cache: c, rpcClient: rpcClient, cfg: cfg}
}

type NodeStatusView struct {
	Name               string                 `json:"name"`
	RPCURL             string                 `json:"rpcUrl"`
	Status             string                 `json:"status"`
	BlockNumber        uint64                 `json:"blockNumber"`
	OTSMode            string                 `json:"otsMode,omitempty"`
	PendingCount       int                    `json:"pendingCount"`
	TotalCreated       int                    `json:"totalCreated"`
	TotalConfirmed     int                    `json:"totalConfirmed"`
	LastProcessedBlock uint64                 `json:"lastProcessedBlock"`
	Components         map[string]interface{} `json:"components,omitempty"`
	LastAnchor         *string                `json:"lastAnchor,omitempty"`
	Coinbase           string                 `json:"coinbase,omitempty"`
}

// GetAllNodes tries: Redis cache → DB → RPC fallback
func (s *NodeService) GetAllNodes(ctx context.Context) ([]NodeStatusView, error) {
	// Try Redis cache first
	names := make([]string, len(s.cfg.Nodes))
	for i, n := range s.cfg.Nodes {
		names[i] = n.Name
	}

	cached, err := s.cache.GetAllNodeStatuses(ctx, names)
	if err == nil && len(cached) == len(s.cfg.Nodes) {
		views := make([]NodeStatusView, len(cached))
		for i, c := range cached {
			views[i] = NodeStatusView{
				Name: c.Name, RPCURL: c.RPCURL, Status: c.Status,
				BlockNumber: c.BlockNumber, OTSMode: c.OTSMode,
				PendingCount: c.PendingCount, TotalCreated: c.TotalCreated,
				TotalConfirmed: c.TotalConfirmed, LastProcessedBlock: c.LastProcessedBlock,
				Components: c.Components, LastAnchor: c.LastAnchor,
				Coinbase: c.Coinbase,
			}
		}
		return views, nil
	}

	// Fallback to DB
	dbNodes, err := s.db.GetAllNodeStatus(ctx)
	if err == nil && len(dbNodes) > 0 {
		views := make([]NodeStatusView, len(dbNodes))
		for i, n := range dbNodes {
			views[i] = NodeStatusView{
				Name: n.NodeName, RPCURL: n.RPCURL, Status: n.Status,
				BlockNumber: n.BlockNumber, OTSMode: n.OTSMode,
				PendingCount: n.PendingCount, TotalCreated: n.TotalCreated,
				TotalConfirmed: n.TotalConfirmed, LastProcessedBlock: n.LastProcessedBlock,
				Components: n.Components, LastAnchor: n.LastAnchor,
				Coinbase: n.Coinbase,
			}
		}
		return views, nil
	}

	// Final fallback: direct RPC (Phase 1 behavior)
	return s.fetchNodesFromRPC(ctx)
}

// GetNode tries: Redis → DB → RPC
func (s *NodeService) GetNode(ctx context.Context, name string) (*NodeStatusView, error) {
	// Try cache
	if cached, err := s.cache.GetNodeStatus(ctx, name); err == nil {
		return &NodeStatusView{
			Name: cached.Name, RPCURL: cached.RPCURL, Status: cached.Status,
			BlockNumber: cached.BlockNumber, OTSMode: cached.OTSMode,
			PendingCount: cached.PendingCount, TotalCreated: cached.TotalCreated,
			TotalConfirmed: cached.TotalConfirmed, LastProcessedBlock: cached.LastProcessedBlock,
			Components: cached.Components, LastAnchor: cached.LastAnchor,
			Coinbase: cached.Coinbase,
		}, nil
	}

	// Try DB
	if dbNode, err := s.db.GetNodeStatus(ctx, name); err == nil {
		return &NodeStatusView{
			Name: dbNode.NodeName, RPCURL: dbNode.RPCURL, Status: dbNode.Status,
			BlockNumber: dbNode.BlockNumber, OTSMode: dbNode.OTSMode,
			PendingCount: dbNode.PendingCount, TotalCreated: dbNode.TotalCreated,
			TotalConfirmed: dbNode.TotalConfirmed, LastProcessedBlock: dbNode.LastProcessedBlock,
			Components: dbNode.Components, LastAnchor: dbNode.LastAnchor,
			Coinbase: dbNode.Coinbase,
		}, nil
	}

	// RPC fallback
	for _, n := range s.cfg.Nodes {
		if n.Name == name {
			return s.fetchOneNodeFromRPC(ctx, n.Name, n.RPCURL)
		}
	}
	return nil, fmt.Errorf("node not found: %s", name)
}

func (s *NodeService) fetchNodesFromRPC(ctx context.Context) ([]NodeStatusView, error) {
	views := make([]NodeStatusView, len(s.cfg.Nodes))
	for i, n := range s.cfg.Nodes {
		v, _ := s.fetchOneNodeFromRPC(ctx, n.Name, n.RPCURL)
		if v != nil {
			views[i] = *v
		} else {
			views[i] = NodeStatusView{Name: n.Name, RPCURL: n.RPCURL, Status: "down"}
		}
	}
	return views, nil
}

func (s *NodeService) fetchOneNodeFromRPC(ctx context.Context, name, url string) (*NodeStatusView, error) {
	v := &NodeStatusView{Name: name, RPCURL: url, Status: "down"}
	if bn, err := s.rpcClient.EthBlockNumber(ctx, url); err == nil {
		v.BlockNumber = bn
	}
	if h, err := s.rpcClient.OTSHealth(ctx, url); err == nil {
		v.Status = h.Status
		v.OTSMode = h.Mode
		v.PendingCount = h.PendingCount
		v.TotalCreated = h.TotalCreated
		v.TotalConfirmed = h.TotalConfirmed
		v.LastProcessedBlock = h.LastProcessedBlock
		v.LastAnchor = h.LastAnchor
		if h.Components != nil {
			v.Components = make(map[string]interface{})
			for k, c := range h.Components {
				v.Components[k] = c
			}
		}
	}
	return v, nil
}
