package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/storage/memory"
	"github.com/mathieupost/jetflow/tracing"
	"github.com/mathieupost/jetflow/transport/jetstream"
	"github.com/nats-io/nats.go"
	natsjetstream "github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow/examples/simplebank/types/gen"
)

func main() {
	id := os.Getenv("CONSUMER_ID")
	if id == "" {
		panic("no CONSUMER_ID specified")
	}

	tp, shutdown, err := tracing.NewProvider("jaeger:4318", fmt.Sprintf("consumer-%s", id))
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
	jetstream.NewConsumer(ctx, id, js, executor)

	log.Println("Consumer started")
	<-ctx.Done()
}
