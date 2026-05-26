package orchestrator

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ProjectStore struct {
	mu       sync.RWMutex
	projects map[string]*Project
}

func NewProjectStore() *ProjectStore {
	return &ProjectStore{
		projects: make(map[string]*Project),
	}
}

func (s *ProjectStore) Create(p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[p.Name]; exists {
		return ErrProjectAlreadyExists
	}

	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	s.projects[p.Name] = p
	return nil
}

func (s *ProjectStore) Get(name string) (*Project, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[name]
	return p, ok
}

func (s *ProjectStore) List() []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Project, 0, len(s.projects))
	for _, p := range s.projects {
		out = append(out, p)
	}
	return out
}

func (s *ProjectStore) Update(p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.projects[p.Name]
	if !exists {
		return ErrProjectNotFound
	}

	existing.Description = p.Description
	existing.Labels = p.Labels
	existing.UpdatedAt = time.Now()
	return nil
}

func (s *ProjectStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[name]; !exists {
		return ErrProjectNotFound
	}
	delete(s.projects, name)
	return nil
}

func (s *ProjectStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.projects)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	var p Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "corps de requête invalide")
		return
	}

	if strings.TrimSpace(p.Name) == "" {
		writeError(w, http.StatusBadRequest, "le nom du projet est requis")
		return
	}

	if err := s.projectStore.Create(&p); err != nil {
		if err == ErrProjectAlreadyExists {
			writeError(w, http.StatusConflict, "conflit: un projet avec ce nom existe déjà")
			return
		}
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	projects := s.projectStore.List()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
	})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	name := extractPathParam(r.URL.Path, "/api/v1/projects/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "nom du projet requis")
		return
	}

	p, ok := s.projectStore.Get(name)
	if !ok {
		writeError(w, http.StatusNotFound, "projet introuvable")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	name := extractPathParam(r.URL.Path, "/api/v1/projects/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "nom du projet requis")
		return
	}

	var updated Project
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeError(w, http.StatusBadRequest, "corps de requête invalide")
		return
	}
	updated.Name = name

	if err := s.projectStore.Update(&updated); err != nil {
		if err == ErrProjectNotFound {
			writeError(w, http.StatusNotFound, "projet introuvable")
			return
		}
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	p, _ := s.projectStore.Get(name)
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	name := extractPathParam(r.URL.Path, "/api/v1/projects/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "nom du projet requis")
		return
	}

	if err := s.projectStore.Delete(name); err != nil {
		if err == ErrProjectNotFound {
			writeError(w, http.StatusNotFound, "projet introuvable")
			return
		}
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "projet supprimé"})
}
