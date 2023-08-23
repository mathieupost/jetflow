package memory

import (
	"context"
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
	operatorKey := call.Name + "." + call.ID
	requestOperatorKey := operatorKey + "." + call.OperationID

	// Load the operator version for the current request.
	committedVersion := s.loadOrCreateVersion(operatorKey)
	operator, ok := s.versionOperatorMapping.Load(requestOperatorKey)
	if !ok {
		// Clone the committed version if there was no previous version for this request.
		operator, ok = s.versionOperatorMapping.Load(committedVersion.key)
		if !ok {
			// Create an initial instance of the operator if it did not yet exist.
			operator, _ = s.versionOperatorMapping.LoadOrStore(
				committedVersion.key,
				s.typeHandlerMapping[call.Name](call.ID),
			)
		}
		// Store the operator version for the current request.
		operator = clone.Clone(operator)
		s.versionOperatorMapping.Store(requestOperatorKey, operator)
		s.keyVersionMapping.Store(requestOperatorKey, version{
			base: committedVersion.key,
			key:  requestOperatorKey,
		})
	} else {
		v, _ := s.keyVersionMapping.Load(requestOperatorKey)
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

	operatorKey := call.Name + "." + call.ID
	committedVersion, err := s.loadVersion(operatorKey)
	if err != nil {
		return errors.Wrap(err, "loading committed version")
	}

	if committedVersion.prepared != "" {
		// Already prepared by another request.
		return errors.New("already prepared")
	}

	requestOperatorKey := operatorKey + "." + call.OperationID
	requestVersion, err := s.loadVersion(requestOperatorKey)
	if err != nil {
		return errors.Wrap(err, "loading request version")
	}

	// Check if the version for the given request is created
	// from the committed version.
	if requestVersion.base != committedVersion.key {
		// A new version is already committed.
		return errors.New("base outdated")
	}

	// Copy and set prepared to the version for the given request.
	newCommittedVersion := committedVersion
	newCommittedVersion.prepared = requestOperatorKey
	updated := s.swapVersion(operatorKey, committedVersion, newCommittedVersion)
	if !updated {
		return errors.New("failed to prepare")
	}

	return nil
}

func (s *Storage) Commit(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Commit")
	defer span.End()

	operatorKey := call.Name + "." + call.ID
	requestOperatorKey := operatorKey + "." + call.OperationID
	oldVersion, err := s.loadVersion(operatorKey)
	if err != nil {
		return errors.Wrap(err, "loading old version")
	}

	if oldVersion.prepared != requestOperatorKey {
		return errors.New("not prepared by this request")
	}

	// Update the committed version to the prepared version.
	newVersion := version{key: oldVersion.prepared}
	updated := s.swapVersion(operatorKey, oldVersion, newVersion)
	if !updated {
		return errors.New("failed to commit")
	}

	// Delete the request version mapping.
	s.keyVersionMapping.Delete(requestOperatorKey)
	// Delete the previous operator version.
	s.versionOperatorMapping.Delete(oldVersion.key)

	return nil
}

func (s *Storage) Rollback(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Rollback")
	defer span.End()

	operatorKey := call.Name + "." + call.ID
	requestOperatorKey := operatorKey + "." + call.OperationID

	// Cleanup version.
	s.keyVersionMapping.Delete(requestOperatorKey)
	s.versionOperatorMapping.Delete(requestOperatorKey)

	// Unprepare if needed.
	committedVersion, err := s.loadVersion(operatorKey)
	if err != nil {
		return errors.Wrap(err, "loading committed version")
	}
	if committedVersion.prepared == requestOperatorKey {
		newCommittedVersion := committedVersion
		newCommittedVersion.prepared = ""
		updated := s.swapVersion(operatorKey, committedVersion, newCommittedVersion)
		if !updated {
			return errors.New("failed to rollback")
		}
	}

	return nil
}

func (s *Storage) loadVersion(key string) (version, error) {
	v, _ := s.keyVersionMapping.Load(key)
	version, ok := v.(version)
	if !ok {
		return version, errors.New("version not found")
	}
	return version, nil
}

func (s *Storage) loadOrCreateVersion(key string) version {
	v, _ := s.keyVersionMapping.LoadOrStore(key, version{key: key})
	return v.(version)
}

func (s *Storage) swapVersion(operatorKey string, old, version version) bool {
	return s.keyVersionMapping.CompareAndSwap(operatorKey, old, version)
}
