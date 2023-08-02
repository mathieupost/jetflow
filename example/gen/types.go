package gen

import "github.com/mathieupost/jetflow"

func FactoryMapping() map[string]jetflow.OperatorFactory {
	return map[string]jetflow.OperatorFactory{
		"User": NewUserProxy,
	}
}
