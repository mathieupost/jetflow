package jetstream

import (
	"context"
	"log"
	"reflect"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
)

const (
	STREAM_NAME_CLIENT   = "CLIENT"
	STREAM_NAME_OPERATOR = "OPERATOR"

	HEADER_KEY_CLIENT_ID  = "ClientID"
	HEADER_KEY_REQUEST_ID = "RequestID"
)

var _ jetflow.Client = (*Client)(nil)

type Client struct {
	id      uuid.UUID
	mapping map[string]jetflow.OperatorFactory
}

func NewClient(ctx context.Context, mapping map[string]jetflow.OperatorFactory) (*Client, error) {
	clientID := uuid.New()

	client := &Client{
		id:      clientID,
		mapping: mapping,
	}

	return client, nil
}

func (c *Client) Send(ctx context.Context, operator jetflow.Operator, message jetflow.OperatorCall) (chan jetflow.Result, error) {
	log.Println("Client.Send", operator.ID(), message.Method, string(message.Params))

	responseChannel := make(chan jetflow.Result, 1)
	responseChannel <- jetflow.Result{}
	return responseChannel, nil
}

func (c *Client) Find(ctx context.Context, id string, operator interface{}) error {
	log.Printf("Client.Find %T %s\n", operator, id)
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
	value.Set(factory(id, c))

	return nil
}
