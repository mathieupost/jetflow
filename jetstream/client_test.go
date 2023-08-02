package jetstream_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow/example/gen"
	"github.com/mathieupost/jetflow/example/types"
	"github.com/mathieupost/jetflow/jetstream"
)

func TestClientFind(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := jetstream.NewClient(ctx, gen.FactoryMapping())
	require.NoError(t, err)

	var testUser types.User
	err = client.Find(ctx, "test_user", &testUser)
	require.NoError(t, err)
	require.NotNil(t, testUser)
	require.IsType(t, &gen.UserProxy{}, testUser)
}
