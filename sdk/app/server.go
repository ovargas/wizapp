package app

import (
	"log"
)

type (
	Server interface {
		Start() error
		Stop() error
	}

	ServerFactory func(config *ApplicationConfig) (Server, error)
)

var (
	serverFactories = make(map[string]ServerFactory)
)

func RegisterServer(name string, factory ServerFactory) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := serverFactories[name]; ok {
		log.Fatalf("Server %s already registered", name)
	}

	serverFactories[name] = factory
}
