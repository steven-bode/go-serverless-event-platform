package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
	"github.com/stevenbode/go-serverless-event-platform/internal/infra"
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

type CreateOrderUseCase struct {
	eventRepo infra.EventRepository
	publisher infra.EventPublisher
	logger    *observability.Logger
	metrics   *observability.Metrics
}

func NewCreateOrderUseCase(eventRepo infra.EventRepository, publisher infra.EventPublisher, logger *observability.Logger, metrics *observability.Metrics) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		eventRepo: eventRepo,
		publisher: publisher,
		logger:    logger,
		metrics:   metrics,
	}
}

type CreateOrderRequest struct {
	OrderID    string
	CustomerID string
	TotalCents int64
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest, correlationID string) (*domain.Order, error) {
	start := time.Now()
	defer func() {
		if uc.metrics != nil {
			duration := time.Since(start).Milliseconds()
			uc.metrics.RecordDuration(ctx, "create_order_duration_ms", float64(duration), map[string]string{
				"correlation_id": correlationID,
			})
		}
	}()

	orderID := domain.OrderID(req.OrderID)
	if orderID == "" {
		orderID = domain.OrderID(uuid.New().String())
	}

	order, err := domain.NewOrder(orderID, domain.CustomerID(req.CustomerID), req.TotalCents)
	if err != nil {
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "create_order_validation_errors", map[string]string{
				"correlation_id": correlationID,
			})
		}
		uc.logger.Error("validation failed", err, map[string]interface{}{
			"order_id":    orderID,
			"customer_id": req.CustomerID,
		})
		return nil, domain.NewValidationError(err, fmt.Sprintf("invalid order: %v", err))
	}

	eventID := uuid.New().String()
	event := domain.NewOrderCreatedEvent(eventID, correlationID, order)

	if err := uc.eventRepo.SaveEvent(ctx, event); err != nil {
		if err == domain.ErrOrderAlreadyExists {
			if uc.metrics != nil {
				uc.metrics.IncrementCounter(ctx, "create_order_idempotency_hits", map[string]string{
					"correlation_id": correlationID,
				})
			}
			uc.logger.Warn("order already exists", map[string]interface{}{
				"order_id": orderID,
				"event_id": eventID,
			})
			return nil, domain.NewNonRetriableError(err, "order already exists")
		}
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "create_order_event_store_errors", map[string]string{
				"correlation_id": correlationID,
			})
		}
		uc.logger.Error("failed to save event", err, map[string]interface{}{
			"order_id": orderID,
			"event_id": eventID,
		})
		return nil, domain.NewRetriableError(err, "failed to save event")
	}

	if err := uc.publisher.PublishEvent(ctx, event); err != nil {
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "create_order_publish_errors", map[string]string{
				"correlation_id": correlationID,
			})
		}
		uc.logger.Error("failed to publish event", err, map[string]interface{}{
			"order_id": orderID,
			"event_id": eventID,
		})
		return nil, domain.NewRetriableError(err, "failed to publish event")
	}

	if uc.metrics != nil {
		uc.metrics.IncrementCounter(ctx, "create_order_success", map[string]string{
			"correlation_id": correlationID,
		})
	}

	return order, nil
}
