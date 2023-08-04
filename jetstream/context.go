package jetstream

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

var requestIDKey ctxKey

func ctxWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func requestKeyFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(*string)
	if !ok || requestID == nil {
		return uuid.NewString()
	}
	return *requestID
}
