package app

import (
	"context"
	"log"

	"github.com/sk25469/kv/internal/network"
	"github.com/sk25469/kv/utils"
	"go.uber.org/fx"
)

type ApplicationParams struct {
	fx.In

	Lifecycle    fx.Lifecycle
	NetworkLayer *network.NetworkService
	HealthCheck  *network.HealthCheckService
}

func StartApplication(params ApplicationParams) {
	params.Lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := params.NetworkLayer.Start(); err != nil {
						log.Fatalf("Error starting network layer: %v", err)
					}
				}()

				params.HealthCheck.StartHealthCheck()
				params.HealthCheck.StartPeriodicHealthChecks(utils.HEALTH_CHECK_INTERVAL)

				return nil
			},
			OnStop: func(ctx context.Context) error {
				return params.NetworkLayer.Stop()
			},
		},
	)
}
