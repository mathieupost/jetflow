package jetstream

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow/example/gen"
	"github.com/mathieupost/jetflow/example/types"
)

func TestClientFind(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup jetstream
	js, clear := initJetStream(t, ctx)
	defer clear()
	client := initClient(t, ctx, js)

	// Find the users.
	var user1 types.User
	err := client.Find(ctx, "user1", &user1)
	require.NoError(t, err)

	var user2 types.User
	err = client.Find(ctx, "user2", &user2)
	require.NoError(t, err)

	// Call function on user1.
	err = user1.TransferBalance(ctx, user2, 1)
	require.NoError(t, err)
}

func initJetStream(t *testing.T, ctx context.Context) (js jetstream.JetStream, clear func()) {
	log.Println("initJetStream")

	nc, err := nats.Connect("0.0.0.0:4222")
	require.NoError(t, err)

	js, err = jetstream.New(nc)
	require.NoError(t, err)

	clear = func() {
		clearJetStream(t, ctx, js)
	}

	return js, clear
}

func clearJetStream(t *testing.T, ctx context.Context, js jetstream.JetStream) {
	log.Println("clearJetStream")

	// Clear all streams.
	streamNameList := js.StreamNames(ctx)
	for name := range streamNameList.Name() {
		log.Println("clearStream", name)

		stream, err := js.Stream(ctx, name)
		require.NoError(t, err)

		// Purge all messages.
		err = stream.Purge(ctx)
		require.NoError(t, err)

		// Delete all consumers.
		consumerNameList := stream.ConsumerNames(ctx)
		for consumer := range consumerNameList.Name() {
			log.Println("DeleteConsumer", consumer)
			err = stream.DeleteConsumer(ctx, consumer)
			require.NoError(t, err)
			log.Println("done DeleteConsumer", consumer)
		}

		// Delete the stream.
		err = js.DeleteStream(ctx, name)
		require.NoError(t, err)
		log.Println("done DeleteStream", name)
	}

	log.Println("done clearJetStream")
}

func initClient(t *testing.T, ctx context.Context, js jetstream.JetStream) *Client {
	log.Println("initClient")

	mapping := gen.FactoryMapping()
	client, err := NewClient(ctx, js, mapping)
	require.NoError(t, err)

	return client
}
