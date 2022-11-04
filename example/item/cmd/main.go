package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"os"
	itemsV1 "wizapp/api/items/v1"
	_ "wizapp/example/item/dftapp"
	"wizapp/example/item/internal/items/service"
	"wizapp/sdk/app"
	"wizapp/sdk/grpc_server"
	"wizapp/sdk/grpc_server/grpc_gateway_server"
	"wizapp/sdk/logger"
	_ "wizapp/sdk/sql_component"
	"wizapp/sdk/temporal_server"
)

var log = logger.Log()

func main() {
	if err := app.Run(os.Args, setup); err != nil {
		log.Fatalf("unable to start application: %v", err)
	}
}

func setup(config *app.ApplicationConfig) error {

	grpclog.SetLogger(logger.Log())

	itemService := service.New()

	grpc_server.RegisterServerService(func(srv *grpc.Server) {
		itemsV1.RegisterItemServiceServer(srv, itemService)
	})

	grpc_gateway_server.RegisterServiceHandlers(func(mux *runtime.ServeMux, conn *grpc.ClientConn) error {
		return itemsV1.RegisterItemServiceHandler(context.Background(), mux, conn)
	})

	temporal_server.WorkerInterceptors()

	temporal_server.OnFatalError(func(err error) {

	})

	temporal_server.WorkerRegistry(func(w temporal_server.Worker) {
		// Register workflows and activities here
	})

	return nil
}
