package registry

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"automated_dev_environment/internal/plugins/contract"
)

type DockerDiscoverer struct {
	sidecar  *SidecarClient
	store    *Store
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	started  bool
	mu       sync.Mutex
}

func NewDockerDiscoverer(sidecar *SidecarClient, store *Store, interval time.Duration) *DockerDiscoverer {
	return &DockerDiscoverer{
		sidecar:  sidecar,
		store:    store,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (d *DockerDiscoverer) Start(ctx context.Context) {
	d.mu.Lock()
	if d.started {
		d.mu.Unlock()
		return
	}
	d.started = true
	d.mu.Unlock()

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		if err := d.discover(ctx); err != nil {
			log.Printf("[registry] initial discovery: %v", err)
		}

		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := d.discover(ctx); err != nil {
					log.Printf("[registry] discovery: %v", err)
				}
			case <-d.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (d *DockerDiscoverer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.started {
		return
	}
	d.started = false
	close(d.stopCh)
	d.wg.Wait()
}

func (d *DockerDiscoverer) discover(ctx context.Context) error {
	containers, err := d.sidecar.ListContainers(ctx, "ade.plugin")
	if err != nil {
		return err
	}

	for _, c := range containers {
		pluginName, ok := c.Labels["ade.plugin.name"]
		if !ok || pluginName == "" {
			continue
		}

		if c.State != "running" {
			continue
		}

		existing, exists := d.store.Get(pluginName)
		if exists && existing.PushRegistered {
			continue
		}

		httpPort := 8081
		if v, ok := c.Labels["ade.plugin.http-port"]; ok {
			if p, err := strconv.Atoi(v); err == nil {
				httpPort = p
			}
		}

		grpcPort := 50051
		if v, ok := c.Labels["ade.plugin.grpc-port"]; ok {
			if p, err := strconv.Atoi(v); err == nil {
				grpcPort = p
			}
		}

		pluginVersion := c.Labels["ade.plugin.version"]

		inst := &PluginInstance{
			Descriptor: &contract.PluginDescriptor{
				Name:         pluginName,
				Version:      pluginVersion,
				ApiVersion:   "v1",
				Capabilities: []*contract.Capability{},
				Endpoints:    map[string]string{},
			},
			GrpcAddress: fmt.Sprintf("%s:%d", c.Name, grpcPort),
			HttpAddress: fmt.Sprintf("%s:%d", c.Name, httpPort),
			Labels:      c.Labels,
		}
		inst.Descriptor.Endpoints["grpc"] = inst.GrpcAddress
		inst.Descriptor.Endpoints["http"] = inst.HttpAddress

		d.store.Register(inst)
		log.Printf("[registry] discovered plugin %q from container %s (http=%s, grpc=%s)", pluginName, c.Name, inst.HttpAddress, inst.GrpcAddress)
	}

	return nil
}
