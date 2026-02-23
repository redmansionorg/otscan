package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Chain    ChainConfig    `yaml:"chain"`
	Nodes    []NodeConfig   `yaml:"nodes"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Indexer  IndexerConfig  `yaml:"indexer"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type ChainConfig struct {
	ID                    int    `yaml:"id"`
	Name                  string `yaml:"name"`
	BreatheBlockInterval  int    `yaml:"breatheBlockInterval"`
}

type NodeConfig struct {
	Name   string `yaml:"name"`
	RPCURL string `yaml:"rpcUrl"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	Addr           string        `yaml:"addr"`
	Password       string        `yaml:"password"`
	DB             int           `yaml:"db"`
	NodeStatusTTL  string        `yaml:"nodeStatusTTL"`
	NodeStatusTTLD time.Duration `yaml:"-"`
}

type IndexerConfig struct {
	NodePollingInterval  string        `yaml:"nodePollingInterval"`
	BatchSyncInterval    string        `yaml:"batchSyncInterval"`
	ClaimSyncInterval    string        `yaml:"claimSyncInterval"`
	NodePollingIntervalD time.Duration `yaml:"-"`
	BatchSyncIntervalD   time.Duration `yaml:"-"`
	ClaimSyncIntervalD   time.Duration `yaml:"-"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 3000
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if len(cfg.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes configured")
	}

	// Parse durations
	if cfg.Redis.NodeStatusTTL != "" {
		cfg.Redis.NodeStatusTTLD, _ = time.ParseDuration(cfg.Redis.NodeStatusTTL)
	}
	if cfg.Redis.NodeStatusTTLD == 0 {
		cfg.Redis.NodeStatusTTLD = 30 * time.Second
	}
	if cfg.Indexer.NodePollingInterval != "" {
		cfg.Indexer.NodePollingIntervalD, _ = time.ParseDuration(cfg.Indexer.NodePollingInterval)
	}
	if cfg.Indexer.NodePollingIntervalD == 0 {
		cfg.Indexer.NodePollingIntervalD = 10 * time.Second
	}
	if cfg.Indexer.BatchSyncInterval != "" {
		cfg.Indexer.BatchSyncIntervalD, _ = time.ParseDuration(cfg.Indexer.BatchSyncInterval)
	}
	if cfg.Indexer.BatchSyncIntervalD == 0 {
		cfg.Indexer.BatchSyncIntervalD = 30 * time.Second
	}
	if cfg.Indexer.ClaimSyncInterval != "" {
		cfg.Indexer.ClaimSyncIntervalD, _ = time.ParseDuration(cfg.Indexer.ClaimSyncInterval)
	}
	if cfg.Indexer.ClaimSyncIntervalD == 0 {
		cfg.Indexer.ClaimSyncIntervalD = 30 * time.Second
	}

	return &cfg, nil
}

func (c *DatabaseConfig) DSN() string {
	sslmode := "disable"
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, sslmode)
}
