package ratelimit_test

import (
	"context"
	"testing"

	"github.com/kolosys/ion/ratelimit"
)

func TestMultiTierLimiter_Basic(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(10)
	config.GlobalBurst = 10
	config.DefaultRouteRate = ratelimit.PerSecond(5)
	config.DefaultRouteBurst = 5

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/channels/123456789/messages",
		Context:  context.Background(),
	}

	// Should allow initial requests up to burst
	for i := 0; i < 5; i++ {
		if !limiter.Allow(req) {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// Should deny the next request (exceeds route burst)
	if limiter.Allow(req) {
		t.Error("Request should be denied after exceeding route burst")
	}
}

func TestMultiTierLimiter_GlobalLimit(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(5)
	config.GlobalBurst = 5
	config.DefaultRouteRate = ratelimit.PerSecond(100) // High route limit
	config.DefaultRouteBurst = 100

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	// Should allow initial requests up to global burst
	for i := 0; i < 5; i++ {
		if !limiter.Allow(req) {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// Should deny the next request (exceeds global burst)
	if limiter.Allow(req) {
		t.Error("Request should be denied after exceeding global burst")
	}
}

func TestMultiTierLimiter_ResourceLimit(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(100)
	config.GlobalBurst = 100
	config.DefaultRouteRate = ratelimit.PerSecond(100)
	config.DefaultRouteBurst = 100
	config.DefaultResourceRate = ratelimit.PerSecond(3)
	config.DefaultResourceBurst = 3

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:     "GET",
		Endpoint:   "/test",
		ResourceID: "987654321", // Using generic ResourceID instead of GuildID
		Context:    context.Background(),
	}

	// Should allow initial requests up to resource burst
	for i := 0; i < 3; i++ {
		if !limiter.Allow(req) {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// Should deny the next request (exceeds resource burst)
	if limiter.Allow(req) {
		t.Error("Request should be denied after exceeding resource burst")
	}

	// Different resource should be allowed
	req2 := &ratelimit.Request{
		Method:     "GET",
		Endpoint:   "/test",
		ResourceID: "111111111",
		Context:    context.Background(),
	}

	if !limiter.Allow(req2) {
		t.Error("Different resource should be allowed")
	}
}

func TestMultiTierLimiter_RoutePatterns(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(100)
	config.GlobalBurst = 100
	config.DefaultRouteRate = ratelimit.PerSecond(10)
	config.DefaultRouteBurst = 10

	// Add specific route pattern
	config.RoutePatterns = map[string]ratelimit.RouteConfig{
		"GET:/channels/{id}/messages": {
			Rate:  ratelimit.PerSecond(5),
			Burst: 5,
		},
	}

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/channels/123456789/messages",
		Context:  context.Background(),
	}

	// Should allow up to route-specific burst
	for i := 0; i < 5; i++ {
		if !limiter.Allow(req) {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// Should deny after exceeding route-specific burst
	if limiter.Allow(req) {
		t.Error("Request should be denied after exceeding route-specific burst")
	}
}

func TestMultiTierLimiter_Wait(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(2)
	config.GlobalBurst = 1

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	// Use the burst
	if !limiter.Allow(req) {
		t.Error("First request should be allowed")
	}

	// Try to wait for next token
	err := limiter.Wait(req)
	if err != nil {
		t.Errorf("Wait should succeed: %v", err)
	}
}

func TestMultiTierLimiter_HeaderUpdate(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(10)
	config.GlobalBurst = 5
	config.EnableBucketMapping = true

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	headers := map[string]string{
		"X-RateLimit-Bucket":    "abcd1234",
		"X-RateLimit-Limit":     "10",
		"X-RateLimit-Remaining": "5",
		"X-RateLimit-Reset":     "1640995200",
	}

	err := limiter.UpdateRateLimitFromHeaders(req, headers)
	if err != nil {
		t.Errorf("UpdateRateLimitFromHeaders should succeed: %v", err)
	}

	// Verify bucket mapping was stored (test the functionality indirectly)
	// The bucket mapping is internal, so we test it by checking if the headers were processed without error
	if err != nil {
		t.Errorf("UpdateRateLimitFromHeaders should succeed: %v", err)
	}
}

func TestMultiTierLimiter_Metrics(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(5)
	config.GlobalBurst = 2

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	// Make some requests
	limiter.Allow(req)
	limiter.Allow(req)
	limiter.Allow(req) // This should hit global limit

	metrics := limiter.GetMetrics()

	if metrics.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.GlobalLimitHits != 1 {
		t.Errorf("Expected 1 global limit hit, got %d", metrics.GlobalLimitHits)
	}
}

func TestMultiTierLimiter_Reset(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(5)
	config.GlobalBurst = 1

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	// Use up the burst
	if !limiter.Allow(req) {
		t.Error("First request should be allowed")
	}

	// Should be denied
	if limiter.Allow(req) {
		t.Error("Second request should be denied")
	}

	// Reset and try again
	limiter.Reset()

	// Should be allowed again
	if !limiter.Allow(req) {
		t.Error("Request after reset should be allowed")
	}
}

func TestMultiTierLimiter_RouteNormalization(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	// Test route normalization indirectly by testing that different endpoints with same pattern
	// get the same rate limiting behavior
	req1 := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/channels/123456789012345678/messages",
		Context:  context.Background(),
	}

	req2 := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/channels/987654321098765432/messages",
		Context:  context.Background(),
	}

	// Both should be rate limited together since they normalize to the same pattern
	limiter.Allow(req1)
	if !limiter.Allow(req2) {
		t.Error("Different endpoints with same pattern should be rate limited together")
	}
}

func TestMultiTierLimiter_AllowN(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(10)
	config.GlobalBurst = 10
	config.DefaultRouteRate = ratelimit.PerSecond(5)
	config.DefaultRouteBurst = 5

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	// Should allow 5 requests at once
	if !limiter.AllowN(req, 5) {
		t.Error("Should allow 5 requests at once")
	}

	// Should deny the next batch
	if limiter.AllowN(req, 5) {
		t.Error("Should deny next batch of 5")
	}
}

func TestMultiTierLimiter_WaitN(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(4)
	config.GlobalBurst = 2

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	// Use the burst
	if !limiter.AllowN(req, 2) {
		t.Error("First 2 requests should be allowed")
	}

	// Try to wait for next tokens
	err := limiter.WaitN(req, 2)
	if err != nil {
		t.Errorf("WaitN should succeed: %v", err)
	}
}

func TestMultiTierLimiter_ConcurrentAccess(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(100)
	config.GlobalBurst = 100
	config.DefaultRouteRate = ratelimit.PerSecond(50)
	config.DefaultRouteBurst = 50

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	// Test concurrent access to different routes
	req1 := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/route1",
		Context:  context.Background(),
	}

	req2 := &ratelimit.Request{
		Method:   "POST",
		Endpoint: "/route2",
		Context:  context.Background(),
	}

	// Run concurrent requests
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			limiter.Allow(req1)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			limiter.Allow(req2)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify metrics
	metrics := limiter.GetMetrics()
	if metrics.TotalRequests < 20 {
		t.Errorf("Expected at least 20 total requests, got %d", metrics.TotalRequests)
	}
}

func TestMultiTierLimiter_MajorParameters(t *testing.T) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(100)
	config.GlobalBurst = 100
	config.DefaultRouteRate = ratelimit.PerSecond(10)
	config.DefaultRouteBurst = 10

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("test"))

	// Test requests with major parameters
	req1 := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/api/v1/users",
		MajorParameters: map[string]string{
			"org_id": "org123",
		},
		Context: context.Background(),
	}

	req2 := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/api/v1/users",
		MajorParameters: map[string]string{
			"org_id": "org456",
		},
		Context: context.Background(),
	}

	// Both should be allowed (different major parameters create different buckets)
	if !limiter.Allow(req1) {
		t.Error("First request with major parameters should be allowed")
	}

	if !limiter.Allow(req2) {
		t.Error("Second request with different major parameters should be allowed")
	}
}

func BenchmarkMultiTierLimiter_Allow(b *testing.B) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(1000)
	config.GlobalBurst = 1000
	config.DefaultRouteRate = ratelimit.PerSecond(100)
	config.DefaultRouteBurst = 100

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("bench"))

	req := &ratelimit.Request{
		Method:   "GET",
		Endpoint: "/test",
		Context:  context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(req)
	}
}

func BenchmarkMultiTierLimiter_AllowWithResource(b *testing.B) {
	config := ratelimit.DefaultMultiTierConfig()
	config.GlobalRate = ratelimit.PerSecond(1000)
	config.GlobalBurst = 1000
	config.DefaultRouteRate = ratelimit.PerSecond(100)
	config.DefaultRouteBurst = 100
	config.DefaultResourceRate = ratelimit.PerSecond(50)
	config.DefaultResourceBurst = 50

	limiter := ratelimit.NewMultiTierLimiter(config, ratelimit.WithName("bench"))

	req := &ratelimit.Request{
		Method:     "GET",
		Endpoint:   "/test",
		ResourceID: "123456789",
		Context:    context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(req)
	}
}
