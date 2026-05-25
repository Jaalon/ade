package sdk

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/ade/plugins-sdk/contract"
)

type capabilityRegistry struct {
	mu   sync.RWMutex
	caps []*contract.Capability
}

func newCapabilityRegistry() *capabilityRegistry {
	return &capabilityRegistry{}
}

func (r *capabilityRegistry) add(cap *contract.Capability) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.caps = append(r.caps, cap)
}

func (r *capabilityRegistry) list() []*contract.Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*contract.Capability, len(r.caps))
	copy(out, r.caps)
	return out
}

func (r *capabilityRegistry) serveHTTP(w http.ResponseWriter, req *http.Request) {
	caps := r.list()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"capabilities": caps}); err != nil {
		http.Error(w, "failed to encode capabilities", http.StatusInternalServerError)
	}
}
