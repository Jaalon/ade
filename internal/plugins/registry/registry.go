package registry

import (
	"context"
	"time"
)

type Registry struct {
	Store         *Store
	Discoverer    *DockerDiscoverer
	HealthChecker *HealthChecker
	SidecarClient *SidecarClient
}

func NewRegistry(sidecarURL string, discoveryInterval time.Duration, healthInterval time.Duration, maxFails int) *Registry {
	store := NewStore()
	sidecar := NewSidecarClient(sidecarURL)
	discoverer := NewDockerDiscoverer(sidecar, store, discoveryInterval)
	health := NewHealthChecker(store, healthInterval, maxFails)

	return &Registry{
		Store:         store,
		Discoverer:    discoverer,
		HealthChecker: health,
		SidecarClient: sidecar,
	}
}

func (r *Registry) Start(ctx context.Context) {
	r.Discoverer.Start(ctx)
	r.HealthChecker.Start(ctx)
}

func (r *Registry) Stop() {
	r.HealthChecker.Stop()
	r.Discoverer.Stop()
}
