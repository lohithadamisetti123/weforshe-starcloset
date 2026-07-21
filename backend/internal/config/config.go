// Package config loads and validates all runtime configuration.
//
// Configuration is environment-first (12-factor): Viper reads an optional .env
// file for local development and is then overlaid by real environment
// variables, which always win. Nothing else in the codebase reads os.Getenv —
// every component receives the typed struct it needs via dependency injection.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the fully-resolved application configuration.
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	JWT            JWTConfig
	Gemini         GeminiConfig
	CORS           CORSConfig
	RateLimit      RateLimitConfig
	Recommendation RecommendationConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port           string
	Env            string
	LogLevel       string
	RequestTimeout time.Duration
	AllowGuest     bool
}

// IsProduction reports whether the service runs with production semantics
// (JSON logs, no stack traces on the wire, Gin release mode).
func (s ServerConfig) IsProduction() bool {
	return strings.EqualFold(s.Env, "production")
}

// DatabaseConfig holds PostgreSQL settings.
type DatabaseConfig struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
	AutoMigrate  bool
	AutoSeed     bool
}

// RedisConfig holds Redis settings.
type RedisConfig struct {
	URL      string
	CacheTTL time.Duration
}

// JWTConfig holds token signing settings.
type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// GeminiConfig holds Google Generative AI settings.
type GeminiConfig struct {
	APIKey  string
	Model   string
	Timeout time.Duration
	Enabled bool
}

// Usable reports whether a live Gemini call should even be attempted. When
// false the assistant runs on the deterministic rule-based extractor alone,
// which keeps the whole product demo-able with zero external dependencies.
func (g GeminiConfig) Usable() bool {
	return g.Enabled && strings.TrimSpace(g.APIKey) != ""
}

// CORSConfig holds cross-origin settings.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

// RateLimitConfig holds token-bucket settings. The assistant gets its own,
// much tighter bucket because every request there can cost a Gemini call.
type RateLimitConfig struct {
	RPS            float64
	Burst          int
	AssistantRPS   float64
	AssistantBurst int
}

// RecommendationConfig holds the tunable knobs of the ranking engine. These are
// deliberately configuration rather than constants: relevance tuning is an
// operational activity, not a code change.
type RecommendationConfig struct {
	Weights         Weights
	DiversityLambda float64
	TopN            int
	CandidatePool   int
}

// Weights are the per-signal contributions to a product's relevance score.
// They are normalised at load time so they always sum to 1.0, which keeps the
// final score interpretable as a 0..1 confidence.
type Weights struct {
	Occasion float64
	Style    float64
	Gender   float64
	Category float64
	Budget   float64
	Color    float64
	Brand    float64
	Text     float64
	Quality  float64
	Value    float64
}

// Sum returns the total of all weights.
func (w Weights) Sum() float64 {
	return w.Occasion + w.Style + w.Gender + w.Category + w.Budget +
		w.Color + w.Brand + w.Text + w.Quality + w.Value
}

// Normalized returns a copy of the weights scaled to sum to 1.0. If the caller
// configured all-zero weights the original is returned unchanged; Load rejects
// that case before this is ever reached.
func (w Weights) Normalized() Weights {
	s := w.Sum()
	if s <= 0 {
		return w
	}
	return Weights{
		Occasion: w.Occasion / s,
		Style:    w.Style / s,
		Gender:   w.Gender / s,
		Category: w.Category / s,
		Budget:   w.Budget / s,
		Color:    w.Color / s,
		Brand:    w.Brand / s,
		Text:     w.Text / s,
		Quality:  w.Quality / s,
		Value:    w.Value / s,
	}
}

