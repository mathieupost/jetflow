package jetflow

import (
	"context"
)

// Operator is the minimal interface for all operators.
type Operator interface {
	ID() string
}

// OperatorHandler handles calls to the methods of an operator.
//
// It will convert an OperatorCall to a method call and convert the result to
// a Result struct.
type OperatorHandler interface {
	Handle(ctx context.Context, client OperatorClient, call Request) ([]byte, error)
}

//go:generate go run github.com/vektra/mockery/v2 --name OperatorClient --case underscore --with-expecter
type OperatorClient interface {
	Find(ctx context.Context, id string, operator interface{}) error
	Call(ctx context.Context, call Request) (res []byte, err error)
}

// ProxyFactoryMapping maps type names to OperatorFactories.
type (
	ProxyFactoryMapping   map[string]ProxyFactory
	HandlerFactoryMapping map[string]HandlerFactory
)

type (
	// ProxyFactory instantiates a OperatorProxy value.
	//
	// Used by the OperatorClient to instantiate an operator to an OperatorProxy, so method
	// calls can be proxied to the actual implementation.
	ProxyFactory   func(id string, client OperatorClient) Operator
	HandlerFactory func(id string) OperatorHandler
)

type Publisher interface {
	Publish(context.Context, Request) (chan Response, error)
}

type RequestHandler interface {
	Handle(ctx context.Context, req Request) Response
}

type Storage interface {
	Get(ctx context.Context, call Request) (OperatorHandler, error)
	Prepare(ctx context.Context, call Request) error
	Commit(ctx context.Context, call Request) error
	Rollback(ctx context.Context, call Request) error
}
