package config

import (
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RateLimitRule represents a single rate limiting rule
type RateLimitRule struct {
	Method  string        `yaml:"method"`
	Path    string        `yaml:"path"`
	Limit   int           `yaml:"limit"`
	Window  time.Duration `yaml:"window"`
	pattern *regexp.Regexp
}

// RateLimitDefault holds default rate limit values
type RateLimitDefault struct {
	Limit  int           `yaml:"limit"`
	Window time.Duration `yaml:"window"`
}

// RateLimitConfig holds the complete rate limiting configuration
type RateLimitConfig struct {
	Default RateLimitDefault `yaml:"default"`
	Rules   []RateLimitRule  `yaml:"rules"`
}

// RateLimit is the global rate limit configuration instance
var RateLimit *RateLimitConfig

// loadRateLimit loads rate limiting configuration from YAML file
func loadRateLimit() *RateLimitConfig {
	cfg := &RateLimitConfig{
		Default: RateLimitDefault{
			Limit:  100,
			Window: time.Minute,
		},
	}

	configPath := GetEnv("RATE_LIMIT_CONFIG", "ratelimit.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg
	}

	for i := range cfg.Rules {
		cfg.Rules[i].pattern = compilePathPattern(cfg.Rules[i].Path)
	}

	return cfg
}

// compilePathPattern converts wildcard patterns to regex
// /api/v1/activities/* -> ^/api/v1/activities/[^/]+$
// /api/v1/activities/*/photos -> ^/api/v1/activities/[^/]+/photos$
func compilePathPattern(pattern string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, `\*`, `[^/]+`)
	return regexp.MustCompile("^" + escaped + "$")
}

// ReloadRateLimit reads ratelimit.yaml from disk and returns a fresh config.
// This is used by the background refresh job to re-read and re-cache config.
func ReloadRateLimit() *RateLimitConfig {
	return loadRateLimit()
}

// CompilePatterns recompiles all path patterns after JSON unmarshaling.
// Must be called after unmarshaling a RateLimitConfig from JSON/Redis.
func (c *RateLimitConfig) CompilePatterns() {
	for i := range c.Rules {
		c.Rules[i].pattern = compilePathPattern(c.Rules[i].Path)
	}
}

// FindRule returns the best matching rule for method+path
// Returns limit and window duration for the matched rule
func (c *RateLimitConfig) FindRule(method, path string) (int, time.Duration) {
	var bestMatch *RateLimitRule
	bestScore := -1

	for i := range c.Rules {
		rule := &c.Rules[i]

		if rule.Method != "*" && rule.Method != method {
			continue
		}

		if !rule.pattern.MatchString(path) {
			continue
		}

		score := len(rule.Path)
		if rule.Method != "*" {
			score += 1000
		}

		if score > bestScore {
			bestScore = score
			bestMatch = rule
		}
	}

	if bestMatch != nil {
		return bestMatch.Limit, bestMatch.Window
	}

	return c.Default.Limit, c.Default.Window
}
