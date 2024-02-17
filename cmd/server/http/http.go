package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bnkamalesh/goapp/internal/api"
	"github.com/bnkamalesh/goapp/internal/pkg/apm"
	"github.com/bnkamalesh/webgo/v7"
	"github.com/bnkamalesh/webgo/v7/middleware/accesslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Config holds all the configuration required to start the HTTP server
type Config struct {
	Host string
	Port uint16

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration

	TemplatesBasePath string
	EnableAccessLog   bool
}

type HTTP struct {
	listener string
	server   *webgo.Router
}

// Start starts the HTTP server
func (h *HTTP) Start() error {
	h.server.Start()
	return nil
}

// NewService returns an instance of HTTP with all its dependencies set
func NewService(cfg *Config, apis api.Server) (*HTTP, error) {
	home, err := loadHomeTemplate(cfg.TemplatesBasePath)
	if err != nil {
		return nil, err
	}

	handlers := &Handlers{
		apis: apis,
		home: home,
	}

	router := webgo.NewRouter(
		&webgo.Config{
			Host:            cfg.Host,
			Port:            strconv.Itoa(int(cfg.Port)),
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			ShutdownTimeout: cfg.WriteTimeout * 2,
		},
		handlers.routes()...,
	)

	if cfg.EnableAccessLog {
		router.Use(accesslog.AccessLog)
		router.UseOnSpecialHandlers(accesslog.AccessLog)
	}
	router.Use(panicRecoverer)

	otelopts := []otelhttp.Option{
		// in this app, /-/ prefixed routes are used for healthchecks, readiness checks etc.
		otelhttp.WithFilter(func(req *http.Request) bool {
			return !strings.HasPrefix(req.URL.Path, "/-/")
		}),
		// the span name formatter is used to reduce the cardinality of metrics generated
		// when using URIs with variables in it
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			wctx := webgo.Context(r)
			if wctx == nil {
				return r.URL.Path
			}
			return wctx.Route.Pattern
		}),
	}

	apmMw := apm.NewHTTPMiddleware(otelopts...)
	router.Use(func(w http.ResponseWriter, r *http.Request, hf http.HandlerFunc) {
		apmMw(hf).ServeHTTP(w, r)
	})

	return &HTTP{
		server:   router,
		listener: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}, nil
}
