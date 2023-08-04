package gen

import (
	"context"
	"encoding/json"
	"log"
	"reflect"

	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types"
)

var _ types.User = (*UserProxy)(nil)

type UserProxy struct {
	OperatorID string
	Client     jetflow.Client
}

func NewUserProxy(id string, client jetflow.Client) reflect.Value {
	proxy := &UserProxy{OperatorID: id, Client: client}
	return reflect.ValueOf(proxy)
}

func (u *UserProxy) ID() string {
	return u.OperatorID
}

type Params_User_TransferBalance struct {
	U2     *UserProxy
	Amount int
}

func (u *UserProxy) TransferBalance(ctx context.Context, u2 types.User, amount int) error {
	log.Println("UserProxy.TransferBalance")
	params := Params_User_TransferBalance{U2: u2.(*UserProxy), Amount: amount}
	bytes, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "marshal params")
	}

	resultChannel, err := u.Client.Send(ctx, u, jetflow.OperatorCall{
		Type:   "User",
		Method: "TransferBalance",
		Params: bytes,
	})
	if err != nil {
		return errors.Wrap(err, "send TransferBalance")
	}

	// Get result
	res := <-resultChannel
	return errors.Wrap(res.Error, "get result TransferBalance")
}

type Params_User_AddBalance struct {
	Amount int
}

func (u *UserProxy) AddBalance(ctx context.Context, amount int) error {
	log.Println("UserProxy.AddBalance")
	params := Params_User_AddBalance{Amount: amount}
	bytes, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "marshal params")
	}

	resultChannel, err := u.Client.Send(ctx, u, jetflow.OperatorCall{
		Type:   "User",
		Method: "AddBalance",
		Params: bytes,
	})
	if err != nil {
		return errors.Wrap(err, "send AddBalance")
	}

	// Get result
	res := <-resultChannel
	return errors.Wrap(res.Error, "get result AddBalance")
}
