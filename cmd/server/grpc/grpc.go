package grpc

import "github.com/bnkamalesh/goapp/internal/api"

type GRPC struct {
	apis api.Server
}

func New(apis api.Server) *GRPC {
	return &GRPC{
		apis: apis,
	}
}
