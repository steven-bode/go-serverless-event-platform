package app

import (
	"context"
	"time"

	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
	"github.com/stevenbode/go-serverless-event-platform/internal/infra"
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

type ApplyOrderCreatedUseCase struct {
	readModelRepo       infra.ReadModelRepository
	processedEventsRepo infra.ProcessedEventsRepository
	logger              *observability.Logger
	metrics             *observability.Metrics
}

func NewApplyOrderCreatedUseCase(
	readModelRepo infra.ReadModelRepository,
	processedEventsRepo infra.ProcessedEventsRepository,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *ApplyOrderCreatedUseCase {
	return &ApplyOrderCreatedUseCase{
		readModelRepo:       readModelRepo,
		processedEventsRepo: processedEventsRepo,
		logger:              logger,
		metrics:             metrics,
	}
}

type OrderCreatedEventDetail struct {
	EventID       string `json:"event_id"`
	CorrelationID string `json:"correlation_id"`
	OrderID       string `json:"order_id"`
	CustomerID    string `json:"customer_id"`
	TotalCents    int64  `json:"total_cents"`
	CreatedAt     string `json:"created_at"`
}

func (uc *ApplyOrderCreatedUseCase) Execute(ctx context.Context, detail OrderCreatedEventDetail) error {
	start := time.Now()
	defer func() {
		if uc.metrics != nil {
			duration := time.Since(start).Milliseconds()
			uc.metrics.RecordDuration(ctx, "apply_order_created_duration_ms", float64(duration), map[string]string{
				"correlation_id": detail.CorrelationID,
			})
		}
	}()

	processed, err := uc.processedEventsRepo.IsProcessed(ctx, detail.EventID)
	if err != nil {
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "apply_order_created_check_errors", map[string]string{
				"correlation_id": detail.CorrelationID,
			})
		}
		uc.logger.Error("failed to check if event is processed", err, map[string]interface{}{
			"event_id": detail.EventID,
		})
		return domain.NewRetriableError(err, "failed to check processed status")
	}

	if processed {
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "apply_order_created_idempotency_hits", map[string]string{
				"correlation_id": detail.CorrelationID,
			})
		}
		return nil
	}

	createdAt, err := time.Parse(time.RFC3339, detail.CreatedAt)
	if err != nil {
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "apply_order_created_parse_errors", map[string]string{
				"correlation_id": detail.CorrelationID,
			})
		}
		uc.logger.Error("failed to parse created_at", err, map[string]interface{}{
			"event_id": detail.EventID,
		})
		return domain.NewNonRetriableError(err, "invalid created_at format")
	}

	order := &domain.Order{
		ID:         domain.OrderID(detail.OrderID),
		CustomerID: domain.CustomerID(detail.CustomerID),
		TotalCents: detail.TotalCents,
		CreatedAt:  createdAt,
	}

	if err := uc.readModelRepo.SaveOrder(ctx, order); err != nil {
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "apply_order_created_read_model_errors", map[string]string{
				"correlation_id": detail.CorrelationID,
			})
		}
		uc.logger.Error("failed to save order to read model", err, map[string]interface{}{
			"order_id": order.ID,
			"event_id": detail.EventID,
		})
		return domain.NewRetriableError(err, "failed to save order")
	}

	if err := uc.processedEventsRepo.MarkAsProcessed(ctx, detail.EventID); err != nil {
		if uc.metrics != nil {
			uc.metrics.IncrementCounter(ctx, "apply_order_created_mark_processed_errors", map[string]string{
				"correlation_id": detail.CorrelationID,
			})
		}
		uc.logger.Error("failed to mark event as processed", err, map[string]interface{}{
			"event_id": detail.EventID,
		})
		return domain.NewRetriableError(err, "failed to mark event as processed")
	}

	if uc.metrics != nil {
		uc.metrics.IncrementCounter(ctx, "apply_order_created_success", map[string]string{
			"correlation_id": detail.CorrelationID,
		})
	}

	return nil
}
