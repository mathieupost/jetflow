package jetflow

import (
	"context"
	"log"
	"reflect"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var _ OperatorClient = (*Client)(nil)

type Client struct {
	mapping    ProxyFactoryMapping
	dispatcher Publisher
}

func NewClient(mapping ProxyFactoryMapping, dispatcher Publisher) *Client {
	return &Client{
		mapping:    mapping,
		dispatcher: dispatcher,
	}
}

func (c *Client) Call(ctx context.Context, call Request) (res []byte, err error) {
	requestID := uuid.NewString()
	requestID = requestID[len(requestID)-12:]
	call.OperationID = OperationIDFromContext(ctx, requestID)
	call.RequestID = requestID

	log.Println("Client.Call:\n", call)

	replyChan, err := c.dispatcher.Publish(ctx, call)
	if err != nil {
		return nil, errors.Wrap(err, "dispatching request")
	}

	select {
	case reply := <-replyChan:
		// Mutate the ctx's involved operators.
		for name, instances := range reply.InvolvedOperators {
			for id := range instances {
				ContextAddInvolvedOperator(ctx, name, id)
			}
		}
		return reply.Values, errors.Wrap(reply.Error, "dispatch call")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "dispatch call")
	}
}

func (c *Client) Find(ctx context.Context, id string, operator interface{}) error {
	value := reflect.ValueOf(operator)
	if value.Kind() != reflect.Pointer {
		return errors.New("operator must be a pointer")
	}

	value = value.Elem()
	if value.Kind() != reflect.Interface {
		return errors.New("operator must be an interface")
	}

	if value.Elem().Kind() != reflect.Invalid {
		return errors.New("operator already initalized")
	}

	name := value.Type().Name()
	factory, ok := c.mapping[name]
	if !ok {
		return errors.Errorf("operator '%s' not found", name)
	}
	operator = factory(id, c)
	value.Set(reflect.ValueOf(operator))

	return nil
}
