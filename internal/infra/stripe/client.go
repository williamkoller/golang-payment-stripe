package stripeinfra

import (
	"context"
	"errors"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"
	"github.com/williamkoller/golang-payment-stripe/internal/app/ports"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/config"
	"go.uber.org/zap"
)

type client struct {
	zl      *zap.Logger
	cfg     *config.Config
	breaker *gobreaker.CircuitBreaker
}

func NewClient(cfg *config.Config, zl *zap.Logger) ports.PaymentGateway {
	stripe.Key = cfg.StripeSecretKey
	st := gobreaker.Settings{
		Name:        "stripe",
		MaxRequests: cfg.CBMaxRequests,
		Interval:    cfg.CBInterval,
		Timeout:     cfg.CBTimeout,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.Requests >= 10 && float64(c.TotalFailures)/float64(c.Requests) >= 0.6
		},
	}
	return &client{zl: zl, cfg: cfg, breaker: gobreaker.NewCircuitBreaker(st)}
}

func (c *client) AuthorizeManual(ctx context.Context, idemKey string, amount int64, currency, email string, useTest bool, testPM string) (string, string, error) {
	if c.cfg.StripeSecretKey == "" {
		return "", "", errors.New("stripe secret key not configured")
	}
	res, err := c.exec(ctx, func() (any, error) {
		params := &stripe.PaymentIntentParams{
			Amount:        stripe.Int64(amount),
			Currency:      stripe.String(currency),
			ReceiptEmail:  stripe.String(email),
			CaptureMethod: stripe.String(string(stripe.PaymentIntentCaptureMethodManual)),
			Confirm:       stripe.Bool(true),
			AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
				Enabled:        stripe.Bool(true),
				AllowRedirects: stripe.String(string(stripe.PaymentIntentAutomaticPaymentMethodsAllowRedirectsNever)),
			},
		}

		if useTest {
			params.PaymentMethod = stripe.String(testPM)
		}

		params.SetIdempotencyKey(idemKey)

		return paymentintent.New(params)
	})
	if err != nil {
		return "", "", err
	}

	pi := res.(*stripe.PaymentIntent)
	return pi.ID, pi.ClientSecret, nil
}

func (c *client) Capture(ctx context.Context, piID string) error {
	_, err := c.exec(ctx, func() (any, error) {
		return paymentintent.Capture(piID, &stripe.PaymentIntentCaptureParams{})
	})
	return err
}

func (c *client) Cancel(ctx context.Context, piID string) error {
	_, err := c.exec(ctx, func() (any, error) {
		return paymentintent.Cancel(piID, &stripe.PaymentIntentCancelParams{})
	})
	return err
}

func (c *client) Refund(ctx context.Context, piID string) error {
	_, err := c.exec(ctx, func() (any, error) {
		return refund.New(&stripe.RefundParams{PaymentIntent: stripe.String(piID)})
	})
	return err
}

func (c *client) VerifyWebhookSignature(payload []byte, sigHeader string) (stripe.Event, error) {
	if c.cfg.StripeWebhookSecret == "" {
		return stripe.Event{}, errors.New("webhook signing secret not configured")
	}

	return webhook.ConstructEvent(payload, sigHeader, c.cfg.StripeWebhookSecret)
}

func (c *client) exec(ctx context.Context, fn func() (any, error)) (any, error) {
	type result struct {
		v   any
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := c.breaker.Execute(fn)
		ch <- result{v, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.v, r.err
	case <-time.After(c.cfg.RequestTimeout):
		return nil, errors.New("stripe call timeout")
	}
}
