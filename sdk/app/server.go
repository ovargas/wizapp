package app

import (
	"errors"
	"log"
)

type (
	Server interface {
		Start() error
		Stop() error
		mustImplementServer()
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

type UnimplementedServer struct{}

func (u UnimplementedServer) Start() error {
	return errors.New("method Start not implemented")
}

func (u UnimplementedServer) Stop() error {
	return errors.New("method Stop not implemented")
}

func (u UnimplementedServer) mustImplementServer() {
}
