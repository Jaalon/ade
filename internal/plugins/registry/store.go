package registry

import (
	"sync"
	"time"

	"automated_dev_environment/internal/plugins/contract"
)

type PluginInstance struct {
	Descriptor     *contract.PluginDescriptor `json:"descriptor"`
	Status         contract.HealthStatusEnum  `json:"status"`
	LastSeen       time.Time                  `json:"last_seen"`
	FailedChecks   int                        `json:"failed_checks"`
	GrpcAddress    string                     `json:"grpc_address,omitempty"`
	HttpAddress    string                     `json:"http_address,omitempty"`
	Labels         map[string]string          `json:"labels,omitempty"`
	PushRegistered bool                       `json:"-"`
}

type Store struct {
	plugins map[string]*PluginInstance
	mu      sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		plugins: make(map[string]*PluginInstance),
	}
}

func (s *Store) Register(inst *PluginInstance) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst.LastSeen = time.Now()
	inst.Status = contract.HealthHealthy
	inst.FailedChecks = 0

	existing, exists := s.plugins[inst.Descriptor.Name]
	if exists {
		inst.Status = existing.Status
		inst.FailedChecks = existing.FailedChecks
	}

	s.plugins[inst.Descriptor.Name] = inst
}

func (s *Store) Unregister(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.plugins[name]; !ok {
		return contract.ErrPluginNotFound
	}
	delete(s.plugins, name)
	return nil
}

func (s *Store) Get(name string) (*PluginInstance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst, ok := s.plugins[name]
	return inst, ok
}

func (s *Store) List() []*PluginInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*PluginInstance, 0, len(s.plugins))
	for _, inst := range s.plugins {
		out = append(out, inst)
	}
	return out
}

func (s *Store) UpdateHealth(name string, status contract.HealthStatusEnum) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.plugins[name]
	if !ok {
		return
	}

	inst.Status = status
	inst.LastSeen = time.Now()

	if status == contract.HealthHealthy {
		inst.FailedChecks = 0
	} else {
		inst.FailedChecks++
	}
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.plugins)
}
