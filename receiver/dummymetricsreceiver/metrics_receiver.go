package dummymetricsreceiver

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type dummyMetricsReceiver struct {
	config       *Config
	nextConsumer consumer.Metrics
	settings     *receiver.Settings

	cancel context.CancelFunc
}

func (r *dummyMetricsReceiver) Start(_ context.Context, _ component.Host) error {
	// Create a new context as specified in the interface documentation
	ctx := context.Background()
	ctx, r.cancel = context.WithCancel(ctx)

	interval, err := time.ParseDuration(r.config.Interval)
	if err != nil {
		return fmt.Errorf("failed to parse interval: %w", err)

	}

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				md, err := r.generateMetric()
				if err != nil {
					r.settings.Logger.Error("Failed to generate metric", zap.Error(err))
					continue
				}
				// nolint:errcheck //
				r.nextConsumer.ConsumeMetrics(ctx, md)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *dummyMetricsReceiver) generateMetric() (pmetric.Metrics, error) {
	host, err := os.Hostname()
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("failed to get hostname: %w", err)
	}

	md := pmetric.NewMetrics()
	resourceMetrics := md.ResourceMetrics().AppendEmpty()
	resourceMetrics.Resource().Attributes().PutStr("k8s.cluster.name", "test-cluster")
	metric := resourceMetrics.
		ScopeMetrics().
		AppendEmpty().
		Metrics().
		AppendEmpty()

	metric.SetName("dummy")
	gauge := metric.SetEmptyGauge()
	for i := 0; i < 5; i++ {
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetIntValue(int64(i))
		dp.Attributes().PutStr("host", host)
	}

	return md, nil
}

func (r *dummyMetricsReceiver) Shutdown(_ context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
