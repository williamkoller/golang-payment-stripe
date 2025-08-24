package saga

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/williamkoller/golang-payment-stripe/internal/app/ports"
	"github.com/williamkoller/golang-payment-stripe/internal/domain/payment"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/config"
	"go.uber.org/zap"
)

type report interface {
	Update(p *payment.Payment) error
}

type PaymentSaga struct {
	zl   *zap.Logger
	repo report
	pg   ports.PaymentGateway
	cfg  *config.Config
}

func NewPaymentSaga(zl *zap.Logger,
	repo report,
	pg ports.PaymentGateway,
	cfg *config.Config) *PaymentSaga {
	return &PaymentSaga{
		zl,
		repo,
		pg,
		cfg,
	}
}

func (s *PaymentSaga) Authorize(ctx context.Context, p *payment.Payment) (*payment.Payment, error) {
	if p.Status != payment.StatusCreated && p.Status != payment.StatusFailed {
		return nil, errors.New("invalid status for authorize")
	}

	if p.Amount >= 10_000_000 {
		p.MarkFailed()
		_ = s.repo.Update(p)
		return nil, errors.New("risk? amount too high")
	}

	idem := fmt.Sprintf("auth-%s", p.ID)
	piID, clientSecret, err := s.pg.AuthorizeManual(ctx, idem, p.Amount, p.Currency, p.Email, s.cfg.StripeEnableTestPM, s.cfg.StripeTestPaymentPM)
	if err != nil {
		p.MarkFailed()
		_ = s.repo.Update(p)
		return nil, err
	}

	_ = p.MarkAuthorized(piID, clientSecret)
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PaymentSaga) Capture(ctx context.Context, p *payment.Payment) (*payment.Payment, error) {
	if p.Status != payment.StatusAuthorized {
		return nil, errors.New("payment is not authorized")
	}

	if err := s.pg.Capture(ctx, p.StripePaymentIntentID); err != nil {
		p.MarkFailed()
		_ = s.repo.Update(p)
		return nil, err
	}

	select {
	case <-ctx.Done():
		_ = s.pg.Refund(context.Background(), p.StripePaymentIntentID)
		p.MarkFailed()
		_ = s.repo.Update(p)
		return nil, ctx.Err()
	case <-time.After(50 * time.Millisecond):
	}

	_ = p.MarkCaptured()

	if err := s.repo.Update(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PaymentSaga) Cancel(ctx context.Context, p *payment.Payment) (*payment.Payment, error) {
	if p.Status != payment.StatusAuthorized && p.Status != payment.StatusCreated {
		return nil, errors.New("invalid status for cancel")
	}
	if p.StripePaymentIntentID != "" {
		_ = s.pg.Cancel(ctx, p.StripePaymentIntentID)
	}
	_ = p.MarkCanceled()
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}
