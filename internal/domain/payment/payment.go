package payment

import (
	"errors"
	"time"
)

type Status string

const (
	StatusCreated    Status = "created"    // criado (antes da autorização)
	StatusAuthorized Status = "authorized" // autorizado (capturável)
	StatusCaptured   Status = "captured"   // capturado
	StatusCanceled   Status = "canceled"   // autorização cancelada
	StatusFailed     Status = "failed"     // falha
	StatusRefunded   Status = "refunded"   // reembolsado
)

type Payment struct {
	ID        string    `json:"id"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Email     string    `json:"email"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Stripe
	StripePaymentIntentID string `json:"stripe_payment_intent_id,omitempty"`
	ClientSecret          string `json:"client_secret,omitempty"`
}

func New(id string, m Money, email Email) (*Payment, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	if err := email.Validate(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Payment{
		ID:        id,
		Amount:    m.Amount,
		Currency:  string(m.Currency),
		Email:     string(email.Normalize()),
		Status:    StatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Payment) MarkAuthorized(piID, clientSecret string) error {
	if p.Status != StatusCreated && p.Status != StatusFailed {
		return errors.New("invalid state for authorization")
	}
	p.Status = StatusAuthorized
	p.StripePaymentIntentID = piID
	p.ClientSecret = clientSecret
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Payment) MarkCaptured() error {
	if p.Status != StatusAuthorized {
		return errors.New("invalid state for capture")
	}
	p.Status = StatusCaptured
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Payment) MarkCanceled() error {
	if p.Status != StatusAuthorized && p.Status != StatusCreated {
		return errors.New("invalid state for cancel")
	}
	p.Status = StatusCanceled
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Payment) MarkFailed() {
	p.Status = StatusFailed
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) MarkRefunded() error {
	if p.Status != StatusCaptured {
		return errors.New("invalid state for refund")
	}
	p.Status = StatusRefunded
	p.UpdatedAt = time.Now().UTC()
	return nil
}
