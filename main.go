package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bnkamalesh/errors"

	"github.com/bnkamalesh/goapp/cmd/server/grpc"
	"github.com/bnkamalesh/goapp/cmd/server/http"
	"github.com/bnkamalesh/goapp/internal/api"
	"github.com/bnkamalesh/goapp/internal/configs"
	"github.com/bnkamalesh/goapp/internal/pkg/apm"
	"github.com/bnkamalesh/goapp/internal/pkg/logger"
	"github.com/bnkamalesh/goapp/internal/pkg/postgres"
	"github.com/bnkamalesh/goapp/internal/pkg/sysignals"
	"github.com/bnkamalesh/goapp/internal/users"
)

// recoverer is used for panic recovery of the application (note: this is not for the HTTP/gRPC servers).
// So that even if the main function panics we can produce required logs for troubleshooting
var exitErr error

func recoverer(ctx context.Context) {
	exitCode := 0
	var exitInfo any
	rec := recover()
	err, _ := rec.(error)
	if err != nil {
		exitCode = 1
		exitInfo = err
	} else if rec != nil {
		exitCode = 2
		exitInfo = rec
	} else if exitErr != nil {
		exitCode = 3
		exitInfo = exitErr
	}

	// exiting after receiving a quit signal can be considered a *clean/successful* exit
	if errors.Is(exitErr, sysignals.ErrSigQuit) {
		exitCode = 0
	}

	// logging this because we have info logs saying "listening to" various port numbers
	// based on the server type (gRPC, HTTP etc.). But it's unclear *from the logs*
	// if the server is up and running, if it exits for any reason
	if exitCode == 0 {
		logger.Info(ctx, fmt.Sprintf("shutdown complete: %+v", exitInfo))
	} else {
		logger.Error(ctx, fmt.Sprintf("shutdown complete (exit: %d): %+v", exitCode, exitInfo))
	}

	os.Exit(exitCode)
}

func shutdown(
	httpServer *http.HTTP,
	grpcServer *grpc.GRPC,
	apmIns *apm.APM,
) {
	const shutdownTimeout = time.Second * 60
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info(ctx, "initiating shutdown")

	wgroup := &sync.WaitGroup{}
	wgroup.Add(1)

	wgroup.Wait()

	// after all the APIs of the application are shutdown (e.g. HTTP, gRPC, Pubsub listener etc.)
	// we should close connections to dependencies like database, cache etc.
	// This should only be done after the APIs are shutdown completely
	wgroup.Add(1)
	go func() {
		defer wgroup.Done()
	}()

	wgroup.Wait()
}

func startAPM(ctx context.Context, cfg *configs.Configs) *apm.APM {
	ap, err := apm.New(ctx, &apm.Options{
		Debug:            cfg.Environment == configs.EnvLocal,
		Environment:      cfg.Environment.String(),
		ServiceName:      cfg.AppName,
		ServiceVersion:   cfg.AppVersion,
		TracesSampleRate: 50.00,
		UseStdOut:        cfg.Environment == configs.EnvLocal,
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to start APM"))
	}
	return ap
}

func startServers(svr api.Server, cfgs *configs.Configs) (*http.HTTP, *grpc.GRPC) {
	hcfg, _ := cfgs.HTTP()
	hserver, err := http.NewService(hcfg, svr)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize HTTP server"))
	}

	err = hserver.Start()
	if err != nil {
		panic(errors.Wrap(err, "failed to start HTTP server"))
	}

	return hserver, nil
}

func main() {
	ctx := context.Background()
	defer recoverer(ctx)
	fatalErr := make(chan error, 1)

	cfgs, err := configs.New()
	if err != nil {
		panic(errors.Wrap(err))
	}

	logger.UpdateDefaultLogger(logger.New(
		cfgs.AppName, cfgs.AppVersion, 0,
		map[string]string{
			"env": cfgs.Environment.String(),
		}),
	)

	apmhandler := startAPM(ctx, cfgs)
	pqdriver, err := postgres.NewPool(cfgs.Postgres())
	if err != nil {
		panic(errors.Wrap(err))
	}

	userPGstore := users.NewPostgresStore(pqdriver, cfgs.UserPostgresTable())
	userSvc := users.NewService(userPGstore)
	svrAPIs := api.NewServer(userSvc, nil)
	hserver, gserver := startServers(svrAPIs, cfgs)

	defer shutdown(
		hserver,
		gserver,
		apmhandler,
	)

	exitErr = <-fatalErr
}
