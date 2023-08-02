package gen

import (
	"context"
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

func (u *UserProxy) TransferBalance(ctx context.Context, u2 types.User, amount int) error {
	log.Println("UserProxy.TransferBalance")

	resultChannel, err := u.Client.Send(ctx, u, jetflow.OperatorCall{
		Method: "TransferBalance",
	})
	if err != nil {
		return errors.Wrap(err, "send TransferBalance")
	}

	// Get result
	res := <-resultChannel
	return errors.Wrap(res.Error, "get result TransferBalance")
}

func (u *UserProxy) AddBalance(ctx context.Context, amount int) error {
	log.Println("UserProxy.AddBalance")

	resultChannel, err := u.Client.Send(ctx, u, jetflow.OperatorCall{
		Method: "AddBalance",
	})
	if err != nil {
		return errors.Wrap(err, "send AddBalance")
	}

	// Get result
	res := <-resultChannel
	return errors.Wrap(res.Error, "get result AddBalance")
}
