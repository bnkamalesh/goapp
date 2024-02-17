package apm

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var (
	// TimeBucketsFast suits if expected response time is very high: 1ms..100ms
	// This buckets suits for cache storages (in-memory cache, Memcache, Redis)
	TimeBucketsFast = []float64{1, 3, 7, 15, 50, 100, 200, 500, 1000, 2000, 5000}

	// TimeBucketsMedium suits for most of GO APIs, where response time is between 50ms..500ms.
	// Works for wide range of systems because provides near-logarithmic buckets distribution.
	TimeBucketsMedium = []float64{1, 5, 15, 50, 100, 250, 500, 750, 1000, 1500, 2000, 3500, 5000}

	// TimeBucketsSlow suits for relatively slow services, where expected response time is > 500ms.
	TimeBucketsSlow = []float64{50, 100, 200, 500, 750, 1000, 1250, 1500, 1750, 2000, 2500, 3000, 4000, 5000}
)

// Meter - metric service
type Meter struct {
	metric.Meter
}

// CounterAdd lazily increments certain counter metric. The label set passed on the first time should
// be present every time, no sparse keys
func (m *Meter) CounterAdd(ctx context.Context, name string, amount float64, attrs ...attribute.KeyValue) {
	counter, err := m.Float64Counter(name)
	if err != nil {
		return
	}
	counter.Add(ctx, amount, metric.WithAttributes(attrs...))
}

// HistogramRecord records value to histogram with predefined boundaries (e.g. request latency)
// The same rules apply as for counter - no sparse label structure
func (m *Meter) HistogramRecord(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	histogram, err := m.Float64Histogram(name)
	if err != nil {
		return
	}
	histogram.Record(ctx, value, metric.WithAttributes(attrs...))
}

// Observe function collect will be called each time the metric is scraped, it should be go-routine safe
func (m *Meter) Observe(name string, collect func() float64, attrs ...attribute.KeyValue) {
	gauge, err := m.Float64ObservableGauge(name)
	if err != nil {
		return
	}
	_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveFloat64(gauge, collect(), metric.WithAttributes(attrs...))
		return nil
	}, gauge)
	if err != nil {
		return
	}
}

// NewMeter create a global meter provider and a custom meter obj for the application's own usage
// we need both obj because the provider helps us integrate with other third party sdk like redis/kafka
func NewMeter(config Options, reader sdkmetric.Reader) (metric.MeterProvider, *Meter, error) {
	// to avoid high cardinality https://github.com/open-telemetry/opentelemetry-go-contrib/issues/3071
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(customViews()...),
		sdkmetric.WithResource(resource.NewWithAttributes(
			"",
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
		),
		),
	)

	meter := provider.Meter("")
	return provider, &Meter{
		Meter: meter,
	}, nil
}

func allowedAttr(v ...string) attribute.Filter {
	m := make(map[string]struct{}, len(v))
	for _, s := range v {
		m[s] = struct{}{}
	}
	return func(kv attribute.KeyValue) bool {
		_, ok := m[string(kv.Key)]
		return ok
	}
}

// Some metrics from otel pkg has high cardinality attributes, use this to filter out necessary attribute only
func customViews() []sdkmetric.View {
	return []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Scope: instrumentation.Scope{
					Name: "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
				},
			},
			sdkmetric.Stream{
				AttributeFilter: allowedAttr(
					"http.method",
					"http.status_code",
					"http.target",
				),
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Scope: instrumentation.Scope{
					Name: "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
				},
			},
			sdkmetric.Stream{
				AttributeFilter: allowedAttr(
					"rpc.service", "rpc.method", "rpc.grpc.status_code"),
			},
		),
	}
}
