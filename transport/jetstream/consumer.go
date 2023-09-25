package jetstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/log"
)

type Consumer struct {
	id        int
	jetstream jetstream.JetStream
	handler   jetflow.RequestHandler
}

func NewConsumer(
	ctx context.Context,
	id int,
	jetstream jetstream.JetStream,
	handler jetflow.RequestHandler,
) *Consumer {
	consumer := &Consumer{
		id:        id,
		jetstream: jetstream,
		handler:   handler,
	}

	consumer.initConsumer(ctx)

	return consumer
}

func (r *Consumer) initConsumer(ctx context.Context) error {
	log.Println("Consumer.initConsumer")
	consumer, err := r.jetstream.CreateOrUpdateConsumer(
		ctx,
		STREAM_NAME_OPERATOR,
		jetstream.ConsumerConfig{
			Durable:       fmt.Sprintf("Consumer-%d", r.id),
			AckPolicy:     jetstream.AckExplicitPolicy,
			FilterSubject: fmt.Sprintf("%s.*.*.%d", STREAM_NAME_OPERATOR, r.id),
		},
	)
	if err != nil {
		return errors.Wrap(err, "create consumer")
	}
	log.Println("Consumer.initConsumer", consumer.CachedInfo().Name)

	consCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		go r.handle(ctx, msg)
	})
	if err != nil {
		return errors.Wrap(err, "execute consumer")
	}

	go func() {
		<-ctx.Done()
		consCtx.Stop()
	}()

	return nil
}

func (r *Consumer) handle(ctx context.Context, msg jetstream.Msg) {
	// Extract the trace context from the message header.
	propagator := propagation.TraceContext{}
	carrier := propagation.HeaderCarrier(msg.Headers())
	ctx = propagator.Extract(ctx, carrier)

	ctx, span := otel.Tracer("").Start(ctx, "jetstream.Consumer.handle")
	defer span.End()

	clientID := msg.Headers().Get("ClientID")

	// Unmarshal the method and parameters.
	call := &jetflow.Request{}
	err := json.Unmarshal(msg.Data(), call)
	if err != nil {
		log.Fatalln("unmarshal request", err, string(msg.Data()))
	}

	// Acknowledge the result
	err = msg.Ack()
	if err != nil {
		log.Fatalln("acknowledge request", err, string(msg.Data()))
	}

	// Handle the request.
	response := r.handler.Handle(ctx, call)

	// Marshal the response.
	data, err := json.Marshal(response)
	if err != nil {
		log.Fatalln(response.RequestID, err.Error(), response)
	}

	ctx, pubspan := otel.Tracer("").Start(ctx, "jetstream.Consumer.publish")
	defer pubspan.End()
	// Send back to the caller.
	subject := "CLIENT." + clientID
	res := nats.NewMsg(subject)
	res.Data = data
	_, err = r.jetstream.PublishMsg(ctx, res)
	if err != nil {
		log.Fatalln(response.RequestID, "Consumer.handle publish to", subject, err.Error())
	}
}
