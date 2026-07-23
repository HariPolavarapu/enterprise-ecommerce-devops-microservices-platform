package events

const (
	TopicOrderCreated      = "order-created"
	TopicPaymentSuccess    = "payment-success"
	TopicPaymentFailed     = "payment-failed"
	TopicInventoryUpdated  = "inventory-updated"
	TopicShipmentCreated   = "shipment-created"
	TopicNotificationReq   = "notification-requested"
	TopicOrderCancelled    = "order-cancelled"
)

type EventPublisher interface {
	Publish(topic string, payload interface{}) error
}

type InMemoryEventPublisher struct {
	messages []string
}

func NewInMemoryEventPublisher() *InMemoryEventPublisher {
	return &InMemoryEventPublisher{}
}

func (p *InMemoryEventPublisher) Publish(topic string, payload interface{}) error {
	p.messages = append(p.messages, topic)
	_ = payload
	return nil
}
