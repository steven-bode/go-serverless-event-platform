package infra

import (
	"context"

	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
)

type EventRepository interface {
	SaveEvent(ctx context.Context, event *domain.Event) error
	GetEventsByOrderID(ctx context.Context, orderID domain.OrderID) ([]*domain.Event, error)
}

type EventPublisher interface {
	PublishEvent(ctx context.Context, event *domain.Event) error
}

type ReadModelRepository interface {
	SaveOrder(ctx context.Context, order *domain.Order) error
	GetOrder(ctx context.Context, orderID domain.OrderID) (*domain.Order, error)
}

type ProcessedEventsRepository interface {
	MarkAsProcessed(ctx context.Context, eventID string) error
	IsProcessed(ctx context.Context, eventID string) (bool, error)
}
