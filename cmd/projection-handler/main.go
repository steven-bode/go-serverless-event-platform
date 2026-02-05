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
	"github.com/stevenbode/go-serverless-event-platform/internal/app"
	"github.com/stevenbode/go-serverless-event-platform/internal/domain"
	"github.com/stevenbode/go-serverless-event-platform/internal/infra"
	"github.com/stevenbode/go-serverless-event-platform/pkg/observability"
)

var useCase *app.ApplyOrderCreatedUseCase

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	cloudwatchClient := cloudwatch.NewFromConfig(cfg)

	ordersReadTable := getEnv("ORDERS_READ_TABLE", "orders_read")
	processedEventsTable := getEnv("PROCESSED_EVENTS_TABLE", "processed_events")
	logLevel := getEnv("LOG_LEVEL", "ERROR")

	logger := observability.NewLoggerWithLevel("", "", observability.LogLevel(logLevel))
	metrics := observability.NewMetrics(cloudwatchClient, logger, "EventPlatform")

	readModelRepo := infra.NewDynamoDBReadModelRepository(
		dynamoClient,
		ordersReadTable,
		logger,
	)

	processedEventsRepo := infra.NewDynamoDBProcessedEventsRepository(
		dynamoClient,
		processedEventsTable,
		logger,
	)

	useCase = app.NewApplyOrderCreatedUseCase(
		readModelRepo,
		processedEventsRepo,
		logger,
		metrics,
	)
}

func handler(ctx context.Context, event events.EventBridgeEvent) error {
	var detail app.OrderCreatedEventDetail
	if err := json.Unmarshal([]byte(event.Detail), &detail); err != nil {
		return fmt.Errorf("failed to unmarshal event detail: %w", err)
	}

	logger := observability.NewLogger(detail.CorrelationID, detail.EventID)

	if err := useCase.Execute(ctx, detail); err != nil {
		logger.Error("failed to apply order created event", err, map[string]interface{}{
			"source":      event.Source,
			"detail_type": event.DetailType,
		})
		if domain.IsRetriable(err) {
			return err
		}
		return nil
	}

	logger.Info("event processed", map[string]interface{}{
		"source":      event.Source,
		"detail_type": event.DetailType,
	})

	return nil
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
