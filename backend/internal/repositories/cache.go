package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache is a JSON-serialising Redis wrapper.
//
// Its defining property is that a nil *Cache — and a Cache whose client is nil,
// which is what a failed Redis connection produces — is fully usable. Every
// method degrades to "miss" or "no-op".
//
// This is the difference between the cache being an optimisation and the cache
// being a dependency. Redis holds nothing here that Postgres does not; if it is
// down, the correct behaviour is slower responses, not 500s. Making that safe
// by construction beats remembering to nil-check at forty call sites.
type Cache struct {
	client *redis.Client
	ttl    time.Duration
	log    *zap.Logger
}

// NewCache constructs a Cache. A nil client disables caching.
func NewCache(client *redis.Client, ttl time.Duration, log *zap.Logger) *Cache {
	return &Cache{client: client, ttl: ttl, log: log}
}

// Enabled reports whether caching is live.
func (c *Cache) Enabled() bool {
	return c != nil && c.client != nil
}

// Get unmarshals a cached value into dest. It returns false on a miss, on a
// deserialisation failure, or when the cache is disabled.
//
// A corrupt entry is treated as a miss and deleted rather than surfaced as an
// error: the caller can always recompute, and a poison entry that fails every
// request until its TTL expires is a far worse outcome than one slow read.
func (c *Cache) Get(ctx context.Context, key string, dest any) bool {
	if !c.Enabled() {
		return false
	}

	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			c.log.Warn("cache get failed", zap.String("key", key), zap.Error(err))
		}
		return false
	}

	if err := json.Unmarshal(raw, dest); err != nil {
		c.log.Warn("cache entry corrupt; evicting",
			zap.String("key", key), zap.Error(err))
		c.Delete(ctx, key)
		return false
	}
	return true
}

// Set stores a value under key with the configured TTL.
func (c *Cache) Set(ctx context.Context, key string, value any) {
	c.SetTTL(ctx, key, value, c.ttl)
}

// SetTTL stores a value with an explicit TTL.
func (c *Cache) SetTTL(ctx context.Context, key string, value any, ttl time.Duration) {
	if !c.Enabled() {
		return
	}

	raw, err := json.Marshal(value)
	if err != nil {
		c.log.Warn("cache marshal failed", zap.String("key", key), zap.Error(err))
		return
	}

	if err := c.client.Set(ctx, key, raw, ttl).Err(); err != nil {
		// A write failure is logged and swallowed. The caller already has the
		// value; failing their request because we could not memoise it would
		// be absurd.
		c.log.Warn("cache set failed", zap.String("key", key), zap.Error(err))
	}
}

// Delete removes keys.
func (c *Cache) Delete(ctx context.Context, keys ...string) {
	if !c.Enabled() || len(keys) == 0 {
		return
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.log.Warn("cache delete failed", zap.Strings("keys", keys), zap.Error(err))
	}
}

// InvalidatePrefix removes every key under a prefix.
//
// It uses SCAN, never KEYS. KEYS is O(N) over the entire keyspace and blocks
// the single-threaded Redis event loop for its whole duration — on a shared
// instance it is an outage waiting for a big enough keyspace. SCAN is
// incremental and cursor-based.
func (c *Cache) InvalidatePrefix(ctx context.Context, prefix string) {
	if !c.Enabled() {
		return
	}

	var (
		cursor  uint64
		removed int
	)
	pattern := prefix + "*"

	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			c.log.Warn("cache scan failed", zap.String("prefix", prefix), zap.Error(err))
			return
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				c.log.Warn("cache bulk delete failed", zap.Error(err))
				return
			}
			removed += len(keys)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}

	if removed > 0 {
		c.log.Debug("cache prefix invalidated",
			zap.String("prefix", prefix), zap.Int("removed", removed))
	}
}

// Cache key namespaces. Grouping by prefix is what makes targeted invalidation
// possible: a catalogue write can drop every product and category key without
// touching anything else.
const (
	KeyPrefixProducts        = "products:"
	KeyPrefixProduct         = "product:"
	KeyPrefixCategories      = "categories:"
	KeyPrefixRecommendations = "recs:"
	KeyPrefixStats           = "stats:"
)

// ProductKey builds the cache key for a single product.
func ProductKey(id string) string { return KeyPrefixProduct + id }

// CategoriesKey is the cache key for the category list.
func CategoriesKey() string { return KeyPrefixCategories + "all" }

// StatsKey is the cache key for the dashboard statistics.
func StatsKey() string { return KeyPrefixStats + "dashboard" }

// RecommendationKey builds the cache key for a stored recommendation.
func RecommendationKey(id string) string { return KeyPrefixRecommendations + id }

// Ping reports cache health.
func (c *Cache) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return fmt.Errorf("cache: disabled")
	}
	return c.client.Ping(ctx).Err()
}
