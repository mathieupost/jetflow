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
		inbox   chan jetflow.Request
		outbox  chan jetflow.Response
		storage *mocks.Storage
		handler *mocks.Handler
		client  *mocks.OperatorClient
	}

	setup := func(t *testing.T) test {
		inbox := make(chan jetflow.Request, 1)
		outbox := make(chan jetflow.Response, 1)
		storage := mocks.NewStorage(t)
		handler := mocks.NewHandler(t)
		client := mocks.NewOperatorClient(t)

		worker := jetflow.NewWorker(storage, client, inbox, outbox)
		worker.Start(ctx)

		return test{inbox, outbox, storage, handler, client}
	}

	ANY := mock.Anything
	match := func(method jetflow.Method) interface{} {
		return mock.MatchedBy(func(m jetflow.Request) bool {
			return m.Method == string(method)
		})
	}

	t.Run("Parent", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodCommit)).Return(nil, nil).Once()

		tt.inbox <- jetflow.Request{
			OperationID: t.Name(),
			RequestID:   t.Name(),
		}
		response := <-tt.outbox

		require.Equal(t, t.Name(), response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})

	t.Run("Child", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()

		tt.inbox <- jetflow.Request{
			OperationID: t.Name(),
			RequestID:   "child",
		}
		response := <-tt.outbox

		require.Equal(t, "child", response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})

	t.Run("ParentHandleError", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		err := errors.New("handle error")
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, err).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodRollback)).Return(nil, nil).Once() // ROLLBACK
		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodCommit)).Return(nil, nil).Once()

		tt.inbox <- jetflow.Request{
			OperationID: t.Name(),
			RequestID:   t.Name(),
		}
		response := <-tt.outbox

		require.Equal(t, t.Name(), response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})

	t.Run("ChildHandleError", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		err := errors.New("handle error")
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, err).Once()

		tt.inbox <- jetflow.Request{
			OperationID: t.Name(),
			RequestID:   "child",
		}
		response := <-tt.outbox

		require.Equal(t, "child", response.RequestID)
		require.ErrorIs(t, response.Error, err)
	})

	t.Run("ParentPrepareError", func(t *testing.T) {
		tt := setup(t)

		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		err := errors.New("handle error")
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, err).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodRollback)).Return(nil, nil).Once() // ROLLBACK
		tt.storage.EXPECT().Get(ANY, ANY).Return(tt.handler, nil).Once()
		tt.handler.EXPECT().Handle(ANY, ANY, ANY).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodPrepare)).Return(nil, nil).Once()
		tt.client.EXPECT().Call(ANY, match(jetflow.MethodCommit)).Return(nil, nil).Once()

		tt.inbox <- jetflow.Request{
			OperationID: t.Name(),
			RequestID:   t.Name(),
		}
		response := <-tt.outbox

		require.Equal(t, t.Name(), response.RequestID)
		require.ErrorIs(t, response.Error, nil)
	})
}
