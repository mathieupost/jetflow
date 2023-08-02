package gen

import (
	"context"
	"log"
	"reflect"

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
	return nil
}

func (u *UserProxy) AddBalance(ctx context.Context, amount int) error {
	log.Println("UserProxy.AddBalance")
	return nil
}
