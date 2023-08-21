package generate_test

import (
	"testing"

	"github.com/mathieupost/jetflow/generate"
)

func TestParsePackage(t *testing.T) {
	generate.ParsePackage("../example/types")
}
