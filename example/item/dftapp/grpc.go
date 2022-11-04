package dftapp

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"wizapp/sdk/grpc_server"
	"wizapp/sdk/grpc_server/grpc_gateway_server"
)

func init() {
	grpc_gateway_server.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))

	grpc_server.WithServerOption(grpc_middleware.WithUnaryServerChain(
		grpc_recovery.UnaryServerInterceptor(),
	))

}
