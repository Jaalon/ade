package orchestrator

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func (s *Server) registerAPIRoutes() {
	s.mux.Handle("/api/v1/plugins/", s.handlePluginRouter())
	s.mux.HandleFunc("/api/v1/projects", s.handleProjectsRouter)
	s.mux.HandleFunc("/api/v1/projects/", s.handleProjectItemRouter)
	s.mux.HandleFunc("/api/v1/workflows", s.handleWorkflowsRouter)
	s.mux.HandleFunc("/api/v1/workflows/", s.handleWorkflowItemRouter)
	s.mux.HandleFunc("/api/v1/config", s.handleConfigRouter)
	s.mux.HandleFunc("/api/v1/reports", s.handleReportsRouter)
	s.mux.HandleFunc("/api/v1/reports/", s.handleReportItemRouter)
	s.mux.HandleFunc("/api/v1/dashboard/stats", s.handleDashboardStats)
	s.mux.HandleFunc("/api/v1/dashboard/activity", s.handleDashboardActivity)
	s.mux.HandleFunc("/api/v1/ws", s.handleWebSocket)
}

func (s *Server) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasPrefix(path, "/api/") || path == "/health" {
			s.mux.ServeHTTP(w, r)
			return
		}

		s.serveStaticOrSPA(w, r)
	})
}

func (s *Server) serveStaticOrSPA(w http.ResponseWriter, r *http.Request) {
	clean := strings.TrimPrefix(r.URL.Path, "/")
	if clean == "" {
		clean = "index.html"
	}

	fs := s.staticFileSystem()

	f, err := fs.Open(clean)
	if err != nil {
		s.serveEmbeddedFile(w, r, "index.html")
		return
	}
	f.Close()

	if filepath.Ext(clean) == "" {
		s.serveEmbeddedFile(w, r, "index.html")
		return
	}

	http.FileServer(fs).ServeHTTP(w, r)
}

func (s *Server) serveEmbeddedFile(w http.ResponseWriter, r *http.Request, name string) {
	fs := s.staticFileSystem()
	data, err := fs.Open(name)
	if err != nil {
		log.Printf("[orchestrateur] frontend non disponible (%s): %v", name, err)
		s.mux.ServeHTTP(w, r)
		return
	}
	defer data.Close()

	stat, err := data.Stat()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	buf := make([]byte, stat.Size())
	data.Read(buf)

	if filepath.Ext(name) == ".html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func (s *Server) handlePluginRouter() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/")

		if strings.HasSuffix(path, "/health") && r.Method == http.MethodGet {
			s.handlePluginHealth(w, r)
			return
		}

		s.registryAPI.ServeHTTP(w, r)
	})
}

func (s *Server) handlePluginHealth(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/plugins/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "nom du plugin requis")
		return
	}
	name := parts[0]

	inst, ok := s.registryStore.Get(name)
	if !ok {
		writeError(w, http.StatusNotFound, "plugin introuvable")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":   inst.Descriptor.Name,
		"status": inst.Status.String(),
	})
}

func (s *Server) handleProjectsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListProjects(w, r)
	case http.MethodPost:
		s.handleCreateProject(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
	}
}

func (s *Server) handleProjectItemRouter(w http.ResponseWriter, r *http.Request) {
	name := extractPathParam(r.URL.Path, "/api/v1/projects/")
	if name == "" || strings.Contains(name, "/") {
		writeError(w, http.StatusBadRequest, "nom du projet requis")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetProject(w, r)
	case http.MethodPut:
		s.handleUpdateProject(w, r)
	case http.MethodDelete:
		s.handleDeleteProject(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
	}
}

func (s *Server) handleWorkflowsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListWorkflows(w, r)
	case http.MethodPost:
		s.handleCreateWorkflow(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
	}
}

func (s *Server) handleWorkflowItemRouter(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r.URL.Path, "/api/v1/workflows/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id du workflow requis")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetWorkflow(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
	}
}

func (s *Server) handleConfigRouter(w http.ResponseWriter, r *http.Request) {
	_ = extractPathParam(r.URL.Path, "/api/v1/config")
	switch r.Method {
	case http.MethodGet:
		s.handleGetConfig(w, r)
	case http.MethodPut:
		s.handleUpdateConfig(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
	}
}

func (s *Server) handleReportsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListReports(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
	}
}

func (s *Server) handleReportItemRouter(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r.URL.Path, "/api/v1/reports/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id du rapport requis")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetReport(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
	}
}
