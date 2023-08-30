package generate

type State struct {
	Package string
	Types   map[string]*Type
}

type Type struct {
	Name    string
	Methods []*Method
}

type Method struct {
	Name       string
	Parameters []*Parameter
	Results    []*Parameter
}

type Parameter struct {
	Name string
	Type *Type
}
