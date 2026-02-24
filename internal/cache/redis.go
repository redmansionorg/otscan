package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/workshop1/otscan/internal/config"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(cfg *config.RedisConfig) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Printf("[cache] connected to Redis at %s", cfg.Addr)
	return &Cache{client: client, ttl: cfg.NodeStatusTTLD}, nil
}

func (c *Cache) Close() {
	c.client.Close()
}

// Node status cache

type CachedNodeStatus struct {
	Name               string                 `json:"name"`
	RPCURL             string                 `json:"rpcUrl"`
	Status             string                 `json:"status"`
	BlockNumber        uint64                 `json:"blockNumber"`
	OTSMode            string                 `json:"otsMode"`
	PendingCount       int                    `json:"pendingCount"`
	TotalCreated       int                    `json:"totalCreated"`
	TotalConfirmed     int                    `json:"totalConfirmed"`
	LastProcessedBlock uint64                 `json:"lastProcessedBlock"`
	Components         map[string]interface{} `json:"components,omitempty"`
	LastAnchor         *string                `json:"lastAnchor,omitempty"`
	Coinbase           string                 `json:"coinbase,omitempty"`
	UpdatedAt          time.Time              `json:"updatedAt"`
}

func nodeKey(name string) string {
	return fmt.Sprintf("otscan:node:%s", name)
}

func (c *Cache) SetNodeStatus(ctx context.Context, status *CachedNodeStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, nodeKey(status.Name), data, c.ttl).Err()
}

func (c *Cache) GetNodeStatus(ctx context.Context, name string) (*CachedNodeStatus, error) {
	data, err := c.client.Get(ctx, nodeKey(name)).Bytes()
	if err != nil {
		return nil, err
	}
	var status CachedNodeStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *Cache) GetAllNodeStatuses(ctx context.Context, nodeNames []string) ([]*CachedNodeStatus, error) {
	var results []*CachedNodeStatus
	for _, name := range nodeNames {
		s, err := c.GetNodeStatus(ctx, name)
		if err != nil {
			continue
		}
		results = append(results, s)
	}
	return results, nil
}

// Dashboard stats cache

func (c *Cache) SetDashboardStats(ctx context.Context, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "otscan:dashboard", b, c.ttl).Err()
}

func (c *Cache) GetDashboardStats(ctx context.Context) (json.RawMessage, error) {
	data, err := c.client.Get(ctx, "otscan:dashboard").Bytes()
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// WaitForRedis retries until connected or timeout.
func WaitForRedis(cfg *config.RedisConfig, timeout time.Duration) (*Cache, error) {
	deadline := time.Now().Add(timeout)
	for {
		c, err := New(cfg)
		if err == nil {
			return c, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for redis: %w", err)
		}
		log.Printf("[cache] waiting for Redis... (%v)", err)
		time.Sleep(2 * time.Second)
	}
}
