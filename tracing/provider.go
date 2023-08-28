package tracing

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/sdk/trace"
)

func NewProvider(endpoint string) (*trace.TracerProvider, func(), error) {
	ctx := context.Background()
	exp, err := newExporter(ctx, endpoint)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating trace exporter")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource()),
	)

	shutdown := func() {
		tp.Shutdown(ctx)
	}

	return tp, shutdown, nil
}
