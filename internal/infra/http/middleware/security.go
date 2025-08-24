package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/config"
	"github.com/williamkoller/golang-payment-stripe/pkg/ulidx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = ulidx.New()
		}
		c.Writer.Header().Set("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'none'")
		c.Writer.Header().Set("X-DNS-Prefetch-Control", "off")
		c.Next()
	}
}

type limiter struct {
	lim *rate.Limiter
}

var lim = limiter{lim: rate.NewLimiter(10, 20)}

func RateLimit(cfg *config.Config) gin.HandlerFunc {
	lim.lim = rate.NewLimiter(rate.Limit(cfg.RateLimitRPS), cfg.RateLimitBurst)
	return func(c *gin.Context) {
		if !lim.lim.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func Timeout(cfg *config.Config, zl *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.RequestTimeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		done := make(chan struct{})
		go func() { c.Next(); close(done) }()
		select {
		case <-ctx.Done():
			zl.Warn("request_timeout", zap.String("path", c.FullPath()))
			c.AbortWithStatus(http.StatusGatewayTimeout)
		case <-done:
		}
	}
}

func GinZapLogger(zl *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		zl.Info("http_request",
			zap.String("id", c.Writer.Header().Get("X-Request-ID")),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ua", c.Request.UserAgent()),
		)
	}
}
