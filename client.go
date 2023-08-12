package jetflow

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var _ OperatorClient = (*Client)(nil)

type Client struct {
	id         string
	mapping    ProxyFactoryMapping
	dispatcher Dispatcher
}

func NewClient(mapping ProxyFactoryMapping, dispatcher Dispatcher) *Client {
	id := uuid.NewString()
	id = id[len(id)-12:]
	return &Client{
		id:         id,
		mapping:    mapping,
		dispatcher: dispatcher,
	}
}

func (c *Client) Call(ctx context.Context, call Request) (res []byte, err error) {
	callID := uuid.NewString()
	callID = callID[len(callID)-12:]
	call.ClientID = c.id
	call.OperationID = OperationIDFromContext(ctx, callID)
	call.RequestID = callID

	replyChan, err := c.dispatcher.Dispatch(ctx, call)
	if err != nil {
		return nil, errors.Wrap(err, "dispatching request")
	}

	select {
	case reply := <-replyChan:
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
