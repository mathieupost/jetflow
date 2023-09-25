package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/storage/memory"
	"github.com/mathieupost/jetflow/tracing"
	"github.com/mathieupost/jetflow/transport/jetstream"
	"github.com/nats-io/nats.go"
	natsjetstream "github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow/examples/simplebank/types/gen"
)

func main() {
	consumerID, err := strconv.Atoi(os.Getenv("CONSUMER_ID"))
	if err != nil {
		panic(errors.Wrap(err, "parsing CONSUMER_ID"))
	}
	consumersAmount, err := strconv.Atoi(os.Getenv("CONSUMERS"))
	if err != nil {
		panic(errors.Wrap(err, "parsing CONSUMERS"))
	}
	if consumerID == consumersAmount {
		consumerID = 0
	}
	natsHost := os.Getenv("NATS_HOST")
	if natsHost == "" {
		natsHost = "nats"
	}
	jaegerHost := os.Getenv("JAEGER_HOST")
	if jaegerHost == "" {
		jaegerHost = "jaeger"
	}

	tp, shutdown, err := tracing.NewProvider(jaegerHost+":4318", fmt.Sprintf("consumer-%d", consumerID))
	if err != nil {
		log.Fatal("new tracing provider", err.Error())
	}
	defer shutdown()
	otel.SetTracerProvider(tp)
	defer func() {
		tp.ForceFlush(context.Background())
		println("flushed")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect(natsHost)
	if err != nil {
		log.Fatal("connecting to NATS", err.Error())
	}

	js, err := natsjetstream.New(nc)
	if err != nil {
		log.Fatal("initializing JetStream instance", err.Error())
	}

	log.Println("Consumer starting")

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := jetstream.NewPublisher(ctx, js, consumersAmount)
	client := jetflow.NewClient(factoryMapping, publisher)

	handlerFactory := gen.HandlerFactoryMapping()
	storage := memory.NewStorage(handlerFactory)

	executor := jetflow.NewExecutor(storage, client)
	jetstream.NewConsumer(ctx, consumerID, js, executor)

	log.Println("Consumer started")
	<-ctx.Done()
}
