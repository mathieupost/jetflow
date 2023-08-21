package jetflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types"
	"github.com/mathieupost/jetflow/example/types/gen"
)

func TestClientFind(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mapping := gen.ProxyFactoryMapping()
	var client jetflow.OperatorClient = jetflow.NewClient(mapping, nil)
	var testUser types.User
	err := client.Find(ctx, "test_user", &testUser)
	require.NoError(t, err)
	require.NoError(t, ctx.Err())
	require.NotNil(t, testUser)
	require.IsType(t, &gen.UserProxy{}, testUser)
}
