package ratelimit

import (
	"context"
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MultiTierLimiter implements a sophisticated multi-tier rate limiting system.
// It supports global, per-route, and per-resource rate limiting with intelligent
// bucket management and flexible API compatibility.
type MultiTierLimiter struct {
	mu sync.RWMutex

	// Global limiter shared across all requests
	global Limiter

	// Route limiters for specific API endpoints
	routes sync.Map // map[string]Limiter

	// Resource limiters for specific resources (organizations, projects, etc.)
	resources sync.Map // map[string]Limiter

	// Bucket mapping for API-style rate limit buckets
	bucketMap sync.Map // map[string]string

	// Configuration
	config *MultiTierConfig
	cfg    *config

	// Metrics and observability
	metrics *MultiTierMetrics

	// Pause state
	pausedUntil time.Time
	pauseTimer  Timer
}

// MultiTierConfig holds configuration for multi-tier rate limiting.
type MultiTierConfig struct {
	// Global rate limit configuration
	GlobalRate  Rate
	GlobalBurst int

	// Default rate limits for routes and resources
	DefaultRouteRate     Rate
	DefaultRouteBurst    int
	DefaultResourceRate  Rate
	DefaultResourceBurst int

	// Queue configuration for request management
	QueueSize        int
	EnablePreemptive bool

	// Bucket management
	EnableBucketMapping bool
	BucketTTL           time.Duration

	// Route pattern matching
	RoutePatterns map[string]RouteConfig
}

// RouteConfig defines rate limiting for specific route patterns.
type RouteConfig struct {
	Rate  Rate
	Burst int
	// Major parameters that affect rate limiting (e.g., org_id, project_id)
	MajorParameters []string
}

// MultiTierMetrics tracks metrics for multi-tier rate limiting.
type MultiTierMetrics struct {
	mu sync.RWMutex

	TotalRequests     int64
	GlobalLimitHits   int64
	RouteLimitHits    int64
	ResourceLimitHits int64
	QueuedRequests    int64
	DroppedRequests   int64
	AvgWaitTime       time.Duration
	MaxWaitTime       time.Duration
	BucketsActive     int64
}

// Request represents a request for rate limiting evaluation.
type Request struct {
	// Route information
	Method   string
	Endpoint string

	// Resource identifiers (generic - applications define their own)
	ResourceID    string // Primary resource identifier
	SubResourceID string // Secondary resource identifier
	UserID        string // User/actor identifier

	// Major parameters for bucket identification
	MajorParameters map[string]string

	// Request metadata
	Priority int
	Context  context.Context
}

// DefaultMultiTierConfig returns a default configuration for multi-tier rate limiting.
// Applications should customize this configuration for their specific needs.
func DefaultMultiTierConfig() *MultiTierConfig {
	return &MultiTierConfig{
		GlobalRate:           PerSecond(100), // Conservative global limit
		GlobalBurst:          100,
		DefaultRouteRate:     PerSecond(50),
		DefaultRouteBurst:    50,
		DefaultResourceRate:  PerSecond(20),
		DefaultResourceBurst: 20,
		QueueSize:            1000,
		EnablePreemptive:     true,
		EnableBucketMapping:  true,
		BucketTTL:            time.Hour,
		RoutePatterns:        make(map[string]RouteConfig), // No default patterns
	}
}

// NewMultiTierLimiter creates a new multi-tier rate limiter.
func NewMultiTierLimiter(config *MultiTierConfig, opts ...Option) *MultiTierLimiter {
	if config == nil {
		config = DefaultMultiTierConfig()
	}

	cfg := newConfig(opts...)

	globalLimiter := NewTokenBucket(config.GlobalRate, config.GlobalBurst,
		WithName(cfg.name+"_global"),
		WithClock(cfg.clock),
		WithJitter(cfg.jitter),
		WithLogger(cfg.obs.Logger),
		WithMetrics(cfg.obs.Metrics),
		WithTracer(cfg.obs.Tracer),
	)

	mtl := &MultiTierLimiter{
		global:  globalLimiter,
		config:  config,
		cfg:     cfg,
		metrics: &MultiTierMetrics{},
	}

	cfg.obs.Logger.Info("multi-tier rate limiter created",
		"name", cfg.name,
		"global_rate", config.GlobalRate.String(),
		"global_burst", config.GlobalBurst,
		"queue_size", config.QueueSize,
	)

	return mtl
}

// Allow checks if a request is allowed without blocking.
func (mtl *MultiTierLimiter) Allow(req *Request) bool {
	return mtl.AllowN(req, 1)
}

// AllowN checks if n requests are allowed without blocking.
func (mtl *MultiTierLimiter) AllowN(req *Request, n int) bool {
	now := mtl.cfg.clock.Now()

	if mtl.IsPaused() {
		mtl.updateMetrics(func(m *MultiTierMetrics) {
			m.GlobalLimitHits++
		})
		return false
	}

	if !mtl.global.AllowN(now, n) {
		mtl.updateMetrics(func(m *MultiTierMetrics) {
			m.GlobalLimitHits++
		})
		return false
	}

	routeLimiter := mtl.getOrCreateRouteLimiter(req)
	if !routeLimiter.AllowN(now, n) {
		mtl.updateMetrics(func(m *MultiTierMetrics) {
			m.RouteLimitHits++
		})
		return false
	}

	if resourceLimiter := mtl.getResourceLimiter(req); resourceLimiter != nil {
		if !resourceLimiter.AllowN(now, n) {
			mtl.updateMetrics(func(m *MultiTierMetrics) {
				m.ResourceLimitHits++
			})
			return false
		}
	}

	mtl.updateMetrics(func(m *MultiTierMetrics) {
		m.TotalRequests += int64(n)
	})

	return true
}

// Wait blocks until the request is allowed or context is canceled.
func (mtl *MultiTierLimiter) Wait(req *Request) error {
	return mtl.WaitN(req, 1)
}

// WaitN blocks until n requests are allowed or context is canceled.
func (mtl *MultiTierLimiter) WaitN(req *Request, n int) error {
	ctx := req.Context
	if ctx == nil {
		ctx = context.Background()
	}

	start := mtl.cfg.clock.Now()

	if err := mtl.waitForPause(ctx); err != nil {
		return err
	}

	// Fast path: try immediate approval
	if mtl.AllowN(req, n) {
		return nil
	}

	// Slow path: wait for each tier
	limiters := []struct {
		limiter Limiter
		name    string
	}{
		{mtl.global, "global"},
		{mtl.getOrCreateRouteLimiter(req), "route"},
	}

	if resourceLimiter := mtl.getResourceLimiter(req); resourceLimiter != nil {
		limiters = append(limiters, struct {
			limiter Limiter
			name    string
		}{resourceLimiter, "resource"})
	}

	for _, l := range limiters {
		if err := l.limiter.WaitN(ctx, n); err != nil {
			mtl.cfg.obs.Logger.Debug("rate limit wait failed",
				"limiter_name", mtl.cfg.name,
				"tier", l.name,
				"error", err,
			)
			return err
		}
	}

	waitTime := mtl.cfg.clock.Now().Sub(start)
	mtl.updateMetrics(func(m *MultiTierMetrics) {
		m.TotalRequests += int64(n)
		if waitTime > m.MaxWaitTime {
			m.MaxWaitTime = waitTime
		}
		if m.AvgWaitTime == 0 {
			m.AvgWaitTime = waitTime
		} else {
			m.AvgWaitTime = (m.AvgWaitTime + waitTime) / 2
		}
	})

	return nil
}

// getOrCreateRouteLimiter gets or creates a route-specific limiter.
func (mtl *MultiTierLimiter) getOrCreateRouteLimiter(req *Request) Limiter {
	routeKey := mtl.generateRouteKey(req)

	if limiter, ok := mtl.routes.Load(routeKey); ok {
		return limiter.(Limiter)
	}

	routeConfig := mtl.findRouteConfig(req.Method, req.Endpoint)

	limiter := NewTokenBucket(
		routeConfig.Rate,
		routeConfig.Burst,
		WithName(fmt.Sprintf("%s_route_%s", mtl.cfg.name, routeKey)),
		WithClock(mtl.cfg.clock),
		WithJitter(mtl.cfg.jitter),
		WithLogger(mtl.cfg.obs.Logger),
		WithMetrics(mtl.cfg.obs.Metrics),
		WithTracer(mtl.cfg.obs.Tracer),
	)

	actual, loaded := mtl.routes.LoadOrStore(routeKey, limiter)
	if loaded {
		return actual.(Limiter)
	}

	mtl.updateMetrics(func(m *MultiTierMetrics) {
		m.BucketsActive++
	})

	return limiter
}

// getResourceLimiter gets a resource-specific limiter if applicable.
func (mtl *MultiTierLimiter) getResourceLimiter(req *Request) Limiter {
	var resourceKey string

	if req.ResourceID != "" {
		resourceKey = "resource:" + req.ResourceID
	} else if req.SubResourceID != "" {
		resourceKey = "subresource:" + req.SubResourceID
	} else if req.UserID != "" {
		resourceKey = "user:" + req.UserID
	} else {
		return nil // No resource limiting needed
	}

	if limiter, ok := mtl.resources.Load(resourceKey); ok {
		return limiter.(Limiter)
	}

	limiter := NewTokenBucket(
		mtl.config.DefaultResourceRate,
		mtl.config.DefaultResourceBurst,
		WithName(fmt.Sprintf("%s_resource_%s", mtl.cfg.name, resourceKey)),
		WithClock(mtl.cfg.clock),
		WithJitter(mtl.cfg.jitter),
		WithLogger(mtl.cfg.obs.Logger),
		WithMetrics(mtl.cfg.obs.Metrics),
		WithTracer(mtl.cfg.obs.Tracer),
	)

	actual, loaded := mtl.resources.LoadOrStore(resourceKey, limiter)
	if loaded {
		return actual.(Limiter)
	}

	mtl.updateMetrics(func(m *MultiTierMetrics) {
		m.BucketsActive++
	})

	return limiter
}

// generateRouteKey creates a unique key for route identification.
func (mtl *MultiTierLimiter) generateRouteKey(req *Request) string {
	pattern := mtl.normalizeRoute(req.Method, req.Endpoint)

	if len(req.MajorParameters) == 0 {
		return pattern
	}

	h := md5.New()
	h.Write([]byte(pattern))
	for key, value := range req.MajorParameters {
		h.Write([]byte(key + ":" + value))
	}

	return fmt.Sprintf("%s_%x", pattern, h.Sum(nil)[:8])
}

// normalizeRoute normalizes an API route for pattern matching.
func (mtl *MultiTierLimiter) normalizeRoute(method, endpoint string) string {
	idPattern := regexp.MustCompile(`\d+`)
	normalized := idPattern.ReplaceAllString(endpoint, "{id}")
	normalized = strings.ReplaceAll(normalized, "//", "/")
	normalized = strings.TrimSuffix(normalized, "/")
	return method + ":" + normalized
}

// findRouteConfig finds the configuration for a specific route.
func (mtl *MultiTierLimiter) findRouteConfig(method, endpoint string) RouteConfig {
	normalized := mtl.normalizeRoute(method, endpoint)

	if config, ok := mtl.config.RoutePatterns[normalized]; ok {
		return config
	}

	for pattern, config := range mtl.config.RoutePatterns {
		if mtl.matchesPattern(normalized, pattern) {
			return config
		}
	}

	return RouteConfig{
		Rate:  mtl.config.DefaultRouteRate,
		Burst: mtl.config.DefaultRouteBurst,
	}
}

// matchesPattern checks if an endpoint matches a route pattern.
func (mtl *MultiTierLimiter) matchesPattern(endpoint, pattern string) bool {
	endpointParts := strings.Split(endpoint, "/")
	patternParts := strings.Split(pattern, "/")

	if len(endpointParts) != len(patternParts) {
		return false
	}

	for i, part := range patternParts {
		if part != "{id}" && part != endpointParts[i] {
			return false
		}
	}

	return true
}

// UpdateRateLimitFromHeaders updates rate limit information from API response headers.
// This is designed for APIs that provide rate limit information in response headers.
func (mtl *MultiTierLimiter) UpdateRateLimitFromHeaders(req *Request, headers map[string]string) error {
	limit := mtl.parseIntHeader(headers, "X-RateLimit-Limit", 0)
	remaining := mtl.parseIntHeader(headers, "X-RateLimit-Remaining", 0)
	resetAfter := mtl.parseFloatHeader(headers, "X-RateLimit-Reset-After", 0)
	global := headers["X-RateLimit-Global"] == "true"
	bucket := headers["X-RateLimit-Bucket"]

	if bucket != "" && mtl.config.EnableBucketMapping {
		routeKey := mtl.generateRouteKey(req)
		mtl.bucketMap.Store(routeKey, bucket)
	}

	if global && resetAfter > 0 {
		mtl.cfg.obs.Logger.Warn("global rate limit hit",
			"limiter_name", mtl.cfg.name,
			"reset_after", resetAfter,
		)
		// Schedule auto-resume
		resetTime := mtl.cfg.clock.Now().Add(time.Duration(resetAfter * float64(time.Second)))
		mtl.PauseUntil(resetTime)
	}

	mtl.cfg.obs.Logger.Debug("rate limit headers processed",
		"limiter_name", mtl.cfg.name,
		"limit", limit,
		"remaining", remaining,
		"reset_after", resetAfter,
		"global", global,
		"bucket", bucket,
	)

	return nil
}

// GetMetrics returns current rate limiting metrics.
func (mtl *MultiTierLimiter) GetMetrics() *MultiTierMetrics {
	mtl.metrics.mu.RLock()
	defer mtl.metrics.mu.RUnlock()

	return &MultiTierMetrics{
		TotalRequests:     mtl.metrics.TotalRequests,
		GlobalLimitHits:   mtl.metrics.GlobalLimitHits,
		RouteLimitHits:    mtl.metrics.RouteLimitHits,
		ResourceLimitHits: mtl.metrics.ResourceLimitHits,
		QueuedRequests:    mtl.metrics.QueuedRequests,
		DroppedRequests:   mtl.metrics.DroppedRequests,
		AvgWaitTime:       mtl.metrics.AvgWaitTime,
		MaxWaitTime:       mtl.metrics.MaxWaitTime,
		BucketsActive:     mtl.metrics.BucketsActive,
	}
}

// updateMetrics safely updates metrics using a function.
func (mtl *MultiTierLimiter) updateMetrics(fn func(*MultiTierMetrics)) {
	mtl.metrics.mu.Lock()
	defer mtl.metrics.mu.Unlock()
	fn(mtl.metrics)
}

// parseIntHeader parses an integer header value.
func (mtl *MultiTierLimiter) parseIntHeader(headers map[string]string, key string, defaultValue int) int {
	if value, ok := headers[key]; ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseFloatHeader parses a float header value.
func (mtl *MultiTierLimiter) parseFloatHeader(headers map[string]string, key string, defaultValue float64) float64 {
	if value, ok := headers[key]; ok {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Reset resets all rate limit buckets (useful for testing).
func (mtl *MultiTierLimiter) Reset() {
	if tb, ok := mtl.global.(*TokenBucket); ok {
		tb.mu.Lock()
		tb.tokens = float64(tb.burst)
		tb.lastRefill = mtl.cfg.clock.Now()
		tb.mu.Unlock()
	}

	mtl.routes.Range(func(key, value interface{}) bool {
		if tb, ok := value.(*TokenBucket); ok {
			tb.mu.Lock()
			tb.tokens = float64(tb.burst)
			tb.lastRefill = mtl.cfg.clock.Now()
			tb.mu.Unlock()
		}
		return true
	})

	mtl.resources.Range(func(key, value interface{}) bool {
		if tb, ok := value.(*TokenBucket); ok {
			tb.mu.Lock()
			tb.tokens = float64(tb.burst)
			tb.lastRefill = mtl.cfg.clock.Now()
			tb.mu.Unlock()
		}
		return true
	})

	mtl.metrics.mu.Lock()
	mtl.metrics.TotalRequests = 0
	mtl.metrics.GlobalLimitHits = 0
	mtl.metrics.RouteLimitHits = 0
	mtl.metrics.ResourceLimitHits = 0
	mtl.metrics.QueuedRequests = 0
	mtl.metrics.DroppedRequests = 0
	mtl.metrics.AvgWaitTime = 0
	mtl.metrics.MaxWaitTime = 0
	mtl.metrics.BucketsActive = 0
	mtl.metrics.mu.Unlock()

	mtl.mu.Lock()
	if mtl.pauseTimer != nil {
		mtl.pauseTimer.Stop()
	}
	mtl.pausedUntil = time.Time{}
	mtl.pauseTimer = nil
	mtl.mu.Unlock()
}

// PauseUntil pauses all requests until the specified time.
// This is useful for handling global rate limits from APIs.
func (mtl *MultiTierLimiter) PauseUntil(until time.Time) {
	mtl.mu.Lock()
	defer mtl.mu.Unlock()

	if mtl.pauseTimer != nil {
		mtl.pauseTimer.Stop()
	}

	mtl.pausedUntil = until
	duration := time.Until(until)

	if duration <= 0 {
		mtl.pausedUntil = time.Time{}
		mtl.pauseTimer = nil
		return
	}

	mtl.cfg.obs.Logger.Warn("rate limiter paused",
		"limiter_name", mtl.cfg.name,
		"until", until,
		"duration", duration,
	)

	// Schedule auto-resume
	mtl.pauseTimer = mtl.cfg.clock.AfterFunc(duration, func() {
		mtl.Resume()
	})
}

// PauseFor pauses all requests for the specified duration.
func (mtl *MultiTierLimiter) PauseFor(duration time.Duration) {
	mtl.PauseUntil(mtl.cfg.clock.Now().Add(duration))
}

// Resume resumes rate limiting after a pause.
func (mtl *MultiTierLimiter) Resume() {
	mtl.mu.Lock()
	defer mtl.mu.Unlock()

	if mtl.pauseTimer != nil {
		mtl.pauseTimer.Stop()
		mtl.pauseTimer = nil
	}

	if !mtl.pausedUntil.IsZero() {
		mtl.cfg.obs.Logger.Info("rate limiter resumed",
			"limiter_name", mtl.cfg.name,
		)
	}

	mtl.pausedUntil = time.Time{}
}

// IsPaused returns whether the limiter is currently paused.
func (mtl *MultiTierLimiter) IsPaused() bool {
	mtl.mu.RLock()
	defer mtl.mu.RUnlock()

	if mtl.pausedUntil.IsZero() {
		return false
	}
	return mtl.cfg.clock.Now().Before(mtl.pausedUntil)
}

// PausedUntil returns the time when the pause will end, or zero if not paused.
func (mtl *MultiTierLimiter) PausedUntil() time.Time {
	mtl.mu.RLock()
	defer mtl.mu.RUnlock()
	return mtl.pausedUntil
}

// waitForPause waits for the pause to end or context to be canceled.
func (mtl *MultiTierLimiter) waitForPause(ctx context.Context) error {
	mtl.mu.RLock()
	pausedUntil := mtl.pausedUntil
	mtl.mu.RUnlock()

	if pausedUntil.IsZero() || mtl.cfg.clock.Now().After(pausedUntil) {
		return nil
	}

	duration := time.Until(pausedUntil)
	if duration <= 0 {
		return nil
	}

	mtl.cfg.obs.Logger.Debug("waiting for pause to end",
		"limiter_name", mtl.cfg.name,
		"duration", duration,
	)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
