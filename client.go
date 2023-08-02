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
}

type OperatorFactory func(id string, client Client) reflect.Value
