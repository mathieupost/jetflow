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

	handleErr := errors.New("error")
	table := []struct {
		name        string
		operationID string
		requestID   string
		handleErr   error
		setup       func(storage *mocks.Storage)
	}{
		{
			name:        "MainRequest",
			operationID: "operation1",
			requestID:   "operation1",
			handleErr:   nil,
			setup: func(storage *mocks.Storage) {
				storage.EXPECT().Prepare(mock.Anything, mock.Anything).Return(true)
				storage.EXPECT().Commit(mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:        "MainRequestFailed",
			operationID: "operation1",
			requestID:   "operation1",
			handleErr:   handleErr,
			setup: func(storage *mocks.Storage) {
				storage.EXPECT().Rollback(mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:        "SubRequest",
			operationID: "operation1",
			requestID:   "request2",
			handleErr:   nil,
			setup: func(storage *mocks.Storage) {
				storage.EXPECT().Prepare(mock.Anything, mock.Anything).Return(true)
				storage.EXPECT().Commit(mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:        "SubRequest",
			operationID: "operation1",
			requestID:   "request2",
			handleErr:   handleErr,
			setup: func(storage *mocks.Storage) {
				storage.EXPECT().Rollback(mock.Anything, mock.Anything).Return()
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			inbox := make(chan jetflow.Request, 1)
			outbox := make(chan jetflow.Response, 1)
			client := mocks.NewOperatorClient(t)
			storage := mocks.NewStorage(t)
			handler := mocks.NewHandler(t)

			worker := jetflow.NewWorker(client, storage, inbox, outbox)
			worker.Start(ctx)

			storage.EXPECT().Get(mock.Anything, mock.Anything).Return(handler)
			handler.EXPECT().Handle(mock.Anything, mock.Anything, mock.Anything).Return(nil, tt.handleErr)
			tt.setup(storage)

			inbox <- jetflow.Request{
				OperationID: tt.operationID,
				RequestID:   tt.requestID,
			}

			response := <-outbox
			require.Equal(t, tt.requestID, response.RequestID)
			require.ErrorIs(t, response.Error, tt.handleErr)
		})
	}
}
