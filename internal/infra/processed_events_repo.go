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
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

type DynamoDBProcessedEventsRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    *observability.Logger
	ttlDays   int
}

func NewDynamoDBProcessedEventsRepository(client *dynamodb.Client, tableName string, logger *observability.Logger) *DynamoDBProcessedEventsRepository {
	return &DynamoDBProcessedEventsRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
		ttlDays:   90,
	}
}

func NewDynamoDBProcessedEventsRepositoryWithTTL(client *dynamodb.Client, tableName string, logger *observability.Logger, ttlDays int) *DynamoDBProcessedEventsRepository {
	return &DynamoDBProcessedEventsRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
		ttlDays:   ttlDays,
	}
}

type ProcessedEventItem struct {
	EventID     string `dynamodbav:"event_id"`
	ProcessedAt string `dynamodbav:"processed_at"`
	TTL         int64  `dynamodbav:"ttl,omitempty"`
}

func (r *DynamoDBProcessedEventsRepository) MarkAsProcessed(ctx context.Context, eventID string) error {
	now := time.Now().UTC()
	ttl := now.AddDate(0, 0, r.ttlDays).Unix()

	item := ProcessedEventItem{
		EventID:     eventID,
		ProcessedAt: now.Format(time.RFC3339),
		TTL:         ttl,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		r.logger.Error("failed to marshal processed event", err)
		return fmt.Errorf("marshal processed event: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(event_id)"),
	})

	if err != nil {
		var condCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckErr) {
			r.logger.Warn("event already processed", map[string]interface{}{
				"event_id": eventID,
			})
			return nil
		}
		r.logger.Error("failed to mark event as processed", err, map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("mark event as processed: %w", err)
	}

	r.logger.Info("event marked as processed", map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

func (r *DynamoDBProcessedEventsRepository) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"event_id": &types.AttributeValueMemberS{Value: eventID},
		},
	})

	if err != nil {
		r.logger.Error("failed to check if event is processed", err, map[string]interface{}{
			"event_id": eventID,
		})
		return false, fmt.Errorf("check processed event: %w", err)
	}

	return result.Item != nil, nil
}
