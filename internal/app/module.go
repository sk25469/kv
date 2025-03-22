package app

import (
	"github.com/sk25469/kv/internal/codec"
	"github.com/sk25469/kv/internal/comm"
	"github.com/sk25469/kv/internal/core"
	"github.com/sk25469/kv/internal/middleware"
	"github.com/sk25469/kv/internal/network"
	node_config "github.com/sk25469/kv/internal/network/model"
	"github.com/sk25469/kv/internal/replication"
	"github.com/sk25469/kv/internal/storage"
	"go.uber.org/fx"
)

// Module provides all fx options for the application
func Module(configPath string) fx.Option {
	return fx.Options(
		fx.Provide(
			NewConfig(configPath),
			comm.NewCommunicationService,
			replication.NewReplicationService,
			storage.NewStorage,
			middleware.NewStorageMiddleware,
			middleware.NewCacheMiddleware,
			core.NewCoreService,
			codec.NewCodecLayerService,
			network.NewNetworkService,
			network.NewHealthCheckService,
		),
		fx.Invoke(StartApplication),
	)
}

// Config provider
func NewConfig(configPath string) func() (*node_config.NodeConfig, error) {
	return func() (*node_config.NodeConfig, error) {
		return node_config.NewNodeConfig(configPath), nil
	}
}
