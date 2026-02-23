package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/workshop1/otscan/internal/config"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(ctx context.Context, cfg *config.DatabaseConfig) (*DB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	db := &DB{Pool: pool}
	if err := db.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) migrate(ctx context.Context) error {
	migrationFiles := []string{
		"001_create_tables.sql",
		"002_add_conflict_indexes.sql",
	}
	dirs := []string{"migrations", "/app/migrations"}

	for _, file := range migrationFiles {
		var sql []byte
		var err error
		for _, dir := range dirs {
			sql, err = os.ReadFile(dir + "/" + file)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Printf("[store] migration %s not found, skipping", file)
			continue
		}
		if _, err := db.Pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec migration %s: %w", file, err)
		}
	}
	log.Printf("[store] migrations applied successfully")
	return nil
}

// WaitForDB retries connection until success or timeout.
func WaitForDB(ctx context.Context, cfg *config.DatabaseConfig, timeout time.Duration) (*DB, error) {
	deadline := time.Now().Add(timeout)
	for {
		db, err := NewDB(ctx, cfg)
		if err == nil {
			return db, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for database: %w", err)
		}
		log.Printf("[store] waiting for database... (%v)", err)
		time.Sleep(2 * time.Second)
	}
}
