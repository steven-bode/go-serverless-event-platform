package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

type EventBridgePublisher struct {
	client  *eventbridge.Client
	busName string
	logger  *observability.Logger
}

func NewEventBridgePublisher(client *eventbridge.Client, busName string, logger *observability.Logger) *EventBridgePublisher {
	return &EventBridgePublisher{
		client:  client,
		busName: busName,
		logger:  logger,
	}
}

func (p *EventBridgePublisher) PublishEvent(ctx context.Context, event *domain.Event) error {
	detail := event.ToEventBridgeDetail()
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		p.logger.Error("failed to marshal event detail", err)
		return fmt.Errorf("marshal event detail: %w", err)
	}

	_, err = p.client.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Source:       aws.String(event.Source),
				DetailType:   aws.String(event.EventType),
				Detail:       aws.String(string(detailJSON)),
				EventBusName: aws.String(p.busName),
			},
		},
	})

	if err != nil {
		p.logger.Error("failed to publish event", err, map[string]interface{}{
			"event_id": event.EventID,
		})
		return fmt.Errorf("publish event: %w", err)
	}

	p.logger.Info("event published", map[string]interface{}{
		"event_id": event.EventID,
		"order_id": event.OrderID,
	})

	return nil
}
