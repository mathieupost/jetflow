package jetflow

import (
	"context"
	"reflect"
)

// Operator is the minimal interface for all operators.
type Operator interface {
	ID() string
}

// OperatorProxy is a proxy which forwards calls to an operator.
//
// It should also implement the methods of a Operator implementation. These
// methods will then convert the method call to an OperatorCall to be send by
// a Client.
type OperatorProxy interface {
	Operator
	Type() string
}

// OperatorHandler handles calls to the methods of an operator.
//
// It will convert an OperatorCall to a method call and convert the result to
// a Result struct.
type OperatorHandler interface {
	Call(ctx context.Context, msg OperatorCall) Result
}

// Client initializes OperatorProxies and publishes their OperatorCalls.
//
// It creates OperatorProxies for each operator and sends their OperatorCalls
// to the right OperatorHandlers, which may also use OperatorProxies for
// operator methods that receive operators in their parameters.
type Client interface {
	Find(ctx context.Context, id string, operator interface{}) error
	Send(context.Context, Operator, OperatorCall) (chan Result, error)
}

// OperatorFactoryMapping maps type names to OperatorFactories.
type OperatorFactoryMapping map[string]OperatorFactory

// OperatorFactory instantiates a ProxyOperator value.
//
// Used by the Client to instantiate an operator to a ProxyOperator, so method
// calls can be proxied to the actual implementation.
type OperatorFactory func(id string, client Client) reflect.Value

type OperatorCall struct {
	Type   string
	Method string
	Params []byte
}

type Result struct {
	Error  error
	Values []byte
}
