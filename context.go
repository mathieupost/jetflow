package jetflow

import "context"

type ctxKey string

// operationIDKey
var operationIDKey ctxKey = "OPERATION_ID"

func OperationIDFromContext(ctx context.Context, callID string) string {
	operationID, ok := ctx.Value(operationIDKey).(string)
	if !ok {
		return callID
	}
	return operationID
}

func ContextWithOperationID(ctx context.Context, operationID string) context.Context {
	return context.WithValue(ctx, operationIDKey, operationID)
}
