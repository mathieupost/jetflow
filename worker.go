package jetflow

import (
	"context"
	"log"

	"github.com/pkg/errors"
)

type Worker struct {
	inbox   chan Request
	outbox  chan Response
	client  OperatorClient
	storage Storage
}

func NewWorker(client OperatorClient, storage Storage, inbox chan Request, outbox chan Response) *Worker {
	w := &Worker{
		inbox:   inbox,
		outbox:  outbox,
		client:  client,
		storage: storage,
	}

	return w
}

func (w *Worker) Start(ctx context.Context) {
	go w.processInbox(ctx)
}

func (w *Worker) processInbox(ctx context.Context) {
	loop := true
	for loop {
		select {
		case req := <-w.inbox:
			go w.processRequest(ctx, req)
		case <-ctx.Done():
			loop = false
		}
	}
	log.Println("Worker stopped")
}

func (w *Worker) processRequest(ctx context.Context, call Request) {
	log.Println(call)

	ctx = ContextWithOperationID(ctx, call.OperationID)

	operator := w.storage.Get(ctx, call)
	res, err := operator.Handle(ctx, w.client, call)
	if err != nil {
		log.Println("Worker handle call error:", err)
		w.outbox <- Response{
			RequestID: call.RequestID,
			Error:     errors.Wrap(err, "handle operator call"),
		}
		w.storage.Rollback(ctx, call)
		return
	}

	w.storage.Prepare(ctx, call)
	w.storage.Commit(ctx, call)

	w.outbox <- Response{
		RequestID: call.RequestID,
		Values:    res,
	}
}
