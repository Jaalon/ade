package sdk

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"automated_dev_environment/internal/plugins/contract"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type PluginServer struct {
	cfg     Config
	grpcSrv *grpc.Server
	httpSrv *http.Server
	mux     *http.ServeMux
	caps    *capabilityRegistry
	health  *healthService
	reg     *Registration
	stopCh  chan struct{}
	errCh   chan error
	started bool
	mu      sync.Mutex
}

type Option func(*PluginServer)

func WithConfig(cfg Config) Option {
	return func(s *PluginServer) {
		s.cfg = cfg
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(s *PluginServer) {
		_ = logger
	}
}

func WithGRPCOptions(opts ...grpc.ServerOption) Option {
	return func(s *PluginServer) {
		s.grpcSrv = grpc.NewServer(opts...)
	}
}

func NewPlugin(opts ...Option) (*PluginServer, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	hs := &healthService{
		getStatus: func() contract.HealthStatusEnum {
			return contract.HealthHealthy
		},
	}

	s := &PluginServer{
		cfg:    cfg,
		mux:    http.NewServeMux(),
		caps:   newCapabilityRegistry(),
		health: hs,
		errCh:  make(chan error, 1),
		stopCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.grpcSrv == nil {
		s.grpcSrv = grpc.NewServer()
	}

	contract.RegisterPluginServiceServer(s.grpcSrv, &grpcPluginServer{
		health: hs,
		caps:   s.caps,
	})

	reflection.Register(s.grpcSrv)

	s.mux.HandleFunc("/health", s.health.serveHTTP)
	s.mux.HandleFunc("/capabilities", s.caps.serveHTTP)

	return s, nil
}

func (s *PluginServer) HandleFunc(path string, handler http.HandlerFunc) {
	s.mux.HandleFunc(path, handler)
}

func (s *PluginServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.grpcSrv.RegisterService(desc, impl)
}

func (s *PluginServer) AddCapability(cap *contract.Capability) {
	s.caps.add(cap)
}

func (s *PluginServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("%w: already started", ErrServerStartup)
	}
	s.started = true
	s.mu.Unlock()

	wrapped := corsMiddleware(recoveryMiddleware(loggingMiddleware(s.mux)))

	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.HTTPPort))
	if err != nil {
		return fmt.Errorf("%w: http port %d: %v", ErrServerStartup, s.cfg.HTTPPort, err)
	}

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GRPCPort))
	if err != nil {
		httpLis.Close()
		return fmt.Errorf("%w: grpc port %d: %v", ErrServerStartup, s.cfg.GRPCPort, err)
	}

	s.httpSrv = &http.Server{
		Handler: wrapped,
	}

	go func() {
		if err := s.httpSrv.Serve(httpLis); err != nil && err != http.ErrServerClosed {
			s.errCh <- fmt.Errorf("%w: http serve: %v", ErrServerStartup, err)
		}
	}()

	go func() {
		if err := s.grpcSrv.Serve(grpcLis); err != nil {
			s.errCh <- fmt.Errorf("%w: grpc serve: %v", ErrServerStartup, err)
		}
	}()

	log.Printf("[sdk] plugin %s v%s started (http=%d, grpc=%d)", s.cfg.Name, s.cfg.Version, s.cfg.HTTPPort, s.cfg.GRPCPort)

	descriptor := &contract.PluginDescriptor{
		Name:         s.cfg.Name,
		Version:      s.cfg.Version,
		Description:  s.cfg.Description,
		ApiVersion:   "v1",
		Capabilities: s.caps.list(),
		Endpoints: map[string]string{
			"grpc": fmt.Sprintf(":%d", s.cfg.GRPCPort),
			"http": fmt.Sprintf(":%d", s.cfg.HTTPPort),
		},
	}

	s.reg = NewRegistration(descriptor, s.cfg.OrchestratorURL, s.cfg.RegisterInterval)
	s.reg.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-s.errCh:
		s.Shutdown(context.Background())
		return err
	case <-sigCh:
		log.Printf("[sdk] received signal, shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.Shutdown(shutdownCtx)
	case <-s.stopCh:
		return nil
	}
}

func (s *PluginServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[sdk] shutting down plugin %s...", s.cfg.Name)

	if s.reg != nil {
		s.reg.Stop()
	}

	shutdownDone := make(chan struct{}, 1)
	go func() {
		s.grpcSrv.GracefulStop()
		if s.httpSrv != nil {
			s.httpSrv.Shutdown(context.Background())
		}
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		log.Printf("[sdk] plugin %s stopped", s.cfg.Name)
	case <-ctx.Done():
		log.Printf("[sdk] plugin %s shutdown timed out", s.cfg.Name)
		return ctx.Err()
	}

	return nil
}

type grpcPluginServer struct {
	contract.UnimplementedPluginServiceServer
	health *healthService
	caps   *capabilityRegistry
}

func (g *grpcPluginServer) HealthCheck(ctx context.Context, req *contract.Empty) (*contract.HealthCheckResponse, error) {
	return g.health.checkGRPC()
}

func (g *grpcPluginServer) GetCapabilities(ctx context.Context, req *contract.Empty) (*contract.CapabilitiesResponse, error) {
	caps := g.caps.list()
	return &contract.CapabilitiesResponse{Capabilities: caps}, nil
}
