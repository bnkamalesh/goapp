// Package apm combines traces/metrics/logs and error aggregators to provide app observability
package apm

import (
	"context"
	"strings"
	"time"

	"github.com/bnkamalesh/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/sync/errgroup"
)

const shutdownTimeout = 2 * time.Second

// APM is the application performance monitoring service
type APM struct {
	appTracer      *Tracer
	tracerProvider trace.TracerProvider

	appMeter      *Meter
	meterProvider metric.MeterProvider
}

// global apm instance, to simplify code/minimize injections
var (
	global *APM
)

// Options used for apm initialization
type Options struct {
	Debug                bool
	Environment          string
	ServiceName          string
	ServiceVersion       string
	TracesSampleRate     float64
	CollectorURL         string
	PrometheusScrapePort uint16
	// UseStdOut if true, will set the metrics exporter and trace exporter as stdout
	UseStdOut bool
}

// New initializes APM service using options provided
// It has an internal tracer and meter created for the application
// It can also access the tracerprovider and meteterprovider, so we can integrate otel with other client (redis/kakfa...)
func New(ctx context.Context, opts *Options) (*APM, error) {
	s := &APM{}

	tracerProvider, tr, err := newTracer(ctx, opts)
	if err != nil {
		return nil, err
	}
	s.tracerProvider = tracerProvider
	s.appTracer = tr

	mProvider, m, err := newMeter(opts)
	if err != nil {
		return nil, err
	}

	s.appMeter = m
	s.meterProvider = mProvider
	SetGlobal(s)

	return s, nil
}

// Shutdown gracefully switch off apm, flushing any data it have
func (s *APM) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	g := errgroup.Group{}
	if s.tracerProvider != nil {
		g.Go(func() error {
			if tp, ok := s.tracerProvider.(*sdktrace.TracerProvider); ok {
				return tp.Shutdown(ctx)
			}
			return nil
		})
	}

	if s.meterProvider != nil {
		g.Go(func() error {
			if mp, ok := s.meterProvider.(*sdkmetric.MeterProvider); ok {
				return mp.Shutdown(ctx)
			}
			return nil
		})
	}

	return g.Wait()
}

// AppTracer gets provided appTracer for traces
func (s *APM) AppTracer() *Tracer {
	if s == nil {
		return nil
	}
	return s.appTracer
}

// Use this to integrate otel with other client pkg (redis/kafka)
func (s *APM) GetTracerProvider() trace.TracerProvider {
	if s.tracerProvider == nil {
		return noop.NewTracerProvider()
	}
	return s.tracerProvider
}

// Use this to integrate otel with other client pkg (redis/kafka)
func (s *APM) GetMeterProvider() metric.MeterProvider {
	if s.meterProvider == nil {
		return sdkmetric.NewMeterProvider()
	}
	return s.meterProvider
}

// AppMeter gets provided appMeter for metrics
func (s *APM) AppMeter() *Meter {
	if s == nil {
		return nil
	}
	return s.appMeter
}

// SetGlobal sets global apm instance
func SetGlobal(apm *APM) {
	global = apm
}

// Global gets global apm instance
func Global() *APM {
	if global == nil {
		apm, _ := New(context.Background(), &Options{UseStdOut: false})
		global = apm
		return apm
	}
	return global
}

func newMeter(opts *Options) (metric.MeterProvider, *Meter, error) {
	exp, err := stdoutmetric.New()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed initializing stdout metric exporter")
	}

	var mReader sdkmetric.Reader
	if opts.UseStdOut {
		mReader = sdkmetric.NewPeriodicReader(
			exp,
			sdkmetric.WithInterval(time.Second*10),
		)
	} else {
		pexp, err := prometheusExporter()
		if err != nil {
			return nil, nil, err
		}
		mReader = pexp
		// uncomment below to start prometheusScraper if required
		go prometheusScraper(opts)
	}

	return NewMeter(
		Options{
			ServiceName:    opts.ServiceName,
			ServiceVersion: opts.ServiceVersion,
		},
		mReader,
	)
}

func newTracer(ctx context.Context, opts *Options) (trace.TracerProvider, *Tracer, error) {
	var (
		exporter      sdktrace.SpanExporter
		err           error
		httpCollector = strings.HasPrefix(opts.CollectorURL, "http")
	)

	if opts.UseStdOut {
		exporter, err = stdouttrace.New()
	} else if httpCollector {
		exporter, err = otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpoint(opts.CollectorURL),
			otlptracehttp.WithInsecure(),
		)
	} else {
		exporter, err = otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(opts.CollectorURL),
			otlptracegrpc.WithInsecure(),
		)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize trace exporter")
	}

	tp, t := NewTracer(ctx, opts, exporter)
	return tp, t, nil
}
