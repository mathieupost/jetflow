package gen

import "github.com/mathieupost/jetflow"

func ProxyFactoryMapping() jetflow.ProxyFactoryMapping {
	return map[string]jetflow.ProxyFactory{
		"User": NewUserProxy,
	}
}

func HandlerFactoryMapping() jetflow.HandlerFactoryMapping {
	return map[string]jetflow.HandlerFactory{
		"User": NewUserHandler,
	}
}
