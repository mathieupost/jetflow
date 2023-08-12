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

	// (mis)use the context to keep track of the involved operators.
	involvedOperators := map[string]map[string]bool{}
	ctx = ContextWithInvolvedOperators(ctx, involvedOperators)
	ContextAddInvolvedOperator(ctx, call.Name, call.ID)

	response := Response{
		RequestID:         call.RequestID,
		InvolvedOperators: involvedOperators,
	}
	defer func() {
		w.outbox <- response
	}()

	operator := w.storage.Get(ctx, call)
	res, err := operator.Handle(ctx, w.client, call)
	if err != nil {
		log.Println("Worker handle call error:", err)
		response.Error = errors.Wrap(err, "handle operator call")
		w.storage.Rollback(ctx, call)
		return
	}

	w.storage.Prepare(ctx, call)
	w.storage.Commit(ctx, call)

	response.Values = res
}
