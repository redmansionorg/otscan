package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/store"
)

const (
	// Event signatures (keccak256)
	// Claimed(bytes32 indexed ruid, address indexed claimant, uint64 submitBlock)
	claimedEventSig = "0x4f0c27e2232052d4dd3cb5e62ed8a6101f2c7c03c8667c45c3863b68be216519"
	// Published(bytes32 indexed ruid, bytes32 indexed auid, bytes32 indexed puid, address claimant)
	publishedEventSig = "0x4e9c366c4e50eb205bc3bcd313554d8c7a77377814e5995e79ec9daf3a747801"

	copyrightRegistryAddr = "0x0000000000000000000000000000000000009000"

	claimSyncBatchSize   = 2000  // blocks per eth_getLogs call
	syncMetaKeyClaimSync = "last_claim_sync_block"
)

func (idx *Indexer) claimSyncerLoop(ctx context.Context) {
	time.Sleep(5 * time.Second)
	idx.syncClaims(ctx)

	interval := idx.cfg.Indexer.ClaimSyncIntervalD
	if interval == 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			idx.syncClaims(ctx)
		}
	}
}

func (idx *Indexer) syncClaims(ctx context.Context) {
	syncCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	url := idx.pickNode()
	if url == "" {
		return
	}

	// Get last synced block from DB
	lastSynced := uint64(0)
	if val, err := idx.db.GetSyncMeta(syncCtx, syncMetaKeyClaimSync); err == nil && val != "" {
		if n, err := strconv.ParseUint(val, 10, 64); err == nil {
			lastSynced = n
		}
	}

	// Get latest block number
	latestBlock, err := idx.rpcClient.EthBlockNumber(syncCtx, url)
	if err != nil {
		log.Printf("[indexer] claim sync: failed to get block number: %v", err)
		return
	}

	if lastSynced >= latestBlock {
		return
	}

	totalSynced := 0
	fromBlock := lastSynced + 1

	for fromBlock <= latestBlock {
		toBlock := fromBlock + claimSyncBatchSize - 1
		if toBlock > latestBlock {
			toBlock = latestBlock
		}

		claimed, published, err := idx.fetchClaimEvents(syncCtx, url, fromBlock, toBlock)
		if err != nil {
			log.Printf("[indexer] claim sync: failed to fetch events %d-%d: %v", fromBlock, toBlock, err)
			break
		}

		// Process Claimed events
		for _, c := range claimed {
			if err := idx.db.UpsertClaim(syncCtx, c); err != nil {
				log.Printf("[indexer] claim sync: failed to upsert claim %s: %v", c.RUID, err)
			}
		}

		// Process Published events
		for _, p := range published {
			if err := idx.db.UpsertClaim(syncCtx, p); err != nil {
				log.Printf("[indexer] claim sync: failed to upsert publish %s: %v", p.RUID, err)
			}
		}

		totalSynced += len(claimed) + len(published)

		// Update sync progress
		if err := idx.db.SetSyncMeta(syncCtx, syncMetaKeyClaimSync, fmt.Sprintf("%d", toBlock)); err != nil {
			log.Printf("[indexer] claim sync: failed to update sync meta: %v", err)
			break
		}

		fromBlock = toBlock + 1
	}

	if totalSynced > 0 {
		log.Printf("[indexer] claim sync: synced %d events (blocks %d-%d)", totalSynced, lastSynced+1, latestBlock)
		idx.broadcast("claim_sync", map[string]interface{}{
			"synced":    totalSynced,
			"lastBlock": latestBlock,
		})
	}
}

func (idx *Indexer) fetchClaimEvents(ctx context.Context, url string, fromBlock, toBlock uint64) ([]*store.ClaimRecord, []*store.ClaimRecord, error) {
	// Fetch Claimed events
	claimedLogs, err := idx.rpcClient.GetLogs(ctx, url, fromBlock, toBlock, copyrightRegistryAddr, []string{claimedEventSig})
	if err != nil {
		return nil, nil, fmt.Errorf("get claimed logs: %w", err)
	}

	// Fetch Published events
	publishedLogs, err := idx.rpcClient.GetLogs(ctx, url, fromBlock, toBlock, copyrightRegistryAddr, []string{publishedEventSig})
	if err != nil {
		return nil, nil, fmt.Errorf("get published logs: %w", err)
	}

	var claimed []*store.ClaimRecord
	for _, entry := range claimedLogs {
		rec := parseClaimedLog(&entry)
		if rec != nil {
			claimed = append(claimed, rec)
		}
	}

	var published []*store.ClaimRecord
	for _, entry := range publishedLogs {
		rec := parsePublishedLog(&entry)
		if rec != nil {
			published = append(published, rec)
		}
	}

	return claimed, published, nil
}

// parseClaimedLog parses a Claimed event log entry.
// Claimed(bytes32 indexed ruid, address indexed claimant, uint64 submitBlock)
// Topics[0] = event sig, Topics[1] = ruid, Topics[2] = claimant (padded)
// Data = submitBlock (uint64 padded to 32 bytes)
func parseClaimedLog(entry *rpc.LogEntry) *store.ClaimRecord {
	if len(entry.Topics) < 3 {
		return nil
	}

	ruid := entry.Topics[1]
	claimant := "0x" + entry.Topics[2][26:] // last 20 bytes of 32-byte padded address

	blockNum := parseHexUint64(entry.BlockNumber)

	// Parse submitBlock from data
	submitBlock := blockNum
	data := strings.TrimPrefix(entry.Data, "0x")
	if len(data) >= 64 {
		if n, ok := new(big.Int).SetString(data[:64], 16); ok {
			submitBlock = n.Uint64()
		}
	}

	return &store.ClaimRecord{
		RUID:        ruid,
		Claimant:    claimant,
		SubmitBlock: submitBlock,
		SubmitTime:  blockNum, // use block number as timestamp proxy
	}
}

// parsePublishedLog parses a Published event log entry.
// Published(bytes32 indexed ruid, bytes32 indexed auid, bytes32 indexed puid, address claimant)
// Topics[0] = event sig, Topics[1] = ruid, Topics[2] = auid, Topics[3] = puid
// Data = claimant (address padded to 32 bytes)
func parsePublishedLog(entry *rpc.LogEntry) *store.ClaimRecord {
	if len(entry.Topics) < 4 {
		return nil
	}

	ruid := entry.Topics[1]
	auid := entry.Topics[2]
	puid := entry.Topics[3]

	var claimant string
	data := strings.TrimPrefix(entry.Data, "0x")
	if len(data) >= 64 {
		claimant = "0x" + data[24:64] // last 20 bytes
	}

	blockNum := parseHexUint64(entry.BlockNumber)

	return &store.ClaimRecord{
		RUID:         ruid,
		Claimant:     claimant,
		Published:    true,
		AUID:         auid,
		PUID:         puid,
		PublishBlock: blockNum,
		PublishTime:  blockNum,
	}
}

func parseHexUint64(hex string) uint64 {
	hex = strings.TrimPrefix(hex, "0x")
	n, _ := strconv.ParseUint(hex, 16, 64)
	return n
}
