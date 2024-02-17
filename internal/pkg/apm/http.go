package apm

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HTTPMiddleware func(h http.Handler) http.Handler

func NewHTTPMiddleware(opts ...otelhttp.Option) HTTPMiddleware {
	if len(opts) == 0 {
		opts = make([]otelhttp.Option, 0, 3)
	}
	gb := Global()
	opts = append(opts,
		otelhttp.WithMeterProvider(gb.GetMeterProvider()),
		otelhttp.WithTracerProvider(gb.GetTracerProvider()),
	)

	return func(h http.Handler) http.Handler {
		return otelhttp.NewHandler(h, "otelhttp", opts...)
	}
}
