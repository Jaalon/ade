package orchestrator

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ActivityStore struct {
	mu     sync.RWMutex
	events []ActivityEvent
	limit  int
	nextID int
}

func NewActivityStore(limit int) *ActivityStore {
	return &ActivityStore{
		events: make([]ActivityEvent, 0, limit),
		limit:  limit,
	}
}

func (a *ActivityStore) Add(eventType, message string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.nextID++
	evt := ActivityEvent{
		ID:        fmt.Sprintf("evt-%d", a.nextID),
		Type:      eventType,
		Message:   message,
		Timestamp: time.Now(),
	}

	if len(a.events) >= a.limit {
		a.events = a.events[1:]
	}
	a.events = append(a.events, evt)
}

func (a *ActivityStore) List() []ActivityEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()

	out := make([]ActivityEvent, len(a.events))
	copy(out, a.events)
	return out
}

func (a *ActivityStore) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.events)
}

func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	plugins := s.registryStore.List()
	activeCount := 0
	for _, p := range plugins {
		if p.Status.String() == "HEALTHY" {
			activeCount++
		}
	}

	stats := DashboardStats{
		Projects:      s.projectStore.Count(),
		Plugins:       len(plugins),
		Workflows:     s.workflowStore.Count(),
		Reports:       s.countReports(),
		ActivePlugins: activeCount,
	}

	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleDashboardActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	events := s.activityStore.List()
	if events == nil {
		events = []ActivityEvent{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
	})
}

func (s *Server) countReports() int {
	summaries, err := s.listReportFiles()
	if err != nil {
		return 0
	}
	return len(summaries)
}
