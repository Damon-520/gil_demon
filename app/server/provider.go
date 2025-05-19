package server

import "github.com/google/wire"

// ProviderSet is server providers.
var ServerProviderSet = wire.NewSet(
	NewGinHttpServer, // HTTP
	NewGRPCServer,    // GRPC
	NewServer,
)
