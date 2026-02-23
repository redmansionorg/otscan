package indexer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/workshop1/otscan/internal/cache"
	"github.com/workshop1/otscan/internal/config"
	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/store"
)

// EventBroadcaster is an interface for pushing live events (e.g. WebSocket hub).
type EventBroadcaster interface {
	BroadcastEvent(eventType string, data interface{})
}

type Indexer struct {
	cfg         *config.Config
	db          *store.DB
	cache       *cache.Cache
	rpcClient   *rpc.Client
	broadcaster EventBroadcaster
	nodeIDs     map[string]int // name -> db id
	lastBatchStatus map[string]string // batchID -> last known status
	mu          sync.RWMutex
	cancel      context.CancelFunc
}

func New(cfg *config.Config, db *store.DB, c *cache.Cache, rpcClient *rpc.Client) *Indexer {
	return &Indexer{
		cfg:             cfg,
		db:              db,
		cache:           c,
		rpcClient:       rpcClient,
		nodeIDs:         make(map[string]int),
		lastBatchStatus: make(map[string]string),
	}
}

// SetBroadcaster sets the event broadcaster (WebSocket hub).
func (idx *Indexer) SetBroadcaster(b EventBroadcaster) {
	idx.broadcaster = b
}

func (idx *Indexer) broadcast(eventType string, data interface{}) {
	if idx.broadcaster != nil {
		idx.broadcaster.BroadcastEvent(eventType, data)
	}
}

func (idx *Indexer) Start(ctx context.Context) {
	ctx, idx.cancel = context.WithCancel(ctx)

	// Register nodes in DB
	for _, n := range idx.cfg.Nodes {
		id, err := idx.db.UpsertNode(ctx, n.Name, n.RPCURL)
		if err != nil {
			log.Printf("[indexer] failed to register node %s: %v", n.Name, err)
			continue
		}
		idx.mu.Lock()
		idx.nodeIDs[n.Name] = id
		idx.mu.Unlock()
		log.Printf("[indexer] registered node %s (id=%d)", n.Name, id)
	}

	// Start pollers
	go idx.nodePollerLoop(ctx)
	go idx.batchSyncerLoop(ctx)
	go idx.claimSyncerLoop(ctx)

	log.Printf("[indexer] started (node poll=%s, batch sync=%s, claim sync=%s)",
		idx.cfg.Indexer.NodePollingIntervalD, idx.cfg.Indexer.BatchSyncIntervalD, idx.cfg.Indexer.ClaimSyncIntervalD)
}

func (idx *Indexer) Stop() {
	if idx.cancel != nil {
		idx.cancel()
	}
}

func (idx *Indexer) nodePollerLoop(ctx context.Context) {
	// Run immediately, then on interval
	idx.pollNodes(ctx)
	ticker := time.NewTicker(idx.cfg.Indexer.NodePollingIntervalD)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			idx.pollNodes(ctx)
		}
	}
}

func (idx *Indexer) batchSyncerLoop(ctx context.Context) {
	// Wait a bit for first node poll to complete
	time.Sleep(3 * time.Second)
	idx.syncBatches(ctx)
	ticker := time.NewTicker(idx.cfg.Indexer.BatchSyncIntervalD)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			idx.syncBatches(ctx)
		}
	}
}

func (idx *Indexer) pickNode() string {
	if len(idx.cfg.Nodes) > 0 {
		return idx.cfg.Nodes[0].RPCURL
	}
	return ""
}
