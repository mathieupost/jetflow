package tracing

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/sdk/trace"
)

func NewProvider(endpoint, name string) (*trace.TracerProvider, func(), error) {
	ctx := context.Background()
	exp, err := newExporter(ctx, endpoint)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating trace exporter")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource(name)),
	)

	shutdown := func() {
		tp.Shutdown(ctx)
	}

	return tp, shutdown, nil
}
