package jetstream

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

// initialCallID
var initialCallID ctxKey = HEADER_KEY_INITIAL_CALL_ID

func ctxWithInitialCallID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, initialCallID, id)
}

func initialCallIDFromCtx(ctx context.Context, currentID string) string {
	initialCallID, ok := ctx.Value(initialCallID).(string)
	if !ok {
		return currentID
	}
	return initialCallID
}

// requestIDKey
var requestIDKey ctxKey = HEADER_KEY_REQUEST_ID

func ctxWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func requestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return uuid.NewString()
	}
	return requestID
}

// involvedOperatorsKey
var involvedOperatorsKey ctxKey = HEADER_KEY_INVOLVED_OPERATORS

// ctxAddInvolvedOperators adds the given operators to the existing involvedOperatorsKey value.
//
// Sort of a hack to pass the involved operators to the subscriber.
func ctxAddInvolvedOperators(ctx context.Context, operators ...string) context.Context {
	value, ok := ctx.Value(involvedOperatorsKey).(*[]string)
	if !ok {
		value = &[]string{}
		ctx = context.WithValue(ctx, involvedOperatorsKey, value)
	}

	newOperators := append(*value, operators...)
	*value = newOperators

	return ctx
}

func involvedOperatorsFromCtx(ctx context.Context) *[]string {
	operators, ok := ctx.Value(involvedOperatorsKey).(*[]string)
	if !ok || operators == nil {
		return &[]string{}
	}
	return operators
}
