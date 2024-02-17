package apm

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

func prometheusExporter() (*prometheus.Exporter, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, errors.Wrap(err, "promexporter.New")
	}

	return exporter, nil
}

func prometheusScraper(opts *Options) {
	mux := http.NewServeMux()
	mux.Handle("/-/metrics", promhttp.Handler())
	server := &http.Server{
		Handler:           mux,
		Addr:              fmt.Sprintf("%d", opts.PrometheusScrapePort),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// logger.Info(
	// 	"[otel/http] starting prometheus scrape endpoint",
	// 	zap.String(
	// 		"addr",
	// 		fmt.Sprintf("localhost:%d/-/metrics", opts.PrometheusScrapePort),
	// 	),
	// )
	err := server.ListenAndServe()
	if err != nil {
		// logger.Error(
		// 	"[otel/http] failed to serve metrics at:",
		// 	zap.Error(err),
		// 	zap.Uint16("port", opts.PrometheusScrapePort),
		// )
		return
	}
}
