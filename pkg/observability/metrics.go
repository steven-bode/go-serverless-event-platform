package observability

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type Metrics struct {
	client    *cloudwatch.Client
	logger    *Logger
	namespace string
}

func NewMetrics(client *cloudwatch.Client, logger *Logger, namespace string) *Metrics {
	return &Metrics{
		client:    client,
		logger:    logger,
		namespace: namespace,
	}
}

func (m *Metrics) PutMetric(ctx context.Context, metricName string, value float64, unit types.StandardUnit, dimensions map[string]string) error {
	dims := make([]types.Dimension, 0, len(dimensions))
	for k, v := range dimensions {
		dims = append(dims, types.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	_, err := m.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(m.namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(value),
				Unit:       unit,
				Dimensions: dims,
			},
		},
	})

	if err != nil {
		m.logger.Error("failed to put metric", err, map[string]interface{}{
			"metric_name": metricName,
		})
		return err
	}

	return nil
}

func (m *Metrics) IncrementCounter(ctx context.Context, metricName string, dimensions map[string]string) error {
	return m.PutMetric(ctx, metricName, 1.0, types.StandardUnitCount, dimensions)
}

func (m *Metrics) RecordDuration(ctx context.Context, metricName string, durationMs float64, dimensions map[string]string) error {
	return m.PutMetric(ctx, metricName, durationMs, types.StandardUnitMilliseconds, dimensions)
}
