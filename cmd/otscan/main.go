package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/workshop1/otscan/internal/api"
	"github.com/workshop1/otscan/internal/cache"
	"github.com/workshop1/otscan/internal/config"
	"github.com/workshop1/otscan/internal/indexer"
	"github.com/workshop1/otscan/internal/rpc"
	"github.com/workshop1/otscan/internal/store"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("OTScan starting...\n")
	fmt.Printf("  Chain: %s (ID: %d)\n", cfg.Chain.Name, cfg.Chain.ID)
	fmt.Printf("  Nodes: %d configured\n", len(cfg.Nodes))
	for _, n := range cfg.Nodes {
		fmt.Printf("    - %s: %s\n", n.Name, n.RPCURL)
	}
	fmt.Printf("  Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpcClient := rpc.NewClient()

	// Connect to PostgreSQL
	fmt.Printf("  Connecting to PostgreSQL at %s:%d...\n", cfg.Database.Host, cfg.Database.Port)
	db, err := store.WaitForDB(ctx, &cfg.Database, 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	fmt.Printf("  PostgreSQL connected\n")

	// Connect to Redis
	fmt.Printf("  Connecting to Redis at %s...\n", cfg.Redis.Addr)
	redisCache, err := cache.WaitForRedis(&cfg.Redis, 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()
	fmt.Printf("  Redis connected\n")

	// Create API server (creates WebSocket hub)
	server := api.NewServer(cfg, rpcClient, db, redisCache)

	// Start indexer with WebSocket broadcaster
	idx := indexer.New(cfg, db, redisCache, rpcClient)
	idx.SetBroadcaster(server.Hub())
	idx.Start(ctx)
	defer idx.Stop()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("\nOTScan ready at http://%s\n", addr)
		if err := server.Router().Run(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	fmt.Println("\nShutting down...")
	cancel()
}
