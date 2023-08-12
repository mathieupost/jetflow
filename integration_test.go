package jetflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/gen"
	"github.com/mathieupost/jetflow/example/types"
	"github.com/mathieupost/jetflow/storage/memory"
)

func TestCall(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	t.Cleanup(cancel)

	requestChan := make(chan jetflow.Request, 10)
	responseChan := make(chan jetflow.Response, 1)

	dispatcher := memory.NewDispatcher(requestChan, responseChan)

	factoryMapping := gen.FactoryMapping()
	client := jetflow.NewClient(factoryMapping, dispatcher)

	handlerFactory := gen.HandlerFactoryMapping()
	storage := memory.NewStorage(handlerFactory)
	worker := jetflow.NewWorker(client, storage, requestChan, responseChan)
	worker.Start(ctx)

	var user1 types.User
	err := client.Find(ctx, "user1", &user1)
	require.NoError(t, err)

	var user2 types.User
	err = client.Find(ctx, "user2", &user2)
	require.NoError(t, err)

	err = user1.TransferBalance(ctx, user2, 10)
	require.NoError(t, err)
}
