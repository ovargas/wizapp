package grpc_gateway_server

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"net/http"
	"wizapp/sdk/app"
	"wizapp/sdk/grpc_server"
	"wizapp/sdk/logger"
)

var (
	log                    = logger.Log()
	registerServiceHandler []func(mux *runtime.ServeMux, conn *grpc.ClientConn) error
)

const (
	ServerName = "grpc-gateway"
)

type (
	Config struct {
		grpc_server.Config `mapstructure:",squash"`
		Gateway            Gateway `mapstructure:"gateway"`
	}

	Gateway struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	server struct {
		server    *http.Server
		isStarted bool
	}
)

var (
	dialOptions    []grpc.DialOption
	serveMuxOption []runtime.ServeMuxOption
)

func init() {
	app.RegisterServer(ServerName, createServer)
}

func RegisterServiceHandlers(fn ...func(mux *runtime.ServeMux, conn *grpc.ClientConn) error) {
	registerServiceHandler = append(registerServiceHandler, fn...)
}

func WithDialOption(options ...grpc.DialOption) {
	dialOptions = append(dialOptions, options...)
}

func WithServeMuxOption(options ...runtime.ServeMuxOption) {
	serveMuxOption = append(serveMuxOption, options...)
}

func createServer(config *app.ApplicationConfig) (app.Server, error) {
	var gwCfg Config
	if err := config.UnmarshalKey(grpc_server.ConfigKey, &gwCfg); err != nil {
		return nil, err
	}

	connection, err := grpc.Dial(fmt.Sprintf("%s:%d", gwCfg.Host, gwCfg.Port), dialOptions...)

	if err != nil {
		return nil, err
	}

	mux := runtime.NewServeMux(serveMuxOption...)

	for _, register := range registerServiceHandler {
		if err := register(mux, connection); err != nil {
			return nil, err
		}
	}
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", gwCfg.Gateway.Host, gwCfg.Gateway.Port),
		Handler: mux,
	}

	return &server{
		server: httpServer,
	}, nil
}

func (s *server) Start() error {
	if s.isStarted {
		return nil
	}

	if err := s.server.ListenAndServe(); err != nil {
		log.Errorf("error starting gateway server: %v", err)
		return err
	}
	return nil
}

func (s *server) Stop() error {
	if !s.isStarted {
		return nil
	}

	if err := s.server.Shutdown(context.TODO()); err != nil {
		log.Errorf("error stopping gateway server: %v", err)
	}
	return nil
}
