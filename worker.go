package jetflow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow/log"
)

type Executor struct {
	client  OperatorClient
	storage Storage
}

func NewExecutor(storage Storage, client OperatorClient) *Executor {
	w := &Executor{
		client:  client,
		storage: storage,
	}

	return w
}

func (w *Executor) Handle(ctx context.Context, req *Request) *Response {
	switch req.Method {
	case string(MethodPrepare):
		err := w.storage.Prepare(ctx, req)
		return req.Response(ctx, nil, err)
	case string(MethodCommit):
		err := w.storage.Commit(ctx, req)
		return req.Response(ctx, nil, err)
	case string(MethodRollback):
		err := w.storage.Rollback(ctx, req)
		return req.Response(ctx, nil, err)
	default:
		return w.handleCall(ctx, req)
	}
}

func (w *Executor) handleCall(ctx context.Context, call *Request) *Response {
	originalRequestID := call.RequestID

	retryCount := 0
	for {
		res, success := w.try(ctx, call)

		if !success {
			// Create a new transaction id for the retry. Otherwise, the retry
			// may use the old state of the involved operators.
			transactionID := fmt.Sprintf("%s-%d", originalRequestID, retryCount)
			transactionID = transactionID[len(transactionID)-12:]
			log.Println(call.OperationID, "->", transactionID,
				"original:", originalRequestID)
			call.OperationID = transactionID
			call.RequestID = transactionID

			retryCount++
			continue // Retry

		}

		res.RequestID = originalRequestID
		return res
	}
}

func (w *Executor) try(ctx context.Context, call *Request) (*Response, bool) {
	log.Println("Executor.processRequest\n", call)

	ctx = ContextWithOperationID(ctx, call.OperationID)

	response := w.handle(ctx, call)
	log.Println("Executor.processRequest response:", response, "\n", call)

	// The initial request has the same id as the operation.
	isInitialRequest := call.OperationID == call.RequestID
	if isInitialRequest {
		ctx, span := otel.Tracer("").Start(ctx, "jetflow.Executor.2pc")
		defer span.End()

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
			w.broadcast(ctx, MethodRollback, operators)

			return nil, false
		} else {
			w.broadcast(ctx, MethodCommit, operators)
		}
	}

	return response, true
}

func (w *Executor) broadcast(ctx context.Context, method Method, operators map[string]map[string]bool) bool {
	ctx, span := otel.Tracer("").Start(ctx, "jetflow.Executor.broadcast."+string(method))
	defer span.End()

	var success atomic.Bool
	success.Store(true)
	var wg sync.WaitGroup
	for name, instances := range operators {
		for id := range instances {
			request := &Request{
				Name:   name,
				ID:     id,
				Method: string(method),
			}
			wg.Add(1)
			go func() {
				_, err := w.client.Call(ctx, request)
				if err != nil {
					log.Println("Executor.broadcast error:", err)
					success.Store(false)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	return success.Load()
}

func (w *Executor) handle(ctx context.Context, call *Request) *Response {
	// (mis)use the context to keep track of the involved operators.
	involvedOperators := map[string]map[string]bool{}
	ctx = ContextWithInvolvedOperators(ctx, involvedOperators)
	ContextAddInvolvedOperator(ctx, call.Name, call.ID)

	operator, err := w.storage.Get(ctx, call)
	if err != nil {
		err = errors.Wrap(err, "getting operator")
		return call.Response(ctx, nil, err)
	}

	_, span := otel.Tracer("").Start(ctx, "operator.Handle")
	operatorspan := &span
	ctx = context.WithValue(ctx, "SPAN", operatorspan)
	defer func() {
		(*operatorspan).End()
	}()

	res, err := operator.Handle(ctx, w.client, call)
	if err != nil {
		log.Println("Executor handle call error:", err)
		err = errors.Wrap(err, "handle operator call")
		return call.Response(ctx, nil, err)
	}

	return call.Response(ctx, res, nil)
}
