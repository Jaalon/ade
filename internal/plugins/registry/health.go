package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"automated_dev_environment/internal/plugins/contract"
)

type HealthChecker struct {
	store    *Store
	interval time.Duration
	maxFails int
	client   *http.Client
	stopCh   chan struct{}
	wg       sync.WaitGroup
	started  bool
	mu       sync.Mutex
}

func NewHealthChecker(store *Store, interval time.Duration, maxFails int) *HealthChecker {
	return &HealthChecker{
		store:    store,
		interval: interval,
		maxFails: maxFails,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

func (h *HealthChecker) Start(ctx context.Context) {
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return
	}
	h.started = true
	h.mu.Unlock()

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.checkAll(ctx)
			case <-h.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (h *HealthChecker) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.started {
		return
	}
	h.started = false
	close(h.stopCh)
	h.wg.Wait()
}

func (h *HealthChecker) checkAll(ctx context.Context) {
	plugins := h.store.List()
	for _, p := range plugins {
		h.checkOne(ctx, p)
	}
}

func (h *HealthChecker) checkOne(ctx context.Context, p *PluginInstance) {
	url := fmt.Sprintf("http://%s/health", p.HttpAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		h.recordFailure(p, err.Error())
		return
	}

	resp, err := h.client.Do(req)
	if err != nil {
		h.recordFailure(p, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.recordFailure(p, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return
	}

	var hs contract.HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&hs); err != nil {
		h.recordFailure(p, "invalid health response")
		return
	}

	h.store.UpdateHealth(p.Descriptor.Name, hs.Status)
}

func (h *HealthChecker) recordFailure(p *PluginInstance, reason string) {
	h.store.UpdateHealth(p.Descriptor.Name, contract.HealthUnhealthy)

	updated, ok := h.store.Get(p.Descriptor.Name)
	if !ok {
		return
	}

	if updated.FailedChecks >= h.maxFails {
		log.Printf("[registry] plugin %q removed after %d failed health checks: %s", p.Descriptor.Name, updated.FailedChecks, reason)
		if err := h.store.Unregister(p.Descriptor.Name); err != nil {
			log.Printf("[registry] failed to unregister %q: %v", p.Descriptor.Name, err)
		}
	}
}
