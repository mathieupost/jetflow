package channel

import (
	"context"

	"go.opentelemetry.io/otel/propagation"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/log"
)

type Consumer struct {
	inbox   chan requestWithHeaders
	outbox  chan *jetflow.Response
	handler jetflow.RequestHandler
}

func NewConsumer(inbox chan requestWithHeaders, outbox chan *jetflow.Response, handler jetflow.RequestHandler) *Consumer {
	s := &Consumer{
		inbox:   inbox,
		outbox:  outbox,
		handler: handler,
	}
	return s
}

func (w *Consumer) Start(ctx context.Context) {
	go w.processInbox(ctx)
}

func (w *Consumer) processInbox(ctx context.Context) {
	loop := true
	for loop {
		select {
		case req := <-w.inbox:
			go func() {
				// Extract the trace context from the message header.
				propagator := propagation.TraceContext{}
				carrier := propagation.HeaderCarrier(req.headers)
				ctx := propagator.Extract(ctx, carrier)

				w.outbox <- w.handler.Handle(ctx, req.Request)
			}()
		case <-ctx.Done():
			loop = false
		}
	}
	log.Println("Consumer stopped")
}
