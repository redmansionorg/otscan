package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/store"
)

type DashboardResponse struct {
	ChainID         int                      `json:"chainId"`
	ChainName       string                   `json:"chainName"`
	NodeCount       int                      `json:"nodeCount"`
	NodesHealthy    int                      `json:"nodesHealthy"`
	LatestBlock     uint64                   `json:"latestBlock"`
	TotalBatches    int                      `json:"totalBatches"`
	PendingBatches  int                      `json:"pendingBatches"`
	AnchoredBatches int                      `json:"anchoredBatches"`
	TotalClaims     int                      `json:"totalClaims"`
	Nodes           []map[string]interface{} `json:"nodes"`
	RecentBatches   []*rpc.BatchSummary      `json:"recentBatches,omitempty"`
	RecentClaims    []store.ClaimRecord      `json:"recentClaims,omitempty"`
}

func (s *Server) handleDashboard(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	// Get nodes from service layer
	nodes, err := s.nodeSvc.GetAllNodes(ctx)
	if err != nil {
		errJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	var maxBlock uint64
	healthy := 0
	nodeItems := make([]map[string]interface{}, len(nodes))
	for i, n := range nodes {
		if n.Status == "healthy" {
			healthy++
		}
		if n.BlockNumber > maxBlock {
			maxBlock = n.BlockNumber
		}
		nodeItems[i] = map[string]interface{}{
			"name":               n.Name,
			"rpcUrl":             n.RPCURL,
			"status":             n.Status,
			"blockNumber":        n.BlockNumber,
			"otsMode":            n.OTSMode,
			"pendingCount":       n.PendingCount,
			"totalCreated":       n.TotalCreated,
			"totalConfirmed":     n.TotalConfirmed,
			"lastProcessedBlock": n.LastProcessedBlock,
			"components":         n.Components,
			"lastAnchor":         n.LastAnchor,
		}
	}

	// Get batch stats from database (more reliable than RPC which may timeout)
	var totalBatches, pendingBatches int
	batchCtx, batchCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if count, err := s.db.GetBatchCount(batchCtx); err == nil {
		totalBatches = count
	}
	if count, err := s.db.GetPendingBatchCount(batchCtx); err == nil {
		pendingBatches = count
	}
	batchCancel()

	// Fallback to RPC if DB returns 0
	if totalBatches == 0 {
		status, err := s.rpcClient.OTSStatus(ctx, s.pickNode())
		if err == nil && status.Stats != nil {
			totalBatches = status.Stats.Total
			if totalBatches == 0 {
				totalBatches = status.Stats.TotalBatches
			}
			pendingBatches = status.PendingCount
		}
	}

	// Get total claims from claims table (includes unbatched)
	// Use a fresh context to avoid timeout from previous operations
	claimsCtx, claimsCancel := context.WithTimeout(context.Background(), 5*time.Second)
	totalClaims, _ := s.db.GetClaimCount(claimsCtx)
	claimsCancel()

	// Get anchored batch count
	anchoredBatches := totalBatches - pendingBatches

	// Get recent batches from DB
	var recentBatches []*rpc.BatchSummary
	if result, err := s.batchSvc.ListBatches(ctx, "", 1, 10); err == nil {
		for _, b := range result.Batches {
			recentBatches = append(recentBatches, &rpc.BatchSummary{
				BatchID:     b.BatchID,
				OnChainID:   uint64(b.OnChainID),
				StartBlock:  b.StartBlock,
				EndBlock:    b.EndBlock,
				RUIDCount:   uint32(b.RUIDCount),
				Status:      b.Status,
				AnchoredBy:  b.AnchoredBy,
				AnchorBlock: b.AnchorBlock,
			})
		}
	}

	// Get recent claims from DB (latest 5)
	recentClaims, _, _ := s.db.ListClaims(ctx, "", 0, 5)

	okJSON(c, DashboardResponse{
		ChainID:         s.cfg.Chain.ID,
		ChainName:       s.cfg.Chain.Name,
		NodeCount:       len(s.cfg.Nodes),
		NodesHealthy:    healthy,
		LatestBlock:     maxBlock,
		TotalBatches:    totalBatches,
		PendingBatches:  pendingBatches,
		AnchoredBatches: anchoredBatches,
		TotalClaims:     totalClaims,
		Nodes:           nodeItems,
		RecentBatches:   recentBatches,
		RecentClaims:    recentClaims,
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleConfig(c *gin.Context) {
	nodeNames := make([]string, len(s.cfg.Nodes))
	for i, n := range s.cfg.Nodes {
		nodeNames[i] = n.Name
	}
	okJSON(c, gin.H{
		"chainId":              s.cfg.Chain.ID,
		"chainName":            s.cfg.Chain.Name,
		"breatheBlockInterval": s.cfg.Chain.BreatheBlockInterval,
		"nodeCount":            len(s.cfg.Nodes),
		"nodes":                nodeNames,
	})
}
