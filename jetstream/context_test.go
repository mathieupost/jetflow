package jetstream

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCtxInitialCallID(t *testing.T) {
	ctx := context.Background()

	expected := "expected"
	ctx = ctxWithInitialCallID(ctx, expected)

	actual := initialCallIDFromCtx(ctx, "wrong")
	require.Equal(t, expected, actual)
}

func TestCtxRequestIDKey(t *testing.T) {
	ctx := context.Background()

	expected := "expected"
	ctx = ctxWithRequestID(ctx, expected)

	actual := requestIDFromCtx(ctx)
	require.Equal(t, expected, actual)
}

func TestCtxInvolvedOperatorsKey(t *testing.T) {
	ctx := context.Background()

	expected1 := "first"
	ctx = ctxAddInvolvedOperators(ctx, expected1)

	actual := involvedOperatorsFromCtx(ctx)
	require.Len(t, *actual, 1)
	require.Equal(t, expected1, (*actual)[0])

	// Add some more.
	expected2 := "second"
	expected3 := "third"
	ctx = ctxAddInvolvedOperators(ctx, expected2, expected3)

	actual = involvedOperatorsFromCtx(ctx)
	require.Len(t, *actual, 3)
	require.Equal(t, expected1, (*actual)[0])
	require.Equal(t, expected2, (*actual)[1])
	require.Equal(t, expected3, (*actual)[2])
}
