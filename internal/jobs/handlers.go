package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	cacheadapter "github.com/valentinesamuel/activelog/internal/cache/adapter/redis"
	cacheTypes "github.com/valentinesamuel/activelog/internal/cache/types"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/queue/types"
)

// HandleWelcomeEmail processes a welcome email job.
// Once an email provider is wired in, this will call emailProvider.Send.
func HandleWelcomeEmail(_ context.Context, payload types.JobPayload) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(payload.Data, &p); err != nil {
		return fmt.Errorf("HandleWelcomeEmail: unmarshal: %w", err)
	}
	log.Printf("[job] welcome email -> userID=%d email=%s name=%s", p.UserID, p.Email, p.Name)
	return nil
}

// HandleWeeklySummary processes a weekly summary email job.
func HandleWeeklySummary(_ context.Context, payload types.JobPayload) error {
	var p WeeklySummaryPayload
	if err := json.Unmarshal(payload.Data, &p); err != nil {
		return fmt.Errorf("HandleWeeklySummary: unmarshal: %w", err)
	}
	log.Printf("[job] weekly summary -> userID=%d", p.UserID)
	return nil
}

// HandleGenerateExport processes a CSV/PDF export generation job.
func HandleGenerateExport(_ context.Context, payload types.JobPayload) error {
	var p ExportPayload
	if err := json.Unmarshal(payload.Data, &p); err != nil {
		return fmt.Errorf("HandleGenerateExport: unmarshal: %w", err)
	}
	log.Printf("[job] generate export -> userID=%d format=%s", p.UserID, p.Format)
	return nil
}

// HandleRefreshRateLimitConfig re-reads ratelimit.yaml and writes a fresh
// CachedRateLimitConfig to Redis DB 3 with a 48-hour TTL.
func HandleRefreshRateLimitConfig(ctx context.Context, _ types.JobPayload) error {
	cfg := config.ReloadRateLimit()

	cachedCfg := struct {
		CachedAt time.Time              `json:"cached_at"`
		Config   config.RateLimitConfig `json:"config"`
	}{
		CachedAt: time.Now(),
		Config:   *cfg,
	}

	data, err := json.Marshal(cachedCfg)
	if err != nil {
		return fmt.Errorf("HandleRefreshRateLimitConfig: marshal: %w", err)
	}

	adapter := cacheadapter.New()
	opts := cacheTypes.CacheOptions{
		DB:           cacheTypes.CacheDBRateLimits,
		PartitionKey: cacheTypes.CachePartitionRateLimitConfig,
	}

	if err := adapter.Set(ctx, "config", string(data), 48*time.Hour, opts); err != nil {
		return fmt.Errorf("HandleRefreshRateLimitConfig: redis set: %w", err)
	}

	log.Printf("[job] rate limit config refreshed in Redis")
	return nil
}
