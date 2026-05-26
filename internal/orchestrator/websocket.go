package orchestrator

import (
	"log"
	"net/http"
)

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "méthode non autorisée")
		return
	}

	log.Printf("[orchestrateur] tentative de connexion WebSocket depuis %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("WebSocket non implémenté. Utilisez l'API REST ou gRPC."))
}
