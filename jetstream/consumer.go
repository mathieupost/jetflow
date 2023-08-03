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
	consumer, err := c.jetstream.CreateOrUpdateConsumer(
		ctx,
		STREAM_NAME_OPERATOR,
		jetstream.ConsumerConfig{
			Durable:   c.id.String(),
			AckPolicy: jetstream.AckExplicitPolicy,
		},
	)
	if err != nil {
		return errors.Wrap(err, "create consumer")
	}

	// Setup a channel to receive messages from the JetStream server.
	c.inbox = make(chan jetstream.Msg, 100)
	_, err = consumer.Consume(func(msg jetstream.Msg) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subject := msg.Subject()
	headers := msg.Headers()
	log.Println("Consumer received on", subject, headers)

	// Unmarshal the method and parameters.
	var message jetflow.OperatorCall
	err := json.Unmarshal(msg.Data(), &message)
	if err != nil {
		panic(err)
	}
	msg.Ack()

	requestID := headers.Get(HEADER_KEY_REQUEST_ID)
	// TODO retrieve/init the operator based on the subject
	// type and id and the requestID.
	// TODO call actual operator
	// TODO send back result

	data, err := json.Marshal(jetflow.Result{})
	if err != nil {
		panic(err)
	}

	// Send back to the caller.
	clientID := headers.Get(HEADER_KEY_CLIENT_ID)
	res := nats.NewMsg("CLIENT." + clientID)
	res.Header.Set(HEADER_KEY_REQUEST_ID, requestID)
	res.Data = data
	_, err = c.jetstream.PublishMsg(ctx, res)
	if err != nil {
		panic(err)
	}
	log.Println("Consumer published msg", res.Subject)
}
