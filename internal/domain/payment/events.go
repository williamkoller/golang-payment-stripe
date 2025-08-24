package payment

type EventType string

const (
	EvtPaymentAuthorized EventType = "payment.authorized"
	EvtPaymentCaptured   EventType = "payment.captured"
	EvtPaymentFailed     EventType = "payment.failed"
	EvtPaymentCanceled   EventType = "payment.canceled"
)

type Event struct {
	Type      EventType
	PaymentID string
	Meta      map[string]string
}
