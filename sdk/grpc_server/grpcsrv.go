package grpc_server

import (
	"fmt"
	"github.com/ovargas/wizapp/sdk/app"
	"github.com/ovargas/wizapp/sdk/logger"
	"google.golang.org/grpc"
	"net"
)

const (
	ServerName = "grpc"
	ConfigKey  = "grpc"
)

type (
	server struct {
		app.UnimplementedServer
		isStarted bool
		server    *grpc.Server
		config    *Config
	}

	Config struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}
)

var (
	log                            = logger.Log()
	serverOptions                  []grpc.ServerOption
	registerServerServiceFunctions []func(srv *grpc.Server)
)

func init() {
	app.RegisterServer(ServerName, createServer)
}

func WithServerOption(options ...grpc.ServerOption) {
	serverOptions = append(serverOptions, options...)
}

func RegisterServerService(fn ...func(srv *grpc.Server)) {
	registerServerServiceFunctions = append(registerServerServiceFunctions, fn...)
}

func createServer(config *app.ApplicationConfig) (app.Server, error) {
	var grpcCfg Config
	if err := config.UnmarshalKey(ConfigKey, &grpcCfg); err != nil {
		return nil, err
	}
	s := grpc.NewServer(serverOptions...)

	for _, fn := range registerServerServiceFunctions {
		fn(s)
	}

	return &server{
		config: &grpcCfg,
		server: s,
	}, nil
}

func (s *server) Start() error {
	if s.isStarted {
		return nil
	}
	s.isStarted = true

	log.Infof("starting listener", s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
	if err != nil {
		log.Fatalf("unable to start listener %s:%d, %v", s.config.Host, s.config.Port, err)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatalf("Unable to close listener %s:%d, %v", s.config.Host, s.config.Port, err)
		}
	}(listener)

	defer func() {
		s.isStarted = false
	}()

	if err := s.server.Serve(listener); err != nil {
		log.Fatalf("Unable to serve grpc in listener %s:%d, %v", s.config.Host, s.config.Port, err)
	}

	return nil
}

func (s *server) Stop() error {
	if s.isStarted {
		s.server.GracefulStop()
		log.Infof("grpc server stopped")
	}
	return nil
}
