package jetstream

import (
	"context"
	"fmt"
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types/gen"
	"github.com/mathieupost/jetflow/storage/memory"
	"github.com/mathieupost/jetflow/transport"
)

func TestCall(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	js := initJetStream(t, ctx)

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := NewPublisher(ctx, js)
	client := jetflow.NewClient(factoryMapping, publisher)

	handlerFactory := gen.HandlerFactoryMapping()
	storage := memory.NewStorage(handlerFactory)

	executor := jetflow.NewExecutor(storage, client)
	_ = NewConsumer(ctx, js, executor)

	transport.IntegrationTest(t, ctx, client)
}

func initJetStream(t *testing.T, ctx context.Context) jetstream.JetStream {
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

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	t.Cleanup(s.Shutdown)
	return js
}
