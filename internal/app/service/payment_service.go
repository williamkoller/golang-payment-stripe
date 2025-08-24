package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/williamkoller/golang-payment-stripe/internal/domain/payment"
	"github.com/williamkoller/golang-payment-stripe/pkg/ulidx"

	"go.uber.org/zap"
)

type Repo interface {
	payment.Repository
}
type Saga interface {
	Authorize(ctx context.Context, p *payment.Payment) (*payment.Payment, error)
	Capture(ctx context.Context, p *payment.Payment) (*payment.Payment, error)
	Cancel(ctx context.Context, p *payment.Payment) (*payment.Payment, error)
}

type PaymentService struct {
	zl   *zap.Logger
	repo Repo
	saga Saga
	val  *validator.Validate
}

func NewPaymentService(zl *zap.Logger, repo Repo, saga Saga) *PaymentService {
	return &PaymentService{
		zl:   zl,
		repo: repo,
		saga: saga,
		val:  validator.New(validator.WithRequiredStructEnabled()),
	}
}

type CreateInput struct {
	Amount   int64  `json:"amount" validate:"required,gt=0"`
	Currency string `json:"currency" validate:"required,alpha,len=3"`
	Email    string `json:"email" validate:"required,email"`
}

func (s *PaymentService) CreateAndAuthorize(ctx context.Context, in CreateInput) (*payment.Payment, error) {
	if err := s.val.Struct(in); err != nil {
		return nil, err
	}
	id := ulidx.New()
	m := payment.Money{Amount: in.Amount, Currency: payment.Currency(strings.ToLower(in.Currency))}
	e := payment.Email(in.Email)
	p, err := payment.New(id, m, e)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	return s.saga.Authorize(ctx, p)
}

func (s *PaymentService) Capture(ctx context.Context, id string) (*payment.Payment, error) {
	p, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	return s.saga.Capture(ctx, p)
}

func (s *PaymentService) Cancel(ctx context.Context, id string) (*payment.Payment, error) {
	p, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return s.saga.Cancel(ctx, p)
}

func (s *PaymentService) Get(ctx context.Context, id string) (*payment.Payment, error) {
	if id == "" {
		return nil, errors.New("id required")
	}
	return s.repo.Get(id)
}
