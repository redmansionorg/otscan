package indexer

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/workshop1/otscan/internal/cache"
)

func (idx *Indexer) pollNodes(ctx context.Context) {
	var wg sync.WaitGroup
	for _, nodeCfg := range idx.cfg.Nodes {
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			idx.pollOneNode(ctx, name, url)
		}(nodeCfg.Name, nodeCfg.RPCURL)
	}
	wg.Wait()
}

func (idx *Indexer) pollOneNode(ctx context.Context, name, url string) {
	pollCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	status := "down"
	var blockNum uint64
	var otsMode string
	var pendingCount, totalCreated, totalConfirmed int
	var lastProcessedBlock uint64
	var components map[string]interface{}
	var lastAnchor *string
	var coinbase string

	// Get block number
	if bn, err := idx.rpcClient.EthBlockNumber(pollCtx, url); err == nil {
		blockNum = bn
	}

	// Get coinbase (mining address)
	if addr, err := idx.rpcClient.EthCoinbase(pollCtx, url); err == nil {
		coinbase = addr
	}

	// Get OTS health
	if health, err := idx.rpcClient.OTSHealth(pollCtx, url); err == nil {
		status = health.Status
		otsMode = health.Mode
		pendingCount = health.PendingCount
		totalCreated = health.TotalCreated
		totalConfirmed = health.TotalConfirmed
		lastProcessedBlock = health.LastProcessedBlock
		lastAnchor = health.LastAnchor

		// Convert typed components to generic map for storage
		if health.Components != nil {
			compJSON, _ := json.Marshal(health.Components)
			json.Unmarshal(compJSON, &components)
		}
	}

	// Get node DB ID
	idx.mu.RLock()
	nodeID, ok := idx.nodeIDs[name]
	idx.mu.RUnlock()
	if !ok {
		return
	}

	// Update DB
	if err := idx.db.UpsertNodeStatus(pollCtx, nodeID, status, blockNum,
		otsMode, pendingCount, totalCreated, totalConfirmed,
		lastProcessedBlock, components, lastAnchor, coinbase); err != nil {
		log.Printf("[indexer] failed to update node status %s: %v", name, err)
	}

	// Insert history (sample every poll)
	idx.db.InsertNodeStatusHistory(pollCtx, nodeID, blockNum, pendingCount, status)

	// Update Redis cache
	cached := &cache.CachedNodeStatus{
		Name:               name,
		RPCURL:             url,
		Status:             status,
		BlockNumber:        blockNum,
		OTSMode:            otsMode,
		PendingCount:       pendingCount,
		TotalCreated:       totalCreated,
		TotalConfirmed:     totalConfirmed,
		LastProcessedBlock: lastProcessedBlock,
		Components:         components,
		LastAnchor:         lastAnchor,
		Coinbase:           coinbase,
		UpdatedAt:          time.Now(),
	}
	if err := idx.cache.SetNodeStatus(pollCtx, cached); err != nil {
		log.Printf("[indexer] failed to cache node status %s: %v", name, err)
	}

	// Broadcast node status update via WebSocket
	idx.broadcast("node_status", map[string]interface{}{
		"name":               name,
		"status":             status,
		"blockNumber":        blockNum,
		"otsMode":            otsMode,
		"pendingCount":       pendingCount,
		"totalCreated":       totalCreated,
		"totalConfirmed":     totalConfirmed,
		"lastProcessedBlock": lastProcessedBlock,
	})
}
