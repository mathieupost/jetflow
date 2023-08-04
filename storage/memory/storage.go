package memory

import (
	"context"
	"sync"

	"github.com/mathieupost/jetflow"
)

type Storage struct {
	mapping   map[string]func() jetflow.OperatorHandler
	operators sync.Map
}

func (s *Storage) GetOperator(ctx context.Context, otype string, oid string, requestID string) (jetflow.OperatorHandler, error) {
	key := otype + "." + oid
	defaultValue := s.mapping[otype]()
	v, _ := s.operators.LoadOrStore(key, defaultValue)
	value := v.(jetflow.OperatorHandler)
	return value, nil
}
