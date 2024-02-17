package apm

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const environmentLabel = "environment"

// Tracer represents common wrapper on any customTracer
type Tracer struct {
	// So far no custom behavior, export all sdk for convenience
	trace.Tracer
}

// New create a global tracerProvider and a custom tracer for the application own usage
// we need both obj because tracerProvider is the way to integrate with other otel sdk
func NewTracer(ctx context.Context, opts *Options, exporter sdktrace.SpanExporter) (trace.TracerProvider, *Tracer) {
	s := &Tracer{}

	batchProcessor := sdktrace.NewBatchSpanProcessor(exporter)
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(opts.TracesSampleRate))
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithSpanProcessor(batchProcessor),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(opts.ServiceName),
			semconv.ServiceVersionKey.String(opts.ServiceVersion),
			attribute.String(environmentLabel, opts.Environment),
		)),
	)
	otel.SetTracerProvider(tp)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		b3.New(b3.WithInjectEncoding(b3.B3SingleHeader|b3.B3MultipleHeader)), // For Istio compatibility
	)
	otel.SetTextMapPropagator(propagator)

	s.Tracer = tp.Tracer(fmt.Sprintf("%s:tracer", opts.ServiceName))

	return tp, s
}
