package transport

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types"
)

func IntegrationTest(t *testing.T, ctx context.Context, client jetflow.OperatorClient) {
	// Create a zipf distribution
	r := rand.New(rand.NewSource(87945723908))
	dist := rand.NewZipf(r, 1.5, 1, 100)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
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
