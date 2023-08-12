package jetflow

import (
	"context"
	"log"
)

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

// involvedOperatorsKey
var involvedOperatorsKey ctxKey = "INVOLVED_OPERATORS"

func InvolvedOperatorsFromContext(ctx context.Context) map[string]map[string]bool {
	operators, ok := ctx.Value(involvedOperatorsKey).(map[string]map[string]bool)
	if !ok {
		return map[string]map[string]bool{}
	}
	return operators
}

func ContextWithInvolvedOperators(ctx context.Context, operators map[string]map[string]bool) context.Context {
	return context.WithValue(ctx, involvedOperatorsKey, operators)
}

// ContextAddInvolvedOperator mutates context's set of involved operators.
func ContextAddInvolvedOperator(ctx context.Context, name, id string) context.Context {
	operators := InvolvedOperatorsFromContext(ctx)
	instances, ok := operators[name]
	if !ok {
		instances = map[string]bool{}
		operators[name] = instances
	}
	instances[id] = true
	log.Println(operators)
	return ContextWithInvolvedOperators(ctx, operators)
}
