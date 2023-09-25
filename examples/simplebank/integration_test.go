package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/log"
	"github.com/mathieupost/jetflow/storage/memory"
	"github.com/mathieupost/jetflow/tracing"
	"github.com/mathieupost/jetflow/transport/channel"
	"github.com/mathieupost/jetflow/transport/jetstream"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	natsjetstream "github.com/nats-io/nats.go/jetstream"
	"github.com/pingcap/go-ycsb/pkg/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow/examples/simplebank/types"
	"github.com/mathieupost/jetflow/examples/simplebank/types/gen"
)

func TestTransportChannel(t *testing.T) {
	tp, shutdown, err := tracing.NewProvider("localhost:4318", "testchannel")
	if err != nil {
		log.Fatal("new tracing provider", err.Error())
	}
	defer shutdown()
	otel.SetTracerProvider(tp)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	publisher, requestChan, responseChan := channel.NewPublisher()

	factoryMapping := gen.ProxyFactoryMapping()
	client := jetflow.NewClient(factoryMapping, publisher)

	handlerFactory := gen.HandlerFactoryMapping()
	storage := memory.NewStorage(handlerFactory)
	executor := jetflow.NewExecutor(storage, client)

	consumer := channel.NewConsumer(requestChan, responseChan, executor)
	consumer.Start(ctx)

	IntegrationTest(t, ctx, client)
	tp.ForceFlush(ctx)
}

func TestTransportJetStream(t *testing.T) {
	tp, shutdown, err := tracing.NewProvider("localhost:4318", "testjetstream")
	if err != nil {
		log.Fatal("new tracing provider", err.Error())
	}
	t.Cleanup(shutdown)
	otel.SetTracerProvider(tp)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	t.Cleanup(cancel)

	consumerAmount := 10
	js := initJetStream(t, ctx)

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := jetstream.NewPublisher(ctx, js, consumerAmount)
	client := jetflow.NewClient(factoryMapping, publisher)

	for i := 0; i < consumerAmount; i++ {
		factoryMapping := gen.ProxyFactoryMapping()
		publisher := jetstream.NewPublisher(ctx, js, consumerAmount)
		client := jetflow.NewClient(factoryMapping, publisher)

		handlerFactory := gen.HandlerFactoryMapping()
		storage := memory.NewStorage(handlerFactory)

		executor := jetflow.NewExecutor(storage, client)
		_ = jetstream.NewConsumer(ctx, i, js, executor)
	}

	IntegrationTest(t, ctx, client)
	tp.ForceFlush(ctx)
}

func initJetStream(t *testing.T, ctx context.Context) natsjetstream.JetStream {
	// Setup a NATS server with JetStream enabled.
	debug := false
	opts := server.Options{
		JetStream:    true,
		StoreDir:     t.TempDir(),
		Port:         server.RANDOM_PORT,
		Trace:        debug,
		TraceVerbose: debug,
		Debug:        debug,
		Logtime:      debug,
	}
	s, err := server.NewServer(&opts)
	require.NoError(t, err)
	s.ConfigureLogger()
	err = server.Run(s)
	require.NoError(t, err)

	nc, err := nats.Connect(fmt.Sprintf("0.0.0.0:%d", opts.Port))
	require.NoError(t, err)

	js, err := natsjetstream.New(nc)
	require.NoError(t, err)

	t.Cleanup(s.Shutdown)
	return js
}

func IntegrationTest(t *testing.T, ctx context.Context, client jetflow.OperatorClient) {
	// Create a zipf distribution
	r := rand.New(rand.NewSource(time.Now().Unix()))
	zipfGen := generator.NewZipfianWithItems(10000000, generator.ZipfianConstant)

	var wg sync.WaitGroup
	var times sync.Map
	userIDs := map[string]bool{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		i := i
		id1 := fmt.Sprintf("user%d", zipfGen.Next(r))
		id2 := fmt.Sprintf("user%d", zipfGen.Next(r))
		for id1 == id2 {
			id2 = fmt.Sprintf("user%d", zipfGen.Next(r))
		}
		userIDs[id1] = true
		userIDs[id2] = true

		go func() {
			defer wg.Done()
			defer func(t time.Time) {
				duration := time.Now().Sub(t)
				times.Store(i, duration)
				log.Println("took", duration)
			}(time.Now())

			var user1 types.User
			err := client.Find(ctx, id1, &user1)
			require.NoError(t, err)

			var user2 types.User
			err = client.Find(ctx, id2, &user2)
			require.NoError(t, err)

			err = fmt.Errorf("true")
			for err != nil {
				_, _, err = user1.TransferBalance(ctx, user2, 10)
				if err != nil {
					require.ErrorContains(t, err, "failed to prepare")
				}
			}

			err = fmt.Errorf("true")
			for err != nil {
				_, _, err = user2.TransferBalance(ctx, user1, 10)
				if err != nil {
					require.ErrorContains(t, err, "failed to prepare")
				}
			}
		}()
	}
	wg.Wait()
	count := 0
	total := time.Duration(0)
	times.Range(func(key, value any) bool {
		count++
		total = total + value.(time.Duration)
		return true
	})

	average := total / time.Duration(count)
	log.Println("Average:", average)

	for id := range userIDs {
		var user types.User
		err := client.Find(ctx, id, &user)
		require.NoError(t, err)
		balance, err := user.GetBalance(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1000000, balance, "for user %s", id)
	}
}
