package channel

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/gen"
	"github.com/mathieupost/jetflow/example/types"
	"github.com/mathieupost/jetflow/storage/memory"
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

	// Create a zipf distribution
	r := rand.New(rand.NewSource(87945723908))
	dist := rand.NewZipf(r, 1.5, 1, 100)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			id1 := fmt.Sprintf("user%d", dist.Uint64())
			id2 := fmt.Sprintf("user%d", dist.Uint64())

			var user1 types.User
			err := client.Find(ctx, id1, &user1)
			require.NoError(t, err)

			var user2 types.User
			err = client.Find(ctx, id2, &user2)
			require.NoError(t, err)

			err = user1.TransferBalance(ctx, user2, 10)
			require.NoError(t, err)

			err = user2.TransferBalance(ctx, user1, 10)
			require.NoError(t, err)
		}()
	}
	wg.Wait()
}
