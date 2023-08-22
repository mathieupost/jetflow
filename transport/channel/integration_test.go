package channel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types/gen"
	"github.com/mathieupost/jetflow/log"
	"github.com/mathieupost/jetflow/storage/memory"
	"github.com/mathieupost/jetflow/tracing"
	"github.com/mathieupost/jetflow/transport"
)

func TestCall(t *testing.T) {
	tp, shutdown, err := tracing.NewProvider()
	if err != nil {
		log.Fatal("new tracing provider", err.Error())
	}
	defer shutdown()
	otel.SetTracerProvider(tp)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	requestChan := make(chan requestWithHeaders, 100)
	responseChan := make(chan *jetflow.Response, 100)

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := NewPublisher(requestChan, responseChan)
	client := jetflow.NewClient(factoryMapping, publisher)

	handlerFactory := gen.HandlerFactoryMapping()
	storage := memory.NewStorage(handlerFactory)

	executor := jetflow.NewExecutor(storage, client)
	consumer := NewConsumer(requestChan, responseChan, executor)
	consumer.Start(ctx)

	transport.IntegrationTest(t, ctx, client)
	tp.ForceFlush(ctx)
}
