package domain

import (
	"encoding/json"
	"time"
)

const (
	EventTypeOrderCreated = "OrderCreated"
	EventSourceOrders     = "app.orders"
	EventVersionV1        = "1.0"
)

type Event struct {
	EventID       string          `json:"event_id"`
	CorrelationID string          `json:"correlation_id"`
	EventType     string          `json:"event_type"`
	Source        string          `json:"source"`
	Version       string          `json:"version"`
	OrderID       OrderID         `json:"order_id"`
	CustomerID    CustomerID      `json:"customer_id"`
	TotalCents    int64           `json:"total_cents"`
	CreatedAt     time.Time       `json:"created_at"`
	Data          json.RawMessage `json:"data,omitempty"`
}

type OrderCreatedEvent struct {
	EventID       string `json:"event_id"`
	CorrelationID string `json:"correlation_id"`
	OrderID       string `json:"order_id"`
	CustomerID    string `json:"customer_id"`
	TotalCents    int64  `json:"total_cents"`
	CreatedAt     string `json:"created_at"`
	Version       string `json:"version"`
}

func NewOrderCreatedEvent(eventID, correlationID string, order *Order) *Event {
	orderCreated := OrderCreatedEvent{
		EventID:       eventID,
		CorrelationID: correlationID,
		OrderID:       string(order.ID),
		CustomerID:    string(order.CustomerID),
		TotalCents:    order.TotalCents,
		CreatedAt:     order.CreatedAt.Format(time.RFC3339),
		Version:       EventVersionV1,
	}

	data, _ := json.Marshal(orderCreated)

	return &Event{
		EventID:       eventID,
		CorrelationID: correlationID,
		EventType:     EventTypeOrderCreated,
		Source:        EventSourceOrders,
		Version:       EventVersionV1,
		OrderID:       order.ID,
		CustomerID:    order.CustomerID,
		TotalCents:    order.TotalCents,
		CreatedAt:     order.CreatedAt,
		Data:          data,
	}
}

func (e *Event) ToEventBridgeDetail() OrderCreatedEvent {
	return OrderCreatedEvent{
		EventID:       e.EventID,
		CorrelationID: e.CorrelationID,
		OrderID:       string(e.OrderID),
		CustomerID:    string(e.CustomerID),
		TotalCents:    e.TotalCents,
		CreatedAt:     e.CreatedAt.Format(time.RFC3339),
		Version:       e.Version,
	}
}
