package memory

import (
	"context"
	"log"
	"sync"

	"github.com/huandu/go-clone"

	"github.com/mathieupost/jetflow"
)

type Storage struct {
	typeHandlerMapping     map[string]func(id string) jetflow.OperatorHandler
	keyVersionMapping      sync.Map
	versionOperatorMapping sync.Map
}

type version struct {
	base     string
	key      string
	prepared string
}

func NewStorage(mapping map[string]func(id string) jetflow.OperatorHandler) *Storage {
	return &Storage{
		typeHandlerMapping: mapping,
	}
}

func (s *Storage) Get(ctx context.Context, otype string, oid string, requestID string) jetflow.OperatorHandler {
	// Get last committed value. Create a new value if not found.
	operatorKey := otype + "." + oid
	requestOperatorKey := operatorKey + "." + requestID
	log.Println("Storage.Get", requestOperatorKey)

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
				s.typeHandlerMapping[otype](oid),
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

func (s *Storage) Prepare(ctx context.Context, otype string, oid string, requestID string) (ok bool) {
	log.Println("Storage.Prepare", otype, oid, requestID)

	operatorKey := otype + "." + oid
	requestOperatorKey := operatorKey + "." + requestID
	committedVersion := s.loadVersion(operatorKey)

	defer func() {
		if !ok {
			log.Println("Storage.Prepare FAILED", operatorKey)
			// Clean up
			s.keyVersionMapping.Delete(requestOperatorKey)
			s.versionOperatorMapping.Delete(requestOperatorKey)
		}
	}()

	if committedVersion.prepared != "" {
		// Already prepared by another request.
		return false
	}

	requestVersion := s.loadVersion(requestOperatorKey)

	// Check if the version for the given request is created
	// from the committed version.
	if requestVersion.base != committedVersion.key {
		// A new version is already committed.
		return false
	}

	// Copy and set prepared to the version for the given request.
	newCommittedVersion := committedVersion
	newCommittedVersion.prepared = requestOperatorKey
	return s.updateVersion(operatorKey, committedVersion, newCommittedVersion)
}

func (s *Storage) Commit(ctx context.Context, otype string, oid string, requestID string) (ok bool) {
	log.Println("Storage.Commit", otype, oid, requestID)

	operatorKey := otype + "." + oid
	requestOperatorKey := operatorKey + "." + requestID
	oldVersion := s.loadVersion(operatorKey)

	defer func() {
		if !ok {
			log.Fatalln("Storage.Commit FAILED")
		}
		// Delete the request version mapping.
		s.keyVersionMapping.Delete(requestOperatorKey)
		// Delete the previous operator version.
		s.versionOperatorMapping.Delete(oldVersion.key)
	}()

	if oldVersion.prepared != requestOperatorKey {
		log.Println("Storage.Commit PREPARED BY OTHER REQUEST", otype, oid, requestID)
		return false
	}

	// Update the committed version to the prepared version.
	newVersion := version{key: oldVersion.prepared}
	return s.updateVersion(operatorKey, oldVersion, newVersion)
}

func (s *Storage) loadVersion(key string) version {
	v, _ := s.keyVersionMapping.Load(key)
	return v.(version)
}

func (s *Storage) loadOrCreateVersion(key string) version {
	v, _ := s.keyVersionMapping.LoadOrStore(key, version{key: key})
	return v.(version)
}

func (s *Storage) updateVersion(operatorKey string, old, version version) bool {
	return s.keyVersionMapping.CompareAndSwap(operatorKey, old, version)
}
