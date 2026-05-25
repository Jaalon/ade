package registry

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	api    *API
	server *http.Server
	port   int
}

func NewServer(api *API, port int) *Server {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/plugins/", api)
	mux.Handle("/api/v1/plugins", api)

	return &Server{
		api:  api,
		port: port,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	log.Printf("[registry] starting registry server on port %d", s.port)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}
