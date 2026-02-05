package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/stevenbode/go-serverless-event-platform/internal/app"
	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
	"github.com/stevenbode/go-serverless-event-platform/internal/infra"
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

var useCase *app.CreateOrderUseCase

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	eventbridgeClient := eventbridge.NewFromConfig(cfg)
	cloudwatchClient := cloudwatch.NewFromConfig(cfg)

	eventStoreTable := getEnv("EVENT_STORE_TABLE", "event_store")
	eventBusName := getEnv("EVENT_BUS_NAME", "app-bus")
	logLevel := getEnv("LOG_LEVEL", "ERROR")

	logger := observability.NewLoggerWithLevel("", "", observability.LogLevel(logLevel))
	metrics := observability.NewMetrics(cloudwatchClient, logger, "EventPlatform")

	eventRepo := infra.NewDynamoDBEventRepository(
		dynamoClient,
		eventStoreTable,
		logger,
	)

	publisher := infra.NewEventBridgePublisher(
		eventbridgeClient,
		eventBusName,
		logger,
	)

	useCase = app.NewCreateOrderUseCase(
		eventRepo,
		publisher,
		logger,
		metrics,
	)
}

type CreateOrderRequest struct {
	OrderID    string `json:"order_id,omitempty"`
	CustomerID string `json:"customer_id"`
	TotalCents int64  `json:"total_cents"`
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	correlationID := observability.GetOrGenerateCorrelationID(req.Headers["x-correlation-id"])
	logger := observability.NewLogger(correlationID, "")

	var createReq CreateOrderRequest
	if err := json.Unmarshal([]byte(req.Body), &createReq); err != nil {
		logger.Error("failed to parse request body", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       `{"error":"invalid request body"}`,
		}, nil
	}

	order, err := useCase.Execute(ctx, app.CreateOrderRequest{
		OrderID:    createReq.OrderID,
		CustomerID: createReq.CustomerID,
		TotalCents: createReq.TotalCents,
	}, correlationID)

	if err != nil {
		httpStatus := 500
		if appErr, ok := err.(*domain.AppError); ok {
			httpStatus = appErr.HTTPStatus
		}
		logger.Error("failed to create order", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: httpStatus,
			Body:       fmt.Sprintf(`{"error":"%s"}`, err.Error()),
			Headers: map[string]string{
				"X-Correlation-Id": correlationID,
			},
		}, nil
	}

	responseBody, _ := json.Marshal(map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"total_cents": order.TotalCents,
		"created_at":  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 201,
		Body:       string(responseBody),
		Headers: map[string]string{
			"Content-Type":     "application/json",
			"X-Correlation-Id": correlationID,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	lambda.Start(handler)
}
