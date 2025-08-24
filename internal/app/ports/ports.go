package ports

import (
	"context"

	"github.com/stripe/stripe-go/v76"
)

type PaymentGateway interface {
	AuthorizeManual(ctx context.Context, idemKey string, amount int64, current, email string, useTest bool, testPM string) (piID, clientSecret string, err error)
	Capture(ctx context.Context, paymentIntendID string) error
	Cancel(ctx context.Context, paymentIntendID string) error
	Refund(ctx context.Context, paymentIntendID string) error
	VerifyWebhookSignature(payload []byte, sigHEader string) (stripe.Event, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}
