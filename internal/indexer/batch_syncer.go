package indexer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/workshop1/otscan/internal/store"
)

func (idx *Indexer) syncBatches(ctx context.Context) {
	syncCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := idx.pickNode()
	if url == "" {
		return
	}

	// Get total batch count from status
	status, err := idx.rpcClient.OTSStatus(syncCtx, url)
	if err != nil {
		log.Printf("[indexer] batch sync: failed to get status: %v", err)
		return
	}

	totalBatches := 0
	if status.Stats != nil {
		totalBatches = status.Stats.TotalBatches
		if totalBatches == 0 {
			totalBatches = status.Stats.Total
		}
	}

	if totalBatches == 0 {
		return
	}

	synced := 0
	// Sync all batches by on-chain ID
	for i := 1; i <= totalBatches; i++ {
		batch, err := idx.rpcClient.GetBatch(syncCtx, url, fmt.Sprintf("%d", i))
		if err != nil {
			continue
		}

		rec := &store.BatchRecord{
			BatchID:        batch.BatchID,
			OnChainID:      int64(batch.OnChainID),
			StartBlock:     batch.StartBlock,
			EndBlock:       batch.EndBlock,
			RootHash:       batch.RootHash,
			OTSDigest:      batch.OTSDigest,
			RUIDCount:      int(batch.RUIDCount),
			TriggerType:    batch.TriggerType,
			Status:         batch.Status,
			BTCTxID:        batch.BTCTxID,
			BTCBlockHeight: batch.BTCBlockHeight,
			BTCTimestamp:   batch.BTCTimestamp,
			CreatedAt:      batch.CreatedAt,
			AnchoredBy:     batch.AnchoredBy,
			AnchorBlock:    batch.AnchorBlock,
		}
		if err := idx.db.UpsertBatch(syncCtx, rec); err != nil {
			log.Printf("[indexer] batch sync: failed to upsert batch %s: %v", batch.BatchID, err)
			continue
		}
		synced++

		// Check for status change and broadcast
		idx.mu.RLock()
		prevStatus := idx.lastBatchStatus[batch.BatchID]
		idx.mu.RUnlock()
		if prevStatus != batch.Status {
			idx.mu.Lock()
			idx.lastBatchStatus[batch.BatchID] = batch.Status
			idx.mu.Unlock()
			if prevStatus != "" { // Don't broadcast on first sync
				idx.broadcast("batch_update", map[string]interface{}{
					"batchID":   batch.BatchID,
					"onChainID": batch.OnChainID,
					"status":    batch.Status,
					"oldStatus": prevStatus,
				})
				log.Printf("[indexer] batch %s status changed: %s -> %s", batch.BatchID, prevStatus, batch.Status)
			}
		}

		// Sync RUIDs for this batch if it has any
		if batch.RUIDCount > 0 {
			idx.syncBatchRUIDs(syncCtx, url, batch.BatchID, batch.RUIDCount)
		}
	}

	// Also sync pending batches
	pending, err := idx.rpcClient.GetPendingBatches(syncCtx, url)
	if err == nil {
		for _, p := range pending {
			rec := &store.BatchRecord{
				BatchID:   p.BatchID,
				OnChainID: int64(p.OnChainID),
				StartBlock: p.StartBlock,
				EndBlock:  p.EndBlock,
				RUIDCount: int(p.RUIDCount),
				Status:    p.Status,
			}
			idx.db.UpsertBatch(syncCtx, rec)
		}
	}

	if synced > 0 {
		log.Printf("[indexer] batch sync: synced %d/%d batches", synced, totalBatches)
	}
}

func (idx *Indexer) syncBatchRUIDs(ctx context.Context, url, batchID string, count uint32) {
	const pageSize uint32 = 100
	var allRUIDs []string

	for offset := uint32(0); offset < count; offset += pageSize {
		limit := pageSize
		if offset+limit > count {
			limit = count - offset
		}
		result, err := idx.rpcClient.GetRUIDs(ctx, url, batchID, offset, limit)
		if err != nil {
			break
		}
		allRUIDs = append(allRUIDs, result.RUIDs...)
	}

	if len(allRUIDs) > 0 {
		if err := idx.db.UpsertBatchRUIDs(ctx, batchID, allRUIDs); err != nil {
			log.Printf("[indexer] failed to sync RUIDs for batch %s: %v", batchID, err)
		}
	}
}
