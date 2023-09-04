package generate

import "testing"

func TestWriter(t *testing.T) {
	user := &Type{
		Name: "User",
	}
	user.Methods = []*Method{
		{
			Name: "TransferBalance",
			Parameters: []*Parameter{
				{Name: "U2", Type: user},
				{Name: "Amount", Type: &Type{Name: "int"}},
			},
			Results: []*Parameter{},
		},
		{
			Name: "AddBalance",
			Parameters: []*Parameter{
				{Name: "Amount", Type: &Type{Name: "int"}},
			},
			Results: []*Parameter{},
		},
	}
	s := &State{
		Package: "github.com/mathieupost/jetflow/examples/simplebank/types",
		Types: map[string]*Type{
			"User": user,
		},
	}
	w := NewWriter(s, "../examples/simplebank/types/gen")
	err := w.Write()
	if err != nil {
		t.Fatal(err)
	}
}
