package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/huandu/go-clone"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/log"
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

	log.Println("Storage.Get\n", call)
	return s.get(ctx, call.Name, call.ID, call.OperationID), nil
}

func (s *Storage) get(ctx context.Context, name string, id string, req string) jetflow.OperatorHandler {
	// Get last committed value. Create a new value if not found.
	operatorKey := name + "." + id
	requestOperatorKey := operatorKey + "." + req

	// Load the operator version for the current request.
	operator, ok := s.versionOperatorMapping.Load(requestOperatorKey)
	if !ok {
		log.Println("Storage.Get clone", requestOperatorKey)
		// Clone the committed version if there was no previous version for this request.
		committedVersion := s.loadOrCreateVersion(operatorKey)
		operator, ok = s.versionOperatorMapping.Load(committedVersion.key)
		if !ok {
			log.Println("Storage.Get create", requestOperatorKey)
			// Create an initial instance of the operator if it did not yet exist.
			operator, _ = s.versionOperatorMapping.LoadOrStore(
				committedVersion.key,
				s.typeHandlerMapping[name](id),
			)
		}
		// Store the operator version for the current request.
		operator = clone.Clone(operator)
		s.versionOperatorMapping.Store(requestOperatorKey, operator)
		s.keyVersionMapping.Store(requestOperatorKey, version{
			base: committedVersion.key,
			key:  requestOperatorKey,
		})
	}

	return operator.(jetflow.OperatorHandler)
}

func (s *Storage) Prepare(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Prepare")
	defer span.End()

	prepared := s.prepare(ctx, call.Name, call.ID, call.OperationID)
	if !prepared {
		return errors.New("failed to prepare")
	}
	return nil
}

func (s *Storage) prepare(ctx context.Context, name string, id string, req string) (ok bool) {
	log.Println(req, "Storage.Prepare", name, id, req)

	operatorKey := name + "." + id
	committedVersion := s.loadVersion(operatorKey)

	if committedVersion.prepared != "" {
		// Already prepared by another request.
		log.Println(req, "Storage.Prepare already prepared", committedVersion.prepared)
		return false
	}

	requestOperatorKey := operatorKey + "." + req
	requestVersion := s.loadVersion(requestOperatorKey)

	// Check if the version for the given request is created
	// from the committed version.
	if requestVersion.base != committedVersion.key {
		// A new version is already committed.
		log.Println(req, "Storage.Prepare already new version", committedVersion.key)
		return false
	}

	// Copy and set prepared to the version for the given request.
	newCommittedVersion := committedVersion
	newCommittedVersion.prepared = requestOperatorKey
	updated := s.updateVersion(operatorKey, committedVersion, newCommittedVersion)
	log.Println(req, "Storage.Prepare old", committedVersion, "new", newCommittedVersion, "updated", updated)
	return updated
}

func (s *Storage) Commit(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Commit")
	defer span.End()

	s.commit(ctx, call.Name, call.ID, call.OperationID)
	return nil
}

func (s *Storage) commit(ctx context.Context, name string, id string, req string) {
	log.Println(req, "Storage.Commit", name, id, req)

	operatorKey := name + "." + id
	requestOperatorKey := operatorKey + "." + req
	oldVersion := s.loadVersion(operatorKey)

	if oldVersion.prepared != requestOperatorKey {
		log.Fatalln(req, "Storage.Commit PREPARED BY OTHER REQUEST", oldVersion.prepared, "!=", requestOperatorKey)
		return
	}

	// Update the committed version to the prepared version.
	newVersion := version{key: oldVersion.prepared}
	updated := s.updateVersion(operatorKey, oldVersion, newVersion)
	if !updated {
		log.Fatalln(req, "Storage.Commit UPDATE FAILED")
	}

	// Delete the request version mapping.
	s.keyVersionMapping.Delete(requestOperatorKey)
	// Delete the previous operator version.
	s.versionOperatorMapping.Delete(oldVersion.key)
}

func (s *Storage) Rollback(ctx context.Context, call *jetflow.Request) error {
	ctx, span := otel.Tracer("").Start(ctx, "memory.Storage.Rollback")
	defer span.End()

	s.rollback(ctx, call.Name, call.ID, call.OperationID)
	return nil
}

func (s *Storage) rollback(ctx context.Context, name string, id string, req string) {
	log.Println(req, "Storage.Rollback", name, id, req)

	operatorKey := name + "." + id
	requestOperatorKey := operatorKey + "." + req

	// Cleanup version.
	s.keyVersionMapping.Delete(requestOperatorKey)
	s.versionOperatorMapping.Delete(requestOperatorKey)

	// Unprepare if needed.
	committedVersion := s.loadVersion(operatorKey)
	if committedVersion.prepared == requestOperatorKey {
		newCommittedVersion := committedVersion
		newCommittedVersion.prepared = ""
		updated := s.updateVersion(operatorKey, committedVersion, newCommittedVersion)
		if !updated {
			log.Fatalln(req, "Storage.Rollback UPDATE FAILED")
		}
	}
}

func (s *Storage) loadVersion(key string) version {
	v, _ := s.keyVersionMapping.Load(key)
	version, ok := v.(version)
	if !ok {
		log.Fatalln("Storage.loadVersion", key, "not found")
	}
	return version
}

func (s *Storage) loadOrCreateVersion(key string) version {
	v, _ := s.keyVersionMapping.LoadOrStore(key, version{key: key})
	return v.(version)
}

func (s *Storage) updateVersion(operatorKey string, old, version version) bool {
	return s.keyVersionMapping.CompareAndSwap(operatorKey, old, version)
}
