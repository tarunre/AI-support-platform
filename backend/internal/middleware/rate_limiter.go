package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ayush/supportiq/internal/utils"
)

// visitor tracks the request count in the current window for a single IP.
type visitor struct {
	count     int
	windowEnd time.Time
}

// ipLimiter is an in-process, per-IP rate limiter.
// For production deployments with multiple replicas, replace with a Redis-backed
// sliding-window implementation.
type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func newIPLimiter(limit int, window time.Duration) *ipLimiter {
	l := &ipLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	// Background goroutine cleans up expired entries to prevent memory growth.
	go func() {
		for range time.Tick(window) {
			l.mu.Lock()
			now := time.Now()
			for ip, v := range l.visitors {
				if now.After(v.windowEnd) {
					delete(l.visitors, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	v, exists := l.visitors[ip]
	if !exists || now.After(v.windowEnd) {
		l.visitors[ip] = &visitor{count: 1, windowEnd: now.Add(l.window)}
		return true
	}
	v.count++
	return v.count <= l.limit
}

// defaultLimiter: 200 requests/minute for authenticated routes.
var defaultLimiter = newIPLimiter(200, time.Minute)

// authLimiter: 20 requests/minute for auth endpoints (brute-force protection).
var authLimiter = newIPLimiter(20, time.Minute)

// RateLimit applies the default per-IP rate limit (200 req/min).
func RateLimit() gin.HandlerFunc {
	return rateLimitWith(defaultLimiter)
}

// RateLimitAuth applies a stricter per-IP rate limit for authentication endpoints (20 req/min).
func RateLimitAuth() gin.HandlerFunc {
	return rateLimitWith(authLimiter)
}

func rateLimitWith(l *ipLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !l.allow(ip) {
			utils.Logger.WithField("ip", ip).Warn("Rate limit exceeded")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Too many requests — please slow down and retry later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
