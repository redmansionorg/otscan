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
	// Sync all batches by on-chain ID (metadata only, no RUID sync here)
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
	}

	// Also sync pending batches
	pending, err := idx.rpcClient.GetPendingBatches(syncCtx, url)
	if err == nil {
		for _, p := range pending {
			rec := &store.BatchRecord{
				BatchID:    p.BatchID,
				OnChainID:  int64(p.OnChainID),
				StartBlock: p.StartBlock,
				EndBlock:   p.EndBlock,
				RUIDCount:  int(p.RUIDCount),
				Status:     p.Status,
			}
			idx.db.UpsertBatch(syncCtx, rec)
		}
	}

	if synced > 0 {
		log.Printf("[indexer] batch sync: synced %d/%d batches", synced, totalBatches)
	}

	// Sync RUIDs in a separate pass with its own timeout.
	// Process one incomplete batch per cycle to avoid overwhelming RPC.
	idx.syncIncompleteRUIDs(ctx, url)
}

// syncIncompleteRUIDs finds the first batch with incomplete RUID sync
// and continues fetching from where it left off. Uses its own 120s timeout.
func (idx *Indexer) syncIncompleteRUIDs(ctx context.Context, url string) {
	ruidCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Find batches where synced RUID count < expected count
	batches, err := idx.db.ListBatchesWithIncompleteRUIDs(ruidCtx)
	if err != nil {
		log.Printf("[indexer] ruid sync: failed to list incomplete batches: %v", err)
		return
	}

	for _, b := range batches {
		if ruidCtx.Err() != nil {
			break
		}

		synced := idx.syncBatchRUIDs(ruidCtx, url, b.BatchID, uint32(b.RUIDCount), b.SyncedRUIDs)
		if synced > 0 {
			log.Printf("[indexer] ruid sync: batch %s synced %d RUIDs (%d/%d total)",
				b.BatchID, synced, b.SyncedRUIDs+synced, b.RUIDCount)
		}
	}
}

// syncBatchRUIDs fetches RUIDs for a batch starting from alreadySynced offset.
// Returns the number of newly synced RUIDs.
func (idx *Indexer) syncBatchRUIDs(ctx context.Context, url, batchID string, expectedCount uint32, alreadySynced int) int {
	const pageSize uint32 = 1000
	totalNew := 0

	for offset := uint32(alreadySynced); offset < expectedCount; offset += pageSize {
		if ctx.Err() != nil {
			break
		}

		limit := pageSize
		if offset+limit > expectedCount {
			limit = expectedCount - offset
		}
		result, err := idx.rpcClient.GetRUIDs(ctx, url, batchID, offset, limit)
		if err != nil {
			log.Printf("[indexer] ruid sync: batch %s offset %d: %v", batchID, offset, err)
			break
		}

		if len(result.RUIDs) == 0 {
			break
		}

		// Insert this page immediately (incremental progress)
		if err := idx.db.BulkInsertBatchRUIDs(ctx, batchID, result.RUIDs); err != nil {
			log.Printf("[indexer] ruid sync: batch %s insert at offset %d: %v", batchID, offset, err)
			break
		}
		totalNew += len(result.RUIDs)
	}

	return totalNew
}
