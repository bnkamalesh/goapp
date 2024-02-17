package apm

import (
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func OtelGrpcStreamInterceptor(ignoredMethods ...string) grpc.StreamServerInterceptor {
	checkList := map[string]struct{}{}
	for _, m := range ignoredMethods {
		checkList[m] = struct{}{}
	}
	return otelgrpc.StreamServerInterceptor(
		otelgrpc.WithInterceptorFilter(func(info *otelgrpc.InterceptorInfo) bool {
			_, ignored := checkList[info.Method]
			return !ignored
		}),
		otelgrpc.WithTracerProvider(Global().GetTracerProvider()),
		otelgrpc.WithMeterProvider(Global().GetMeterProvider()),
	)
}

func OtelGrpcUnaryInterceptor(ignoredMethods ...string) grpc.UnaryServerInterceptor {
	checkList := map[string]struct{}{}
	for _, m := range ignoredMethods {
		checkList[m] = struct{}{}
	}
	return otelgrpc.UnaryServerInterceptor(
		otelgrpc.WithInterceptorFilter(func(info *otelgrpc.InterceptorInfo) bool {
			_, ignored := checkList[info.Method]
			return !ignored
		}),
		otelgrpc.WithTracerProvider(Global().GetTracerProvider()),
		otelgrpc.WithMeterProvider(Global().GetMeterProvider()),
	)
}

func OtelGrpcClientStreamInterceptor(ignoredMethods ...string) grpc.StreamClientInterceptor {
	checkList := map[string]struct{}{}
	for _, m := range ignoredMethods {
		checkList[m] = struct{}{}
	}

	return otelgrpc.StreamClientInterceptor(
		otelgrpc.WithInterceptorFilter(func(info *otelgrpc.InterceptorInfo) bool {
			_, ignored := checkList[info.Method]
			return !ignored
		}),
		otelgrpc.WithTracerProvider(Global().GetTracerProvider()),
		otelgrpc.WithMeterProvider(Global().GetMeterProvider()),
	)
}

func OtelGrpcClientUnaryInterceptor(ignoredMethods ...string) grpc.UnaryClientInterceptor {
	checkList := map[string]struct{}{}
	for _, m := range ignoredMethods {
		checkList[m] = struct{}{}
	}

	return otelgrpc.UnaryClientInterceptor(
		otelgrpc.WithInterceptorFilter(func(info *otelgrpc.InterceptorInfo) bool {
			_, ignored := checkList[info.Method]
			return !ignored
		}),
		otelgrpc.WithTracerProvider(Global().GetTracerProvider()),
		otelgrpc.WithMeterProvider(Global().GetMeterProvider()),
	)
}

func NewGrpcClient(address string, port int) (*grpc.ClientConn, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(OtelGrpcClientUnaryInterceptor()),
		grpc.WithStreamInterceptor(OtelGrpcClientStreamInterceptor()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", address, port), dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gRPC client: %w", err)
	}
	return conn, err
}
