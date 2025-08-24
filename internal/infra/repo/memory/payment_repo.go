package memory

import (
	"errors"
	"sync"
	"time"

	"github.com/williamkoller/golang-payment-stripe/internal/domain/payment"
)

type PaymentRepo struct {
	mu   sync.RWMutex
	byID map[string]*payment.Payment
	byPI map[string]string // piID -> paymentID
}

func NewPaymentRepo() *PaymentRepo {
	return &PaymentRepo{
		byID: make(map[string]*payment.Payment),
		byPI: make(map[string]string),
	}
}

func (r *PaymentRepo) Create(p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[p.ID]; ok {
		return errors.New("payment already exists")
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	r.byID[p.ID] = clone(p)
	if p.StripePaymentIntentID != "" {
		r.byPI[p.StripePaymentIntentID] = p.ID
	}
	return nil
}

func (r *PaymentRepo) Get(id string) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return clone(p), nil
}

func (r *PaymentRepo) Update(p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.byID[p.ID]
	if !ok {
		return errors.New("not found")
	}
	p.UpdatedAt = time.Now().UTC()
	r.byID[p.ID] = clone(p)
	if p.StripePaymentIntentID != "" {
		r.byPI[p.StripePaymentIntentID] = p.ID
	}
	return nil
}

func (r *PaymentRepo) GetByPaymentIntent(piID string) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byPI[piID]
	if !ok {
		return nil, errors.New("not found")
	}
	return clone(r.byID[id]), nil
}

func clone(p *payment.Payment) *payment.Payment {
	cp := *p
	return &cp
}
