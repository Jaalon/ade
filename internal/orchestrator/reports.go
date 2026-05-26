package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type reportSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Path   string `json:"path,omitempty"`
}

func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	reports, err := s.listReportFiles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erreur de lecture des rapports")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"reports": reports,
	})
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	id := extractPathParam(r.URL.Path, "/api/v1/reports/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id du rapport requis")
		return
	}

	reportPath := filepath.Join(s.cfg.DataDir, id+".json")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "rapport introuvable")
		return
	}

	var report interface{}
	if err := json.Unmarshal(data, &report); err != nil {
		writeError(w, http.StatusInternalServerError, "format de rapport invalide")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) listReportFiles() ([]reportSummary, error) {
	entries, err := os.ReadDir(s.cfg.DataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []reportSummary{}, nil
		}
		return nil, fmt.Errorf("lecture du répertoire de rapports: %w", err)
	}

	var summaries []reportSummary
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		path := filepath.Join(s.cfg.DataDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var raw struct {
			Status string `json:"status"`
		}
		json.Unmarshal(data, &raw)

		summaries = append(summaries, reportSummary{
			ID:     id,
			Status: raw.Status,
			Path:   path,
		})
	}
	return summaries, nil
}
