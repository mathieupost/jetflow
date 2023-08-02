package jetstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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
	id               uuid.UUID
	mapping          map[string]jetflow.OperatorFactory
	jetstream        jetstream.JetStream
	responseChannels sync.Map
}

func NewClient(ctx context.Context, jetstream jetstream.JetStream, mapping map[string]jetflow.OperatorFactory) (*Client, error) {
	clientID := uuid.New()

	client := &Client{
		id:        clientID,
		mapping:   mapping,
		jetstream: jetstream,
	}
	client.initStreams(ctx)

	return client, nil
}

func (c *Client) Send(ctx context.Context, operator jetflow.OperatorProxy, message jetflow.OperatorCall) (chan jetflow.Result, error) {
	log.Println("Client.Send", operator.Type(), operator.ID(), message.Method, string(message.Params))
	requestID := uuid.New().String()

	// Marshal the message
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, errors.Wrap(err, "marshal message")
	}

	// Create nats message
	subject := fmt.Sprintf("%s.%s.%s", STREAM_NAME_OPERATOR, operator.Type(), operator.ID())
	msg := nats.NewMsg(subject)
	msg.Header.Add(HEADER_KEY_CLIENT_ID, c.id.String())
	msg.Header.Add(HEADER_KEY_REQUEST_ID, requestID)
	msg.Data = payload

	// Publish the message to the OPERATOR stream.
	_, err = c.jetstream.PublishMsg(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "publish message")
	}

	// Create a channel to receive the response.
	responseChannel := make(chan jetflow.Result, 1)
	c.responseChannels.Store(requestID, responseChannel)
	responseChannel <- jetflow.Result{} // TODO remove when we receive responses.
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

func (c *Client) initStreams(ctx context.Context) error {
	_, err := c.jetstream.CreateStream(ctx, jetstream.StreamConfig{
		Name:     STREAM_NAME_OPERATOR,
		Subjects: []string{STREAM_NAME_OPERATOR + ".*.*"},
	})
	if err != nil {
		return errors.Wrap(err, "create operator stream")
	}

	return nil
}
