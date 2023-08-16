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
	Handle(context.Context, OperatorClient, *Request) ([]byte, error)
}

//go:generate go run github.com/vektra/mockery/v2 --name OperatorClient --case underscore --with-expecter
type OperatorClient interface {
	Find(ctx context.Context, id string, operator interface{}) error
	Call(context.Context, *Request) (res []byte, err error)
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
	Publish(context.Context, *Request) (chan *Response, error)
}

type RequestHandler interface {
	Handle(context.Context, *Request) *Response
}

type Storage interface {
	Get(context.Context, *Request) (OperatorHandler, error)
	Prepare(context.Context, *Request) error
	Commit(context.Context, *Request) error
	Rollback(context.Context, *Request) error
}
