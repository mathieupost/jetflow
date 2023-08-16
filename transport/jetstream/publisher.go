package jetstream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/log"
)

var _ jetflow.Publisher = (*Publisher)(nil)

type Publisher struct {
	jetstream        jetstream.JetStream
	id               string
	responseChannels sync.Map
}

func NewPublisher(ctx context.Context, jetstream jetstream.JetStream) *Publisher {
	id := uuid.NewString()
	id = id[len(id)-12:]

	d := &Publisher{
		id:               id,
		jetstream:        jetstream,
		responseChannels: sync.Map{},
	}

	d.initStreams(ctx)
	d.initConsumer(ctx)

	return d
}

func (d *Publisher) Publish(ctx context.Context, call jetflow.Request) (chan jetflow.Response, error) {
	// Setup the channel to which the response will be sent.
	responseChan := make(chan jetflow.Response, 10)
	d.responseChannels.Store(call.RequestID, responseChan)

	// Marshal the message
	payload, err := json.Marshal(call)
	if err != nil {
		return nil, errors.Wrap(err, "marshal message")
	}

	// Create nats message
	subject := fmt.Sprintf("%s.%s.%s", STREAM_NAME_OPERATOR, call.Name, call.ID)
	msg := nats.NewMsg(subject)
	msg.Header.Set("ClientID", d.id)
	msg.Data = payload

	// Publish the message to the OPERATOR stream.
	_, err = d.jetstream.PublishMsg(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(err, "publish message")
	}

	return responseChan, nil
}

func (d *Publisher) initStreams(ctx context.Context) error {
	_, err := d.jetstream.CreateStream(ctx, jetstream.StreamConfig{
		Name:      STREAM_NAME_CLIENT,
		Subjects:  []string{STREAM_NAME_CLIENT + ".*"},
		Retention: jetstream.WorkQueuePolicy,
	})
	if err != nil {
		return errors.Wrap(err, "create client stream")
	}

	_, err = d.jetstream.CreateStream(ctx, jetstream.StreamConfig{
		Name:      STREAM_NAME_OPERATOR,
		Subjects:  []string{STREAM_NAME_OPERATOR + ".*.*"},
		Retention: jetstream.WorkQueuePolicy,
	})
	if err != nil {
		return errors.Wrap(err, "create operator stream")
	}

	return nil
}

func (d *Publisher) initConsumer(ctx context.Context) error {
	log.Println("Client.initConsumer")
	consumer, err := d.jetstream.CreateOrUpdateConsumer(
		ctx,
		STREAM_NAME_CLIENT,
		jetstream.ConsumerConfig{
			Durable:       "Client-" + d.id,
			AckPolicy:     jetstream.AckExplicitPolicy,
			FilterSubject: fmt.Sprintf("%s.%s", STREAM_NAME_CLIENT, d.id),
		},
	)
	if err != nil {
		return errors.Wrap(err, "create consumer")
	}

	_, err = consumer.Consume(func(msg jetstream.Msg) {
		// Unmarshal the response
		var response jetflow.Response
		err := json.Unmarshal(msg.Data(), &response)
		if err != nil {
			log.Fatalln("unmarshal result", err, string(msg.Data()))
		}
		// Acknowledge the result
		err = msg.Ack()
		if err != nil {
			log.Fatalln("acknowledge result", err, string(msg.Data()))
		}

		d.handleResponse(response)
	})
	if err != nil {
		return errors.Wrap(err, "init message consumer")
	}

	return nil
}

func (d *Publisher) handleResponse(response jetflow.Response) {
	c, ok := d.responseChannels.LoadAndDelete(response.RequestID)
	if !ok {
		log.Fatalln("response not found", response)
	}

	responseChan := c.(chan jetflow.Response)
	responseChan <- response
}
