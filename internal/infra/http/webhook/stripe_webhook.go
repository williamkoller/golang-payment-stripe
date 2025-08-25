package webhook

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/williamkoller/golang-payment-stripe/internal/domain/payment"
	"go.uber.org/zap"
)

type StripeVerifier interface {
	VerifyWebhookSignature(payload []byte, sigHeader string) (stripe.Event, error)
}
type Repo interface {
	GetByPaymentIntent(piID string) (*payment.Payment, error)
	Update(p *payment.Payment) error
}

type Handler struct {
	zl *zap.Logger
	sv StripeVerifier
	r  Repo
}

func NewStripeWebhook(zl *zap.Logger, sv StripeVerifier, r Repo) *Handler {
	return &Handler{zl: zl, sv: sv, r: r}
}

func (h *Handler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	event, err := h.sv.VerifyWebhookSignature(body, c.GetHeader("Stripe-Signature"))
	if err != nil {
		h.zl.Warn("webhook_verify_failed", zap.String("err", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "payment_intent.requires_capture":
		var pi stripe.PaymentIntent
		if json.Unmarshal(event.Data.Raw, &pi) == nil {
			if p, err := h.r.GetByPaymentIntent(pi.ID); err == nil {
				_ = p.MarkAuthorized(pi.ID, pi.ClientSecret)
				_ = h.r.Update(p)
			}
		}
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if json.Unmarshal(event.Data.Raw, &pi) == nil {
			if p, err := h.r.GetByPaymentIntent(pi.ID); err == nil {
				_ = p.MarkCaptured()
				_ = h.r.Update(p)
			}
		}
	case "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if json.Unmarshal(event.Data.Raw, &pi) == nil {
			if p, err := h.r.GetByPaymentIntent(pi.ID); err == nil {
				_ = p.MarkCanceled()
				_ = h.r.Update(p)
			}
		}
	default:
		// ignore outros tipos
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *Handler) HandleTest(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	h.zl.Info("webhook_test_received", zap.String("body", string(body)))
	c.JSON(http.StatusOK, gin.H{"received": true, "test": true})
}
