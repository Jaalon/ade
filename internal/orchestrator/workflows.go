package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type WorkflowStore struct {
	mu        sync.RWMutex
	workflows map[string]*Workflow
}

func NewWorkflowStore() *WorkflowStore {
	return &WorkflowStore{
		workflows: make(map[string]*Workflow),
	}
}

func (s *WorkflowStore) Create(w *Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w.ID == "" {
		w.ID = fmt.Sprintf("wf-%d", rand.Int63())
	}
	now := time.Now()
	w.StartedAt = &now
	s.workflows[w.ID] = w
	return nil
}

func (s *WorkflowStore) Get(id string) (*Workflow, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.workflows[id]
	return w, ok
}

func (s *WorkflowStore) List() []*Workflow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Workflow, 0, len(s.workflows))
	for _, w := range s.workflows {
		out = append(out, w)
	}
	return out
}

func (s *WorkflowStore) ListByProject(projectName string) []*Workflow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Workflow
	for _, w := range s.workflows {
		if w.ProjectName == projectName {
			out = append(out, w)
		}
	}
	return out
}

func (s *WorkflowStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.workflows)
}

func extractPathParam(path, prefix string) string {
	name := strings.TrimPrefix(path, prefix)
	name = strings.TrimSuffix(name, "/")
	return name
}

func (s *Server) handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	var wf Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		writeError(w, http.StatusBadRequest, "corps de requête invalide")
		return
	}

	if strings.TrimSpace(wf.Name) == "" {
		writeError(w, http.StatusBadRequest, "le nom du workflow est requis")
		return
	}

	wf.Status = "pending"
	if err := s.workflowStore.Create(&wf); err != nil {
		writeError(w, http.StatusInternalServerError, "erreur interne")
		return
	}

	writeJSON(w, http.StatusCreated, wf)
}

func (s *Server) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	projectName := r.URL.Query().Get("project")
	var workflows []*Workflow
	if projectName != "" {
		workflows = s.workflowStore.ListByProject(projectName)
	} else {
		workflows = s.workflowStore.List()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"workflows": workflows,
	})
}

func (s *Server) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	id := extractPathParam(r.URL.Path, "/api/v1/workflows/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id du workflow requis")
		return
	}

	wf, ok := s.workflowStore.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "workflow introuvable")
		return
	}

	writeJSON(w, http.StatusOK, wf)
}
