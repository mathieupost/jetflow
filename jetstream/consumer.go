package jetstream

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
)

type Consumer struct {
	id        uuid.UUID
	jetstream jetstream.JetStream
	client    *Client
	inbox     chan jetstream.Msg
}

func NewConsumer(ctx context.Context, jetstream jetstream.JetStream, client *Client) (*Consumer, error) {
	consumerID := uuid.New()

	consumer := &Consumer{
		id:        consumerID,
		jetstream: jetstream,
		client:    client,
	}
	consumer.initConsumer(ctx)

	return consumer, nil
}

func (c *Consumer) initConsumer(ctx context.Context) error {
	log.Println("Consumer.initConsumer")
	consumer, err := c.jetstream.CreateOrUpdateConsumer(
		ctx,
		STREAM_NAME_OPERATOR,
		jetstream.ConsumerConfig{
			Durable:   "Consumer-" + c.id.String(),
			AckPolicy: jetstream.AckExplicitPolicy,
		},
	)
	if err != nil {
		return errors.Wrap(err, "create consumer")
	}
	log.Println("Consumer.initConsumer", consumer.CachedInfo().Name)

	// Setup a channel to receive messages from the JetStream server.
	c.inbox = make(chan jetstream.Msg, 100)
	_, err = consumer.Consume(func(msg jetstream.Msg) {
		log.Println("Consumer.consume put in inbox")
		c.inbox <- msg
	})
	if err != nil {
		return errors.Wrap(err, "init message consumer")
	}

	// Handle messages from the JetStream server.
	go func() {
		for msg := range c.inbox {
			c.handleOperatorCall(msg)
		}
	}()

	return nil
}

func (c *Consumer) handleOperatorCall(msg jetstream.Msg) {
	log.Println("Consumer.handleOperatorCall")
	headers := msg.Headers()
	clientID := headers.Get(HEADER_KEY_CLIENT_ID)
	requestID := headers.Get(HEADER_KEY_REQUEST_ID)
	callID := headers.Get(HEADER_KEY_CALL_ID)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ctx = ctxWithRequestID(ctx, requestID)

	// Unmarshal the method and parameters.
	var message jetflow.OperatorCall
	err := json.Unmarshal(msg.Data(), &message)
	if err != nil {
		panic(err)
	}
	msg.Ack()

	// TODO retrieve/init the operator.
	// TODO call actual operator
	// TODO fill the result

	data, err := json.Marshal(jetflow.Result{})
	if err != nil {
		panic(err)
	}

	// Send back to the caller.
	res := nats.NewMsg("CLIENT." + clientID)
	res.Header.Set(HEADER_KEY_CALL_ID, callID)
	res.Data = data
	_, err = c.jetstream.PublishMsg(ctx, res)
	if err != nil {
		panic(err)
	}
	log.Println("Consumer published msg", res.Subject)
}
