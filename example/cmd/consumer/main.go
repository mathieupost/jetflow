package main

import (
	"context"
	"log"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/storage/memory"
	"github.com/mathieupost/jetflow/tracing"
	"github.com/mathieupost/jetflow/transport/jetstream"
	"github.com/nats-io/nats.go"
	natsjetstream "github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow/example/types/gen"
)

func main() {
	tp, shutdown, err := tracing.NewProvider()
	if err != nil {
		log.Fatal("new tracing provider", err.Error())
	}
	defer shutdown()
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect("nats")
	if err != nil {
		log.Fatal("connecting to NATS", err.Error())
	}

	js, err := natsjetstream.New(nc)
	if err != nil {
		log.Fatal("initializing JetStream instance", err.Error())
	}

	log.Println("Consumer starting")

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := jetstream.NewPublisher(ctx, js)
	client := jetflow.NewClient(factoryMapping, publisher)

	handlerFactory := gen.HandlerFactoryMapping()
	storage := memory.NewStorage(handlerFactory)

	executor := jetflow.NewExecutor(storage, client)
	jetstream.NewConsumer(ctx, js, executor)

	log.Println("Consumer started")
	<-ctx.Done()
}