// Load reads configuration from the environment (and an optional .env file),
// applies defaults, and validates the result.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./..")
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		// A missing .env is normal (containers inject real env vars).
		// A malformed .env is not, and must fail loudly.
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("config: reading .env: %w", err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:           v.GetString("PORT"),
			Env:            v.GetString("ENV"),
			LogLevel:       v.GetString("LOG_LEVEL"),
			RequestTimeout: time.Duration(v.GetInt("REQUEST_TIMEOUT_SECONDS")) * time.Second,
			AllowGuest:     v.GetBool("ALLOW_GUEST"),
		},
		Database: DatabaseConfig{
			URL:          v.GetString("DATABASE_URL"),
			MaxOpenConns: v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: v.GetInt("DB_MAX_IDLE_CONNS"),
			AutoMigrate:  v.GetBool("DB_AUTO_MIGRATE"),
			AutoSeed:     v.GetBool("DB_AUTO_SEED"),
		},
		Redis: RedisConfig{
			URL:      v.GetString("REDIS_URL"),
			CacheTTL: time.Duration(v.GetInt("CACHE_TTL_MINUTES")) * time.Minute,
		},
		JWT: JWTConfig{
			Secret:     v.GetString("JWT_SECRET"),
			AccessTTL:  time.Duration(v.GetInt("JWT_ACCESS_TTL_MINUTES")) * time.Minute,
			RefreshTTL: time.Duration(v.GetInt("JWT_REFRESH_TTL_HOURS")) * time.Hour,
		},
		Gemini: GeminiConfig{
			APIKey:  strings.TrimSpace(v.GetString("GEMINI_API_KEY")),
			Model:   v.GetString("GEMINI_MODEL"),
			Timeout: time.Duration(v.GetInt("GEMINI_TIMEOUT_SECONDS")) * time.Second,
			Enabled: v.GetBool("GEMINI_ENABLED"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   splitAndTrim(v.GetString("CORS_ALLOWED_ORIGINS")),
			AllowCredentials: v.GetBool("CORS_ALLOW_CREDENTIALS"),
		},
		RateLimit: RateLimitConfig{
			RPS:            v.GetFloat64("RATE_LIMIT_RPS"),
			Burst:          v.GetInt("RATE_LIMIT_BURST"),
			AssistantRPS:   v.GetFloat64("ASSISTANT_RATE_LIMIT_RPS"),
			AssistantBurst: v.GetInt("ASSISTANT_RATE_LIMIT_BURST"),
		},
		Recommendation: RecommendationConfig{
			Weights: Weights{
				Occasion: v.GetFloat64("REC_W_OCCASION"),
				Style:    v.GetFloat64("REC_W_STYLE"),
				Gender:   v.GetFloat64("REC_W_GENDER"),
				Category: v.GetFloat64("REC_W_CATEGORY"),
				Budget:   v.GetFloat64("REC_W_BUDGET"),
				Color:    v.GetFloat64("REC_W_COLOR"),
				Brand:    v.GetFloat64("REC_W_BRAND"),
				Text:     v.GetFloat64("REC_W_TEXT"),
				Quality:  v.GetFloat64("REC_W_QUALITY"),
				Value:    v.GetFloat64("REC_W_VALUE"),
			},
			DiversityLambda: v.GetFloat64("REC_DIVERSITY_LAMBDA"),
			TopN:            v.GetInt("REC_TOP_N"),
			CandidatePool:   v.GetInt("REC_CANDIDATE_POOL"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg.Recommendation.Weights = cfg.Recommendation.Weights.Normalized()

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENV", "development")
	v.SetDefault("LOG_LEVEL", "debug")
	v.SetDefault("REQUEST_TIMEOUT_SECONDS", 20)
	v.SetDefault("ALLOW_GUEST", true)

	v.SetDefault("DATABASE_URL", "postgres://myntra:myntra@localhost:5432/myntra_bharat_ai?sslmode=disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)
	v.SetDefault("DB_AUTO_MIGRATE", true)
	v.SetDefault("DB_AUTO_SEED", true)

	v.SetDefault("REDIS_URL", "redis://localhost:6379/0")
	v.SetDefault("CACHE_TTL_MINUTES", 10)

	v.SetDefault("JWT_SECRET", "")
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 60)
	v.SetDefault("JWT_REFRESH_TTL_HOURS", 168)

	v.SetDefault("GEMINI_API_KEY", "")
	v.SetDefault("GEMINI_MODEL", "gemini-2.0-flash")
	v.SetDefault("GEMINI_TIMEOUT_SECONDS", 8)
	v.SetDefault("GEMINI_ENABLED", true)

	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000,http://localhost:4173")
	v.SetDefault("CORS_ALLOW_CREDENTIALS", true)

	v.SetDefault("RATE_LIMIT_RPS", 20)
	v.SetDefault("RATE_LIMIT_BURST", 40)
	v.SetDefault("ASSISTANT_RATE_LIMIT_RPS", 2)
	v.SetDefault("ASSISTANT_RATE_LIMIT_BURST", 5)

	// Signal weights — these are normalised at load time so they don't need
	// to sum to 1. Occasion and Category are the dominant signals: an Indian
	// fashion AI must get the occasion right above all else.
	v.SetDefault("REC_W_OCCASION", 0.30) // primary signal — must dominate
	v.SetDefault("REC_W_CATEGORY", 0.20) // category suitability is occasion-driven
	v.SetDefault("REC_W_STYLE", 0.18)    // style fit (Ethnic/Formal/Casual etc.)
	v.SetDefault("REC_W_GENDER", 0.12)   // hard filter expressed as soft weight
	v.SetDefault("REC_W_BUDGET", 0.10)   // budget fit
	v.SetDefault("REC_W_QUALITY", 0.05)  // rating quality tiebreaker
	v.SetDefault("REC_W_VALUE", 0.03)    // discount / value
	v.SetDefault("REC_W_COLOR", 0.01)
	v.SetDefault("REC_W_BRAND", 0.01)
	v.SetDefault("REC_W_TEXT", 0.00)
	v.SetDefault("REC_DIVERSITY_LAMBDA", 0.72)
	v.SetDefault("REC_TOP_N", 6)
	v.SetDefault("REC_CANDIDATE_POOL", 200)
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return errors.New("config: DATABASE_URL is required")
	}
	if len(c.JWT.Secret) < 16 {
		if c.Server.IsProduction() {
			return errors.New("config: JWT_SECRET must be at least 16 characters in production")
		}
		// Developers should not have to invent a secret just to run `go run`.
		c.JWT.Secret = "dev-only-insecure-jwt-secret-do-not-ship"
	}
	if c.Recommendation.Weights.Sum() <= 0 {
		return errors.New("config: all REC_W_* weights are zero; the ranker would be undefined")
	}
	if c.Recommendation.TopN <= 0 {
		c.Recommendation.TopN = 6
	}
	if c.Recommendation.CandidatePool < c.Recommendation.TopN {
		c.Recommendation.CandidatePool = 200
	}
	if c.Recommendation.DiversityLambda < 0 || c.Recommendation.DiversityLambda > 1 {
		return errors.New("config: REC_DIVERSITY_LAMBDA must be within [0,1]")
	}
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	return nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
