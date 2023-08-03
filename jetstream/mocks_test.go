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
