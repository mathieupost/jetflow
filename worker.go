package jetflow

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
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
			go func() {
				switch req.Method {
				case string(MethodPrepare):
					err := w.storage.Prepare(ctx, req)
					w.outbox <- req.Response(ctx, nil, err)
				case string(MethodCommit):
					err := w.storage.Commit(ctx, req)
					w.outbox <- req.Response(ctx, nil, err)
				case string(MethodRollback):
					err := w.storage.Rollback(ctx, req)
					w.outbox <- req.Response(ctx, nil, err)
				default:
					w.processRequest(ctx, req)
				}
			}()
		case <-ctx.Done():
			loop = false
		}
	}
	log.Println("Worker stopped")
}

func (w *Worker) processRequest(ctx context.Context, call Request) {
	originalRequestID := call.RequestID

	retry := true
	for retry {
		log.Println("Worker.processRequest\n", call)

		ctx := ContextWithOperationID(ctx, call.OperationID)

		retry = false

		response := w.handle(ctx, call)
		log.Println("Worker.processRequest response:", response, "\n", call)

		// The initial request has the same id as the operation.
		isInitialRequest := call.OperationID == call.RequestID
		if isInitialRequest {
			operators := response.InvolvedOperators
			success := response.Error == nil

			// Try to prepare all involved operators.
			if success {
				prepared := w.broadcast(ctx, MethodPrepare, operators)
				success = prepared
			}

			// Rollback and retry if we either got an error or if we could not
			// prepare all involved operators.
			if !success {
				retry = true
				w.broadcast(ctx, MethodRollback, operators)

				// Create a new transaction id for the retry. Otherwise, the retry
				// may use the old state of the involved operators.
				transactionID := uuid.NewString()
				transactionID = transactionID[len(transactionID)-12:]
				log.Println(call.OperationID, "->", transactionID,
					"original:", originalRequestID)
				call.OperationID = transactionID
				call.RequestID = transactionID
				continue
			} else {
				w.broadcast(ctx, MethodCommit, operators)
			}
		}

		response.RequestID = originalRequestID
		w.outbox <- response
	}
}

func (w *Worker) broadcast(ctx context.Context, method Method, operators map[string]map[string]bool) bool {
	var success atomic.Bool
	success.Store(true)
	var wg sync.WaitGroup
	for name, instances := range operators {
		for id := range instances {
			request := Request{
				Name:   name,
				ID:     id,
				Method: string(method),
			}
			wg.Add(1)
			go func() {
				_, err := w.client.Call(ctx, request)
				if err != nil {
					log.Println("Worker.broadcast error:", err)
					success.Store(false)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	return success.Load()
}

func (w *Worker) handle(ctx context.Context, call Request) Response {
	// (mis)use the context to keep track of the involved operators.
	involvedOperators := map[string]map[string]bool{}
	ctx = ContextWithInvolvedOperators(ctx, involvedOperators)
	ContextAddInvolvedOperator(ctx, call.Name, call.ID)

	operator, err := w.storage.Get(ctx, call)
	if err != nil {
		err = errors.Wrap(err, "getting operator")
		return call.Response(ctx, nil, err)
	}

	res, err := operator.Handle(ctx, w.client, call)
	if err != nil {
		log.Println("Worker handle call error:", err)
		err = errors.Wrap(err, "handle operator call")
		return call.Response(ctx, nil, err)
	}

	return call.Response(ctx, res, nil)
}
