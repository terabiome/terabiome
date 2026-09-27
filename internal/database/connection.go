package database

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/terabiome/terabiome/internal/config"
)

type ConnectionType string

const (
	ConnectionTypePostgresql ConnectionType = "postgresql"
)

var pool *pgxpool.Pool

func Init() {
	mu := sync.Mutex{}
	mu.Lock()

	defer mu.Unlock()

	cfg := config.Get().Database

	// only support postgres
	switch cfg.Type {
	case string(ConnectionTypePostgresql):
	default:
		log.Fatalf("unsupported connection type: %v", cfg.Type)
	}
	log.Printf("setting up connection type %s\n", cfg.Type)

	// start connecting
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		log.Fatalf("failed to parse config for PGX pool: %v", err)
	}
	// set default values (can set via config later)
	poolCfg.MinConns = 1
	poolCfg.MaxConns = 2
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	ctx := context.Background()
	pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("failed to establish PGX pool connection: %v", err)
	}
	defer pool.Close()

	// validate connection
	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("failed to validate PGX pool connection: %v", err)
	}
}

func GetPool() *pgxpool.Pool {
	// allow nil-reference panic attacks
	return pool
}
