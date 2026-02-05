package infra

import (
	"context"
	"testing"

	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
)

type MockEventRepository struct {
	events map[string]bool
}

func NewMockEventRepository() *MockEventRepository {
	return &MockEventRepository{
		events: make(map[string]bool),
	}
}

func (m *MockEventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if m.events[event.EventID] {
		return domain.ErrOrderAlreadyExists
	}
	m.events[event.EventID] = true
	return nil
}

func (m *MockEventRepository) GetEventsByOrderID(ctx context.Context, orderID domain.OrderID) ([]*domain.Event, error) {
	return nil, nil
}

type MockProcessedEventsRepository struct {
	processed map[string]bool
}

func NewMockProcessedEventsRepository() *MockProcessedEventsRepository {
	return &MockProcessedEventsRepository{
		processed: make(map[string]bool),
	}
}

func (m *MockProcessedEventsRepository) MarkAsProcessed(ctx context.Context, eventID string) error {
	if m.processed[eventID] {
		return nil
	}
	m.processed[eventID] = true
	return nil
}

func (m *MockProcessedEventsRepository) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	return m.processed[eventID], nil
}

func TestIdempotency_EventRepository(t *testing.T) {
	repo := NewMockEventRepository()
	ctx := context.Background()

	event := &domain.Event{
		EventID:       "event-123",
		CorrelationID: "corr-123",
		EventType:     domain.EventTypeOrderCreated,
		Source:        domain.EventSourceOrders,
		OrderID:       "order-123",
		CustomerID:    "customer-456",
		TotalCents:    10000,
	}

	err := repo.SaveEvent(ctx, event)
	if err != nil {
		t.Errorf("unexpected error on first save: %v", err)
	}

	err = repo.SaveEvent(ctx, event)
	if err != domain.ErrOrderAlreadyExists {
		t.Errorf("expected ErrOrderAlreadyExists on duplicate save, got: %v", err)
	}
}

func TestIdempotency_ProcessedEventsRepository(t *testing.T) {
	repo := NewMockProcessedEventsRepository()
	ctx := context.Background()

	eventID := "event-123"

	processed, err := repo.IsProcessed(ctx, eventID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if processed {
		t.Errorf("expected event not to be processed initially")
	}

	err = repo.MarkAsProcessed(ctx, eventID)
	if err != nil {
		t.Errorf("unexpected error on mark as processed: %v", err)
	}

	processed, err = repo.IsProcessed(ctx, eventID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !processed {
		t.Errorf("expected event to be processed after marking")
	}

	err = repo.MarkAsProcessed(ctx, eventID)
	if err != nil {
		t.Errorf("marking already processed event should not error: %v", err)
	}
}
