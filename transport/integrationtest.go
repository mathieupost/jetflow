package transport

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types"
	"github.com/mathieupost/jetflow/log"
)

func IntegrationTest(t *testing.T, ctx context.Context, client jetflow.OperatorClient) {
	// Create a zipf distribution
	r := rand.New(rand.NewSource(87945723908))
	dist := rand.NewZipf(r, 1.5, 1, 100)

	var wg sync.WaitGroup
	var times sync.Map
	for i := 0; i < 100; i++ {
		wg.Add(1)
		i := i
		id1 := fmt.Sprintf("user%d", dist.Uint64())
		id2 := fmt.Sprintf("user%d", dist.Uint64())

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

			err = user1.TransferBalance(ctx, user2, 10)
			require.NoError(t, err)

			err = user2.TransferBalance(ctx, user1, 10)
			require.NoError(t, err)
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
}
