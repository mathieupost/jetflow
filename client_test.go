package jetflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mathieupost/jetflow"
)

type TestType interface{}

type TestTypeProxy struct {
	id string
}

func (t TestTypeProxy) ID() string {
	return t.id
}

func TestClientFind(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mapping := jetflow.ProxyFactoryMapping{
		"TestType": func(id string, client jetflow.OperatorClient) jetflow.Operator {
			return &TestTypeProxy{id: id}
		},
	}
	var client jetflow.OperatorClient = jetflow.NewClient(mapping, nil)
	var testType TestType
	err := client.Find(ctx, "test_type", &testType)
	require.NoError(t, err)
	require.NoError(t, ctx.Err())
	require.NotNil(t, testType)
	require.IsType(t, &TestTypeProxy{}, testType)
}
