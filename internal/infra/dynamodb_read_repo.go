package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

type DynamoDBReadModelRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    *observability.Logger
}

func NewDynamoDBReadModelRepository(client *dynamodb.Client, tableName string, logger *observability.Logger) *DynamoDBReadModelRepository {
	return &DynamoDBReadModelRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
	}
}

type OrderItem struct {
	OrderID    string `dynamodbav:"order_id"`
	CustomerID string `dynamodbav:"customer_id"`
	TotalCents int64  `dynamodbav:"total_cents"`
	CreatedAt  string `dynamodbav:"created_at"`
}

func (r *DynamoDBReadModelRepository) SaveOrder(ctx context.Context, order *domain.Order) error {
	item := OrderItem{
		OrderID:    string(order.ID),
		CustomerID: string(order.CustomerID),
		TotalCents: order.TotalCents,
		CreatedAt:  order.CreatedAt.Format(time.RFC3339),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		r.logger.Error("failed to marshal order", err)
		return fmt.Errorf("marshal order: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})

	if err != nil {
		r.logger.Error("failed to save order", err, map[string]interface{}{
			"order_id": order.ID,
		})
		return fmt.Errorf("save order: %w", err)
	}

	r.logger.Info("order saved to read model", map[string]interface{}{
		"order_id": order.ID,
	})

	return nil
}

func (r *DynamoDBReadModelRepository) GetOrder(ctx context.Context, orderID domain.OrderID) (*domain.Order, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"order_id": &types.AttributeValueMemberS{Value: string(orderID)},
		},
	})

	if err != nil {
		r.logger.Error("failed to get order", err, map[string]interface{}{
			"order_id": orderID,
		})
		return nil, fmt.Errorf("get order: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var item OrderItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		r.logger.Error("failed to unmarshal order", err)
		return nil, fmt.Errorf("unmarshal order: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)

	return &domain.Order{
		ID:         domain.OrderID(item.OrderID),
		CustomerID: domain.CustomerID(item.CustomerID),
		TotalCents: item.TotalCents,
		CreatedAt:  createdAt,
	}, nil
}
