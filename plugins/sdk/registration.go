package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ade/plugins-sdk/contract"
)

type Registration struct {
	descriptor *contract.PluginDescriptor
	orchURL    string
	interval   time.Duration
	client     *http.Client
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
	stopped    bool
}

func NewRegistration(descriptor *contract.PluginDescriptor, orchURL string, interval time.Duration) *Registration {
	return &Registration{
		descriptor: descriptor,
		orchURL:    orchURL,
		interval:   interval,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

func (r *Registration) Start(ctx context.Context) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := r.RegisterNow(ctx); err != nil {
			log.Printf("[sdk] initial registration failed: %v (will retry)", err)
		}
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := r.RegisterNow(ctx); err != nil {
					log.Printf("[sdk] re-registration failed: %v", err)
				}
			case <-r.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *Registration) RegisterNow(ctx context.Context) error {
	req := &contract.RegisterRequest{
		Name:         r.descriptor.Name,
		Version:      r.descriptor.Version,
		ApiVersion:   r.descriptor.ApiVersion,
		Description:  r.descriptor.Description,
		GrpcAddress:  r.descriptor.Endpoints["grpc"],
		HttpAddress:  r.descriptor.Endpoints["http"],
		Capabilities: r.descriptor.Capabilities,
		Endpoints:    r.descriptor.Endpoints,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.orchURL+"/api/v1/plugins/register", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrRegistration
	}

	return nil
}

func (r *Registration) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopped {
		return
	}
	r.stopped = true
	close(r.stopCh)
	r.wg.Wait()
}
