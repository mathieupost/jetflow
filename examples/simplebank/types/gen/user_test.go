package gen

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/mocks"
)

func TestTransferBalance(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewOperatorClient(t)
	client.EXPECT().
		Call(mock.Anything, mock.Anything).
		Return(nil, nil)

	proxy := NewUserProxy("user1", client)
	handler := NewUserHandler("user1")

	args := User_TransferBalance_Args{
		U2:     proxy.(*UserProxy),
		Amount: 10,
	}
	data, err := json.Marshal(args)
	call := &jetflow.Request{
		Method: "TransferBalance",
		Args:   data,
	}
	_, err = handler.Handle(ctx, client, call)
	require.NoError(t, err)
}
