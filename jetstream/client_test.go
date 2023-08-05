package jetstream

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow/example/gen"
	"github.com/mathieupost/jetflow/example/types"
)

func TestClientFind(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	js := &mockJetStream{}
	client := initClient(t, ctx, js)

	var testUser types.User
	err := client.Find(ctx, "test_user", &testUser)
	require.NoError(t, err)
	require.NotNil(t, testUser)
	require.IsType(t, &gen.UserProxy{}, testUser)
}

func TestClientSend(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	t.Cleanup(cancel)

	// Setup jetstream
	js := initJetStream(t, ctx)
	client := initClient(t, ctx, js)

	// Setup consumer
	storage := &mockStorage{}
	_, err := NewConsumer(ctx, js, client, storage)
	require.NoError(t, err)

	// Find the users.
	var user1 types.User
	err = client.Find(ctx, "user1", &user1)
	require.NoError(t, err)

	var user2 types.User
	err = client.Find(ctx, "user2", &user2)
	require.NoError(t, err)

	// Call function on user1.
	err = user1.TransferBalance(ctx, user2, 1)
	require.NoError(t, err)
}

func initJetStream(t *testing.T, ctx context.Context) jetstream.JetStream {
	// Setup a NATS server with JetStream enabled.
	opts := server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      server.RANDOM_PORT,
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

func initClient(t *testing.T, ctx context.Context, js jetstream.JetStream) *Client {
	log.Println("initClient")

	mapping := gen.FactoryMapping()
	client, err := NewClient(ctx, js, mapping)
	require.NoError(t, err)

	return client
}
