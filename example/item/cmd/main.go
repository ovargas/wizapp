package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	itemsV1 "github.com/ovargas/wizapp/example/api/items/v1"
	_ "github.com/ovargas/wizapp/example/item/dftapp"
	"github.com/ovargas/wizapp/example/item/internal/items/service"
	"github.com/ovargas/wizapp/sdk/app"
	"github.com/ovargas/wizapp/sdk/grpc_server"
	"github.com/ovargas/wizapp/sdk/grpc_server/grpc_gateway_server"
	"github.com/ovargas/wizapp/sdk/logger"
	_ "github.com/ovargas/wizapp/sdk/sql_component"
	"github.com/ovargas/wizapp/sdk/temporal_server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"os"
)

var log = logger.Log()

func main() {
	app.Usage = "Item demo application"

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
