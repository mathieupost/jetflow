package channel

import (
	"context"
	"log"

	"github.com/mathieupost/jetflow"
)

type Consumer struct {
	inbox   chan jetflow.Request
	outbox  chan jetflow.Response
	handler jetflow.RequestHandler
}

func NewConsumer(inbox chan jetflow.Request, outbox chan jetflow.Response, handler jetflow.RequestHandler) *Consumer {
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
				w.outbox <- w.handler.Handle(ctx, req)
			}()
		case <-ctx.Done():
			loop = false
		}
	}
	log.Println("Consumer stopped")
}
