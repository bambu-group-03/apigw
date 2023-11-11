package amqprpc

import (
	"github.com/bambu-group-03/apigw/internal/usecase"
	"github.com/bambu-group-03/apigw/pkg/rabbitmq/rmq_rpc/server"
)

// NewRouter -.
func NewRouter(t usecase.Translation) map[string]server.CallHandler {
	routes := make(map[string]server.CallHandler)
	{
		newTranslationRoutes(routes, t)
	}

	return routes
}
