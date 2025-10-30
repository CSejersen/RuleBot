package integration

import (
	"context"
	"home_automation_server/types"
)

type DiscoveryClient interface {
	Discover(ctx context.Context) ([]types.Device, []types.Entity, error)
}
