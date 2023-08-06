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
