package jetstream

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
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
