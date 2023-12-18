package jetflow_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/mocks"
)

func TestProcessRequest(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	type test struct {
		worker  jetflow.RequestHandler
		storage *mocks.Storage
		handler *mocks.OperatorHandler
		client  *mocks.OperatorClient
	}

	setup := func(t *testing.T) test {
		storage := mocks.NewStorage(t)
		handler := mocks.NewOperatorHandler(t)
		client := mocks.NewOperatorClient(t)

		worker := jetflow.NewExecutor(storage, client)

		return test{worker, storage, handler, client}
	}

	ANY := mock.Anything
	match := func(method jetflow.Method) interface{} {
		return mock.MatchedBy(func(m *jetflow.Request) bool {
			return m.Method == string(method)
		})
	}

	t.Run("Parent", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodCommit)).Return(nil, nil).Maybe() // FIXME: Async call

		request := &jetflow.Request{
			TransactionID: t.Name(),
			RequestID:     t.Name(),
		}
		response := tt.worker.Handle(ctx, request)

		require.Equal(t, t.Name(), response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})

	t.Run("Child", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()

		request := &jetflow.Request{
			TransactionID: t.Name(),
			RequestID:     "child",
		}
		response := tt.worker.Handle(ctx, request)

		require.Equal(t, "child", response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})

	t.Run("ParentHandleError", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		err := errors.New("handle error")
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, err).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodRollback)).Return(nil, nil).Maybe() // ROLLBACK
		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodCommit)).Return(nil, nil).Maybe() // FIXME: Async call

		request := &jetflow.Request{
			TransactionID: t.Name(),
			RequestID:     t.Name(),
		}
		response := tt.worker.Handle(ctx, request)

		require.Equal(t, t.Name(), response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})

	t.Run("ChildHandleError", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		err := errors.New("handle error")
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, err).Once()

		request := &jetflow.Request{
			TransactionID: t.Name(),
			RequestID:     "child",
		}
		response := tt.worker.Handle(ctx, request)

		require.Equal(t, "child", response.RequestID)
		require.ErrorIs(t, response.Error, err)
	})

	t.Run("ParentPrepareError", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		err := errors.New("handle error")
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, err).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodRollback)).Return(nil, nil).Maybe() // ROLLBACK
		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodCommit)).Return(nil, nil).Maybe() // FIXME: Async call

		request := &jetflow.Request{
			TransactionID: t.Name(),
			RequestID:     t.Name(),
		}
		response := tt.worker.Handle(ctx, request)

		require.Equal(t, t.Name(), response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})
}
