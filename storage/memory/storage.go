package memory

import (
	"context"
	"reflect"
	"sync"

	"github.com/huandu/go-clone"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow"
)

var _ jetflow.Storage = (*Storage)(nil)

type Storage struct {
	typeHandlerMapping     jetflow.HandlerFactoryMapping
	keyVersionMapping      sync.Map
	versionOperatorMapping sync.Map
}

type version struct {
	base     string
	key      string
	prepared string
}

func NewStorage(mapping jetflow.HandlerFactoryMapping) *Storage {
	return &Storage{
		typeHandlerMapping: mapping,
	}
}

func (s *Storage) Get(ctx context.Context, call *jetflow.Request) (jetflow.OperatorHandler, error) {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Get")
	defer span.End()

	// Get last committed value. Create a new value if not found.
	operatorKey := call.TypeName + "." + call.InstanceID
	versionKey := operatorKey + "." + call.TransactionID

	// Load the operator version for the current request.
	committedVersion := s.keyVersionMappingLoadOrStore(operatorKey)
	operator, ok := s.versionOperatorMapping.Load(versionKey)
	if !ok {
		// Clone the committed version if there was no previous version for this request.
		operator, ok = s.versionOperatorMapping.Load(committedVersion.key)
		if !ok {
			// Create an initial instance of the operator if it did not yet exist.
			operator, _ = s.versionOperatorMapping.LoadOrStore(
				committedVersion.key,
				s.typeHandlerMapping[call.TypeName](call.InstanceID),
			)
		}
		// Store the operator version for the current request.
		operator = clone.Clone(operator)
		s.versionOperatorMapping.Store(versionKey, operator)
		s.keyVersionMapping.Store(versionKey, version{
			base: committedVersion.key,
			key:  versionKey,
		})
	} else {
		// Check if the version is not yet outdated because of a
		// new committed version.
		v, _ := s.keyVersionMapping.Load(versionKey)
		requestVersion := v.(version)
		if requestVersion.base != committedVersion.key {
			return nil, errors.New("outdated version")
		}
	}

	return operator.(jetflow.OperatorHandler), nil
}

func (s *Storage) Prepare(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Prepare")
	defer span.End()

	operatorKey := call.TypeName + "." + call.InstanceID
	committedVersion, err := s.keyVersionMappingLoad(operatorKey)
	if err != nil {
		return errors.Wrap(err, "loading committed version")
	}

	versionKey := operatorKey + "." + call.TransactionID
	version, err := s.keyVersionMappingLoad(versionKey)
	if err != nil {
		return errors.Wrap(err, "loading request version")
	}

	// Check if the version for the given transaction is created
	// from the committed version.
	if version.base != committedVersion.key {
		// A new version is already committed.
		return errors.New("base outdated")
	}

	operator, ok := s.versionOperatorMapping.Load(version.key)
	if !ok {
		return errors.Errorf("request operator does not exist for %s", versionKey)
	}
	baseOperator, ok := s.versionOperatorMapping.Load(version.base)
	if !ok {
		return errors.Errorf("base operator (%s) does not exist for %s", version.base, versionKey)
	}
	equal := reflect.DeepEqual(operator, baseOperator)
	if equal {
		// Operator was not written
		return nil
	}

	if committedVersion.prepared != "" {
		// Already prepared by another request.
		return errors.New("already prepared")
	}

	// Copy and set prepared to the version for the given request.
	newCommittedVersion := committedVersion
	newCommittedVersion.prepared = versionKey
	updated := s.keyVersionMappingSwap(operatorKey, committedVersion, newCommittedVersion)
	if !updated {
		return errors.New("failed to prepare")
	}

	return nil
}

func (s *Storage) Commit(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Commit")
	defer span.End()

	operatorKey := call.TypeName + "." + call.InstanceID
	versionKey := operatorKey + "." + call.TransactionID

	// Delete the transaction version mapping.
	defer s.keyVersionMapping.Delete(versionKey)

	committedVersion, err := s.keyVersionMappingLoad(operatorKey)
	if err != nil {
		return errors.Wrap(err, "loading old version")
	}

	if committedVersion.prepared != versionKey {
		return errors.New("not prepared by this request")
	}

	// Update the committed version to the prepared version.
	newVersion := version{key: committedVersion.prepared}
	updated := s.keyVersionMappingSwap(operatorKey, committedVersion, newVersion)
	if !updated {
		return errors.New("failed to commit")
	}

	// Delete the previous operator version.
	s.versionOperatorMapping.Delete(committedVersion.key)

	return nil
}

func (s *Storage) Rollback(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Rollback")
	defer span.End()

	operatorKey := call.TypeName + "." + call.InstanceID
	versionKey := operatorKey + "." + call.TransactionID

	// Cleanup version.
	s.keyVersionMapping.Delete(versionKey)
	s.versionOperatorMapping.Delete(versionKey)

	// Unprepare if needed.
	committedVersion, err := s.keyVersionMappingLoad(operatorKey)
	if err != nil {
		return errors.Wrap(err, "loading committed version")
	}
	if committedVersion.prepared == versionKey {
		newCommittedVersion := committedVersion
		newCommittedVersion.prepared = ""
		updated := s.keyVersionMappingSwap(operatorKey, committedVersion, newCommittedVersion)
		if !updated {
			return errors.New("failed to rollback")
		}
	}

	return nil
}

func (s *Storage) keyVersionMappingLoad(key string) (version, error) {
	v, _ := s.keyVersionMapping.Load(key)
	version, ok := v.(version)
	if !ok {
		return version, errors.New("version not found")
	}
	return version, nil
}

func (s *Storage) keyVersionMappingLoadOrStore(key string) version {
	v, _ := s.keyVersionMapping.LoadOrStore(key, version{key: key})
	return v.(version)
}

func (s *Storage) keyVersionMappingSwap(operatorKey string, old, version version) bool {
	return s.keyVersionMapping.CompareAndSwap(operatorKey, old, version)
}
