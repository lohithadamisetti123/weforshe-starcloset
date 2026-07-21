package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// limiterStore hands out one token bucket per client key, creating it on first
// sight and reaping idle ones so the map does not grow for the life of the
// process.
//
// Keying by client rather than sharing one global bucket is what makes the
// limiter fair: one enthusiastic shopper's assistant session should not eat
// into another's budget on a shared token pool.
type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*bucket
	rps      rate.Limit
	burst    int
	idleTTL  time.Duration
}

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newLimiterStore(rps float64, burst int) *limiterStore {
	s := &limiterStore{
		limiters: make(map[string]*bucket),
		rps:      rate.Limit(rps),
		burst:    burst,
		idleTTL:  10 * time.Minute,
	}
	go s.reapLoop()
	return s
}

func (s *limiterStore) get(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.limiters[key]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(s.rps, s.burst)}
		s.limiters[key] = b
	}
	b.lastSeen = time.Now()
	return b.limiter
}

// reapLoop evicts buckets nobody has used recently.
//
// Without this, a limiter keyed by guest session ID — which is effectively
// unbounded cardinality, one per browser that has ever visited — would leak
// memory for the process lifetime. The loop is deliberately dumb: wake up,
// sweep, sleep. Nothing about correctness depends on its timing.
func (s *limiterStore) reapLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-s.idleTTL)
		s.mu.Lock()
		for key, b := range s.limiters {
			if b.lastSeen.Before(cutoff) {
				delete(s.limiters, key)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimit builds a token-bucket limiter middleware.
//
// The client key prefers the authenticated user ID over the raw IP: behind a
// shared office NAT or a mobile carrier's CGNAT, many genuine shoppers share
// one IP, and rate-limiting by IP there punishes all of them for one person's
// traffic. The IP is still the fallback for the pre-auth case, which for this
// API is effectively never — see middleware.Auth, which mints a guest identity
// before any rate-limited route is reached.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	store := newLimiterStore(rps, burst)

	return func(c *gin.Context) {
		key := UserIDString(c)
		if key == "" {
			key = c.ClientIP()
		}

		limiter := store.get(key)
		if !limiter.Allow() {
			c.Header("Retry-After", "1")
			c.Header("X-RateLimit-Limit", strconv.Itoa(burst))
			c.Header("X-RateLimit-Remaining", "0")
			utils.Fail(c, &utils.AppError{
				Status:  http.StatusTooManyRequests,
				Code:    utils.CodeRateLimited,
				Message: "Too many requests. Please slow down and try again shortly",
			})
			return
		}
		c.Next()
	}
}
