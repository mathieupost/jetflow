package jetstream

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/gen"
)

// Mock jetstream.JetStream
type mockJetStream struct {
	jetstream.JetStream
}

func (m *mockJetStream) CreateStream(ctx context.Context, cfg jetstream.StreamConfig) (jetstream.Stream, error) {
	return nil, nil
}

func (m *mockJetStream) CreateOrUpdateConsumer(ctx context.Context, stream string, cfg jetstream.ConsumerConfig) (jetstream.Consumer, error) {
	return &mockConsumer{}, nil
}

// Mock jetstream.Consumer
type mockConsumer struct {
	jetstream.Consumer
}

func (m *mockConsumer) Consume(handler jetstream.MessageHandler, opts ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error) {
	return nil, nil
}

type mockStorage struct{}

func (m *mockStorage) GetOperator(ctx context.Context, otype string, oid string, requestID string) (jetflow.OperatorHandler, error) {
	return gen.NewUserHandler(oid), nil
}

type mockOperatorHandler struct{}

func (m *mockOperatorHandler) Call(ctx context.Context, client jetflow.Client, msg jetflow.OperatorCall) jetflow.Result {
	return jetflow.Result{}
}
