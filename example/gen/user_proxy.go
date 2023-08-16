package gen

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types"
)

var (
	_ types.User       = (*UserProxy)(nil)
	_ jetflow.Operator = (*UserProxy)(nil)
)

type UserProxy struct {
	id     string
	client jetflow.OperatorClient
}

func NewUserProxy(id string, client jetflow.OperatorClient) jetflow.Operator {
	return &UserProxy{id: id, client: client}
}

func (u *UserProxy) ID() string {
	return u.id
}

type User_TransferBalance_Args struct {
	U2     *UserProxy
	Amount int
}

func (u *UserProxy) TransferBalance(ctx context.Context, u2 types.User, amount int) error {
	args := User_TransferBalance_Args{u2.(*UserProxy), amount}
	data, err := json.Marshal(args)
	if err != nil {
		return errors.Wrap(err, "marshalling TransferBalance args")
	}
	call := &jetflow.Request{
		Name:   "User",
		ID:     u.id,
		Method: "TransferBalance",
		Args:   data,
	}
	_, err = u.client.Call(ctx, call)
	return errors.Wrap(err, "call client UserProxy.TransferBalance")
}

type User_AddBalance_Args struct {
	Amount int
}

func (u *UserProxy) AddBalance(ctx context.Context, amount int) error {
	args := User_AddBalance_Args{amount}
	data, err := json.Marshal(args)
	if err != nil {
		return errors.Wrap(err, "marshalling TransferBalance args")
	}
	call := &jetflow.Request{
		Name:   "User",
		ID:     u.id,
		Method: "AddBalance",
		Args:   data,
	}
	_, err = u.client.Call(ctx, call)
	return errors.Wrap(err, "call client UserProxy.AddBalance")
}

// MarshalJSON implements json.Marshaler.
func (u UserProxy) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.id)
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *UserProxy) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &u.id)
}
