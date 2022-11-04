package app

import (
	"log"
)

type (
	Component interface {
		Command() *Command
	}

	ComponentFactory func(config *ApplicationConfig) (Component, error)
)

var (
	componentFactories = make(map[string]ComponentFactory)
)

func RegisterComponent(name string, factory ComponentFactory) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := componentFactories[name]; ok {
		log.Fatalf("Component %s already registered", name)
	}

	componentFactories[name] = factory
}
