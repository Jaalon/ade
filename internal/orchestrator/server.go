package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"automated_dev_environment/internal/plugins/registry"
	"google.golang.org/grpc"
)

var Version = "dev"

type Server struct {
	cfg             Config
	mux             *http.ServeMux
	httpSrv         *http.Server
	grpcSrv         *grpc.Server
	errCh           chan error
	stopCh          chan struct{}
	started         bool
	mu              sync.Mutex
	projectStore    *ProjectStore
	workflowStore   *WorkflowStore
	registryStore   *registry.Store
	registryAPI     *registry.API
	runtimeSettings map[string]string
	settingsMu      sync.RWMutex
	activityStore   *ActivityStore
}

func NewServer(cfg Config) *Server {
	mux := http.NewServeMux()
	regStore := registry.NewStore()
	regAPI := registry.NewAPI(regStore)
	return &Server{
		cfg:             cfg,
		mux:             mux,
		errCh:           make(chan error, 1),
		stopCh:          make(chan struct{}),
		projectStore:    NewProjectStore(),
		workflowStore:   NewWorkflowStore(),
		registryStore:   regStore,
		registryAPI:     regAPI,
		runtimeSettings: make(map[string]string),
		activityStore:   NewActivityStore(100),
	}
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.registerAPIRoutes()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"version": Version,
	})
}

func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("serveur déjà démarré")
	}
	s.started = true
	s.mu.Unlock()

	s.setupRoutes()

	grpcSrv, _, err := startGRPCServer(s.registryStore, &s.cfg, func() map[string]string {
		s.settingsMu.RLock()
		defer s.settingsMu.RUnlock()
		cp := make(map[string]string, len(s.runtimeSettings))
		for k, v := range s.runtimeSettings {
			cp[k] = v
		}
		return cp
	})
	if err != nil {
		return err
	}
	s.grpcSrv = grpcSrv

	addr := fmt.Sprintf(":%d", s.cfg.RESTPort)
	lisHTTP, err := net.Listen("tcp", addr)
	if err != nil {
		s.grpcSrv.Stop()
		return fmt.Errorf("port %d indisponible: %w", s.cfg.RESTPort, err)
	}

	handler := recoveryMiddleware(loggingMiddleware(corsMiddleware(s.handler())))

	s.httpSrv = &http.Server{
		Handler: handler,
	}

	go func() {
		if err := s.httpSrv.Serve(lisHTTP); err != nil && err != http.ErrServerClosed {
			s.errCh <- fmt.Errorf("serveur HTTP: %w", err)
		}
	}()

	log.Printf("[orchestrateur] démarré (rest=%d, grpc=%d, version=%s)", s.cfg.RESTPort, s.cfg.GRPCPort, Version)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-s.errCh:
		s.Shutdown(context.Background())
		return err
	case <-sigCh:
		log.Printf("[orchestrateur] signal reçu, arrêt...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.Shutdown(shutdownCtx)
	case <-s.stopCh:
		return nil
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[orchestrateur] arrêt en cours...")

	if s.grpcSrv != nil {
		s.grpcSrv.GracefulStop()
	}

	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(ctx); err != nil {
			log.Printf("[orchestrateur] erreur d'arrêt HTTP: %v", err)
		}
	}

	log.Printf("[orchestrateur] arrêté")
	return nil
}

func (s *Server) Stop() {
	close(s.stopCh)
}
