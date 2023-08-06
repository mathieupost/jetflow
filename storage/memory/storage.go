package memory

import (
	"context"
	"log"
	"sync"

	"github.com/huandu/go-clone"

	"github.com/mathieupost/jetflow"
)

type Storage struct {
	mapping              map[string]func(id string) jetflow.OperatorHandler
	operatorStore        sync.Map
	requestOperatorStore sync.Map
}

func NewStorage(mapping map[string]func(id string) jetflow.OperatorHandler) *Storage {
	return &Storage{
		mapping: mapping,
	}
}

func (s *Storage) Get(ctx context.Context, otype string, oid string, requestID string) (jetflow.OperatorHandler, error) {
	defaultValue := s.mapping[otype](oid)

	// Get last committed value. Create a new value if not found.
	operatorKey := otype + "." + oid
	operator, exists := s.operatorStore.LoadOrStore(operatorKey, defaultValue)

	// Get value for this request. Use last committed value if not found.
	requestOperatorKey := operatorKey + "." + requestID
	operator = clone.Clone(operator)
	requestOperator, loaded := s.requestOperatorStore.LoadOrStore(requestOperatorKey, operator)

	log.Printf("GetOperator(%s) created: %t, new version: %t, pointer: %p\n", requestOperatorKey, !exists, !loaded, requestOperator)
	value := requestOperator.(jetflow.OperatorHandler)
	return value, nil
}
