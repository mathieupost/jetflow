package jetflow

import (
	"context"
	"reflect"
)

type Operator interface {
	ID() string
}

type Client interface {
	Find(ctx context.Context, id string, operator interface{}) error
	Send(context.Context, Operator, OperatorCall) (chan Result, error)
}

type OperatorFactory func(id string, client Client) reflect.Value

type OperatorCall struct {
	Method string
	Params []byte
}

type Result struct {
	Error  error
	Values []byte
}
