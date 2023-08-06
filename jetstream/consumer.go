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
	storage   jetflow.Storage
}

func NewConsumer(ctx context.Context, jetstream jetstream.JetStream, client *Client, storage jetflow.Storage) (*Consumer, error) {
	consumerID := uuid.New()

	consumer := &Consumer{
		id:        consumerID,
		jetstream: jetstream,
		client:    client,
		storage:   storage,
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

	_, err = consumer.Consume(func(msg jetstream.Msg) {
		go c.handleOperatorCall(msg)
	})
	if err != nil {
		return errors.Wrap(err, "execute consumer")
	}

	return nil
}

func (c *Consumer) handleOperatorCall(msg jetstream.Msg) {
	headers := msg.Headers()
	clientID := headers.Get(HEADER_KEY_CLIENT_ID)
	requestID := headers.Get(HEADER_KEY_REQUEST_ID)
	callID := headers.Get(HEADER_KEY_CALL_ID)
	log.Println("Consumer.handleOperatorCall", callID)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ctx = ctxWithRequestID(ctx, requestID)

	// Unmarshal the method and parameters.
	var call jetflow.OperatorCall
	err := json.Unmarshal(msg.Data(), &call)
	if err != nil {
		log.Fatalln(err.Error(), msg.Data())
	}
	msg.Ack()

	operatorID := call.ID
	operatorType := call.Type

	var retries int
	var success bool
	var result jetflow.Result
	for !success {
		log.Println("Consumer.handleOperatorCall retries:", retries)
		retries++

		operator := c.storage.Get(ctx, operatorType, operatorID, requestID)
		result = operator.Call(ctx, c.client, call)

		// TODO ONLY if clientID was the original clientID, we should PREPARE and
		// COMMIT this and the other operators.
		success = c.storage.Prepare(ctx, operatorType, operatorID, requestID)
		if !success {
			continue
		}

		success = c.storage.Commit(ctx, operatorType, operatorID, requestID)
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Fatalln(err.Error(), result)
	}

	// Send back to the caller.
	res := nats.NewMsg("CLIENT." + clientID)
	res.Header.Set(HEADER_KEY_CALL_ID, callID)
	res.Data = data
	_, err = c.jetstream.PublishMsg(ctx, res)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("Consumer published msg", res.Subject)
}
