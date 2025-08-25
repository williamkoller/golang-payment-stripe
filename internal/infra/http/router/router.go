package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/williamkoller/golang-payment-stripe/internal/app/service"
	"github.com/williamkoller/golang-payment-stripe/internal/domain/payment"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/config"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/http/handlers"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/http/middleware"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/http/webhook"
	"go.uber.org/zap"
)

// Interfaces mínimas para reduzir acoplamento e evitar type-assert confuso
type StripeVerifier interface {
	VerifyWebhookSignature(payload []byte, sigHeader string) (stripe.Event, error)
}
type PaymentRepo interface {
	GetByPaymentIntent(piID string) (*payment.Payment, error)
	Update(p *payment.Payment) error
}

func Build(
	zl *zap.Logger,
	cfg *config.Config,
	svc *service.PaymentService,
	sv StripeVerifier,
	repo PaymentRepo,
) *gin.Engine {

	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.SecurityHeaders(),
		middleware.RateLimit(cfg),
		middleware.Timeout(cfg, zl),
		middleware.GinZapLogger(zl),
	)

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Payments
	ph := handlers.NewPaymentHandler(svc)
	r.POST("/v1/payments", ph.Create)
	r.GET("/v1/payments/:id", ph.Get)
	r.POST("/v1/payments/:id/capture", ph.Capture)
	r.POST("/v1/payments/:id/cancel", ph.Cancel)

	// Webhook Stripe
	wh := webhook.NewStripeWebhook(zl, sv, repo)
	r.POST("/v1/webhooks/stripe", wh.Handle)

	// Endpoint de teste para webhook (remover em produção)
	if cfg.Env != "prod" {
		r.POST("/v1/webhooks/stripe/test", wh.HandleTest)
	}

	return r
}
