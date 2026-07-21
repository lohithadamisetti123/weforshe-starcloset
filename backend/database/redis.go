package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/myntra/bharat-ai-backend/internal/config"
)

// NewRedis opens a Redis client.
//
// Redis is a cache here, never a store of record, so a failure to connect
// returns (nil, nil) with a warning rather than an error. Refusing to serve the
// catalogue because a *cache* is down would convert a latency problem into an
// outage — the repositories all treat a nil client as "cache disabled" and read
// straight through to Postgres.
func NewRedis(ctx context.Context, cfg config.RedisConfig, log *zap.Logger) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("database: parsing REDIS_URL: %w", err)
	}

	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.PoolSize = 20
	opts.MinIdleConns = 2

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		log.Warn("redis unreachable; continuing with caching disabled",
			zap.Error(err), zap.String("url", redactRedisURL(cfg.URL)))
		_ = client.Close()
		return nil, nil
	}

	log.Info("redis connected", zap.Duration("cacheTTL", cfg.CacheTTL))
	return client, nil
}

// CloseRedis releases the client.
func CloseRedis(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}

// redactRedisURL strips credentials so a connection string never reaches the
// logs intact.
func redactRedisURL(url string) string {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return "invalid-url"
	}
	return fmt.Sprintf("redis://%s/%d", opts.Addr, opts.DB)
}
