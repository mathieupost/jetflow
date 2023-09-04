package memory

import (
	"context"
	"testing"

	"github.com/mathieupost/jetflow"
	"github.com/stretchr/testify/require"
)

func TestPrepare(t *testing.T) {
	ctx := context.Background()
	var s jetflow.Storage
	s = NewStorage(jetflow.HandlerFactoryMapping{
		"TestType": NewTestTypeHandler,
	})

	setup := func(ctx context.Context, trID, opID string) (jetflow.OperatorHandler, *jetflow.Request) {
		call := &jetflow.Request{
			OperationID: trID,
			RequestID:   "req_id",
			Name:        "TestType",
			ID:          opID,
			Args:        []byte("1"),
		}
		operator, err := s.Get(ctx, call)
		require.NoError(t, err)
		return operator, call
	}

	t.Run("Success", func(t *testing.T) {
		operator, call := setup(ctx, "1", t.Name())
		operator.Handle(ctx, nil, call)
		err := s.Prepare(ctx, call)
		require.NoError(t, err)
	})

	t.Run("AlreadyPrepared", func(t *testing.T) {
		operator, call1 := setup(ctx, "1", t.Name())
		operator.Handle(ctx, nil, call1)
		operator, call2 := setup(ctx, "2", t.Name())
		operator.Handle(ctx, nil, call2)

		err := s.Prepare(ctx, call1)
		require.NoError(t, err)

		err = s.Prepare(ctx, call2)
		require.ErrorContains(t, err, "already prepared")
	})

	t.Run("AlreadyPreparedButNoChanges", func(t *testing.T) {
		operator, call1 := setup(ctx, "1", t.Name())
		operator.Handle(ctx, nil, call1)
		operator, call2 := setup(ctx, "2", t.Name())
		call2.Args = []byte{}
		operator.Handle(ctx, nil, call2)

		err := s.Prepare(ctx, call1)
		require.NoError(t, err)

		err = s.Prepare(ctx, call2)
		require.NoError(t, err)
	})

	t.Run("BaseOutdated", func(t *testing.T) {
		operator, call1 := setup(ctx, "1", t.Name())
		operator.Handle(ctx, nil, call1)
		operator, call2 := setup(ctx, "2", t.Name())
		operator.Handle(ctx, nil, call2)

		err := s.Prepare(ctx, call1)
		require.NoError(t, err)
		err = s.Commit(ctx, call1)
		require.NoError(t, err)

		err = s.Prepare(ctx, call2)
		require.ErrorContains(t, err, "base outdated")
	})

}

type TestTypeHandler struct {
	instance TestType
}

func (h *TestTypeHandler) Handle(ctx context.Context, client jetflow.OperatorClient, call *jetflow.Request) (bytes []byte, err error) {
	h.instance.(*testType).field += len(call.Args)
	return
}

func NewTestTypeHandler(id string) jetflow.OperatorHandler {
	instance := NewTestType(id)
	return &TestTypeHandler{instance}
}

type TestType interface {
	jetflow.Operator
}

func NewTestType(id string) TestType {
	return &testType{id: id, field: 1}
}

type testType struct {
	id    string
	field int
}

// ID implements jetflow.Operator interface.
func (t *testType) ID() string {
	return t.id
}
