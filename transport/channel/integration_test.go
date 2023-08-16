package channel

import (
	"context"
	"testing"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/gen"
	"github.com/mathieupost/jetflow/storage/memory"
	"github.com/mathieupost/jetflow/transport"
)

func TestCall(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	requestChan := make(chan jetflow.Request, 100)
	responseChan := make(chan jetflow.Response, 100)

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := NewPublisher(requestChan, responseChan)
	client := jetflow.NewClient(factoryMapping, publisher)

	handlerFactory := gen.HandlerFactoryMapping()
	storage := memory.NewStorage(handlerFactory)

	executor := jetflow.NewExecutor(storage, client)
	consumer := NewConsumer(requestChan, responseChan, executor)
	consumer.Start(ctx)

	transport.IntegrationTest(t, ctx, client)
}
