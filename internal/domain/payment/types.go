package payment

import (
	"errors"
	"strings"
)

type Currency string

func (c Currency) String() string { return string(c) }

type Money struct {
	Amount   int64    // em centavos
	Currency Currency // "brl", "usd"
}

func (m Money) Validate() error {
	if m.Amount <= 0 {
		return errors.New("amount must be > 0")
	}
	if len(m.Currency) != 3 {
		return errors.New("currency must be 3-letter code")
	}
	return nil
}

type Email string

func (e Email) Normalize() Email {
	return Email(strings.ToLower(string(e)))
}

func (e Email) Validate() error {
	s := string(e)
	if len(s) < 3 || !strings.Contains(s, "@") {
		return errors.New("invalid email")
	}
	return nil
}
