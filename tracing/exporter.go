package tracing

import (
	"context"
	"io/ioutil"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func newExporter(ctx context.Context, endpoint string) (trace.SpanExporter, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	return otlptrace.New(ctx, client)
}

func newStdoutExporter(ctx context.Context, endpoint string) (trace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func newNoopExporter(ctx context.Context, endpoint string) (trace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithWriter(ioutil.Discard))
}
