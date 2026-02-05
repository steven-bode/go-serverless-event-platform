package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

type DynamoDBEventRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    *observability.Logger
}

func NewDynamoDBEventRepository(client *dynamodb.Client, tableName string, logger *observability.Logger) *DynamoDBEventRepository {
	return &DynamoDBEventRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
	}
}

type EventItem struct {
	EventID       string `dynamodbav:"event_id"`
	OrderID       string `dynamodbav:"order_id"`
	EventType     string `dynamodbav:"event_type"`
	Source        string `dynamodbav:"source"`
	Version       string `dynamodbav:"version"`
	CorrelationID string `dynamodbav:"correlation_id"`
	CreatedAt     string `dynamodbav:"created_at"`
	Data          string `dynamodbav:"data"`
}

func (r *DynamoDBEventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	item := EventItem{
		EventID:       event.EventID,
		OrderID:       string(event.OrderID),
		EventType:     event.EventType,
		Source:        event.Source,
		Version:       event.Version,
		CorrelationID: event.CorrelationID,
		CreatedAt:     event.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		Data:          string(event.Data),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		r.logger.Error("failed to marshal event", err)
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(event_id)"),
	})

	if err != nil {
		var condCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckErr) {
			r.logger.Warn("event already exists", map[string]interface{}{
				"event_id": event.EventID,
			})
			return domain.ErrOrderAlreadyExists
		}
		r.logger.Error("failed to save event", err, map[string]interface{}{
			"event_id": event.EventID,
		})
		return fmt.Errorf("save event: %w", err)
	}

	r.logger.Info("event saved", map[string]interface{}{
		"event_id": event.EventID,
		"order_id": event.OrderID,
	})

	return nil
}

func (r *DynamoDBEventRepository) GetEventsByOrderID(ctx context.Context, orderID domain.OrderID) ([]*domain.Event, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("order_id-index"),
		KeyConditionExpression: aws.String("order_id = :order_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":order_id": &types.AttributeValueMemberS{Value: string(orderID)},
		},
	})

	if err != nil {
		r.logger.Error("failed to query events", err, map[string]interface{}{
			"order_id": orderID,
		})
		return nil, fmt.Errorf("query events: %w", err)
	}

	var events []*domain.Event
	for _, item := range result.Items {
		var eventItem EventItem
		if err := attributevalue.UnmarshalMap(item, &eventItem); err != nil {
			r.logger.Error("failed to unmarshal event", err)
			continue
		}

		event := &domain.Event{
			EventID:       eventItem.EventID,
			CorrelationID: eventItem.CorrelationID,
			EventType:     eventItem.EventType,
			Source:        eventItem.Source,
			Version:       eventItem.Version,
			OrderID:       domain.OrderID(eventItem.OrderID),
			CreatedAt:     parseTime(eventItem.CreatedAt),
			Data:          []byte(eventItem.Data),
		}
		events = append(events, event)
	}

	return events, nil
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000Z", s)
	return t
}
