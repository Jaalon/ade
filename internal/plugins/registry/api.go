package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"automated_dev_environment/internal/plugins/contract"
)

type API struct {
	store    *Store
	register func(*contract.RegisterRequest) (*PluginInstance, error)
}

func NewAPI(store *Store) *API {
	return &API{
		store: store,
		register: func(req *contract.RegisterRequest) (*PluginInstance, error) {
			inst := &PluginInstance{
				Descriptor: &contract.PluginDescriptor{
					Name:         req.Name,
					Version:      req.Version,
					Description:  req.Description,
					ApiVersion:   req.ApiVersion,
					Capabilities: req.Capabilities,
					Endpoints:    req.Endpoints,
				},
				GrpcAddress:    req.GrpcAddress,
				HttpAddress:    req.HttpAddress,
				PushRegistered: true,
			}
			if inst.Descriptor.Endpoints == nil {
				inst.Descriptor.Endpoints = make(map[string]string)
			}
			inst.Descriptor.Endpoints["grpc"] = req.GrpcAddress
			inst.Descriptor.Endpoints["http"] = req.HttpAddress

			store.Register(inst)
			log.Printf("[registry] plugin %q registered via push (http=%s, grpc=%s)", req.Name, req.HttpAddress, req.GrpcAddress)
			return inst, nil
		},
	}
}

func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req contract.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	inst, err := a.register(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"accepted": true,
		"plugin":   inst,
	})
}

func (a *API) ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	plugins := a.store.List()
	if plugins == nil {
		plugins = []*PluginInstance{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plugins": plugins,
	})
}

func (a *API) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name := extractName(r.URL.Path, "/api/v1/plugins/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "plugin name is required")
		return
	}

	inst, ok := a.store.Get(name)
	if !ok {
		writeError(w, http.StatusNotFound, "plugin not found")
		return
	}

	writeJSON(w, http.StatusOK, inst)
}

func (a *API) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name := extractName(r.URL.Path, "/api/v1/plugins/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "plugin name is required")
		return
	}

	if err := a.store.Unregister(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "plugin unregistered"})
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	switch {
	case path == "/api/v1/plugins/register":
		a.RegisterHandler(w, r)
	case path == "/api/v1/plugins" || path == "/api/v1/plugins/":
		a.ListHandler(w, r)
	case strings.HasPrefix(path, "/api/v1/plugins/"):
		if r.Method == http.MethodDelete {
			a.DeleteHandler(w, r)
		} else {
			a.GetHandler(w, r)
		}
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func extractName(path, prefix string) string {
	name := strings.TrimPrefix(path, prefix)
	name = strings.TrimSuffix(name, "/")
	return name
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
