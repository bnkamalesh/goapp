// Package kafka implements the Kafka subscription functionality
package kafka

import "github.com/bnkamalesh/goapp/internal/api"

type Kafka struct {
	apis api.Subscriber
}

func New(apis api.Subscriber) *Kafka {
	return &Kafka{
		apis: apis,
	}
}
