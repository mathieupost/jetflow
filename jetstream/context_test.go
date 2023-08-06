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
