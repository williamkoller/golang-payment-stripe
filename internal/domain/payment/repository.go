package payment

type Repository interface {
	Create(p *Payment) error
	Get(id string) (*Payment, error)
	Update(p *Payment) error
	GetByPaymentIntent(piID string) (*Payment, error)
}
