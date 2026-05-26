package orchestrator

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "automated_dev_environment/api/grpc"
	"automated_dev_environment/internal/plugins/contract"
	"automated_dev_environment/internal/plugins/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedOrchestratorServiceServer
	store       *registry.Store
	config      *Config
	getSettings func() map[string]string
}

func newGRPCServer(store *registry.Store, cfg *Config, getSettings func() map[string]string) *grpcServer {
	return &grpcServer{
		store:       store,
		config:      cfg,
		getSettings: getSettings,
	}
}

func (s *grpcServer) RegisterPlugin(ctx context.Context, req *pb.RegisterPluginRequest) (*pb.RegisterPluginResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "plugin name is required")
	}

	desc := &contract.PluginDescriptor{
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		ApiVersion:  req.ApiVersion,
		Endpoints:   req.Endpoints,
	}

	for _, cap := range req.Capabilities {
		desc.Capabilities = append(desc.Capabilities, &contract.Capability{
			Name:        cap.Name,
			Version:     cap.Version,
			Description: cap.Description,
		})
	}

	inst := &registry.PluginInstance{
		Descriptor:  desc,
		GrpcAddress: req.GrpcAddress,
		HttpAddress: req.HttpAddress,
		LastSeen:    time.Now(),
	}

	s.store.Register(inst)

	return &pb.RegisterPluginResponse{
		Accepted: true,
		Message:  fmt.Sprintf("plugin %q enregistr\u00e9", req.Name),
	}, nil
}

func (s *grpcServer) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	projectName := os.Getenv("ADE_PROJECT_NAME")
	if projectName == "" {
		projectName = "default"
	}

	var settings map[string]string
	if s.getSettings != nil {
		settings = s.getSettings()
	}

	return &pb.GetConfigResponse{
		ProjectName:         projectName,
		OrchestratorVersion: Version,
		RestPort:            int32(s.config.RESTPort),
		GrpcPort:            int32(s.config.GRPCPort),
		Settings:            settings,
	}, nil
}

func (s *grpcServer) ListPlugins(ctx context.Context, req *pb.ListPluginsRequest) (*pb.ListPluginsResponse, error) {
	plugins := s.store.List()
	infos := make([]*pb.PluginInfo, 0, len(plugins))

	for _, p := range plugins {
		info := &pb.PluginInfo{
			Name:        p.Descriptor.Name,
			Version:     p.Descriptor.Version,
			Status:      p.Status.String(),
			HttpAddress: p.HttpAddress,
			GrpcAddress: p.GrpcAddress,
		}
		for _, cap := range p.Descriptor.Capabilities {
			info.Capabilities = append(info.Capabilities, &pb.Capability{
				Name:        cap.Name,
				Version:     cap.Version,
				Description: cap.Description,
			})
		}
		infos = append(infos, info)
	}

	return &pb.ListPluginsResponse{Plugins: infos}, nil
}

func (s *grpcServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.PluginName == "" {
		return nil, status.Error(codes.InvalidArgument, "plugin name is required")
	}

	if _, ok := s.store.Get(req.PluginName); !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("plugin %q introuvable", req.PluginName))
	}

	s.store.UpdateHealth(req.PluginName, contract.HealthHealthy)

	return &pb.HeartbeatResponse{Accepted: true}, nil
}

func startGRPCServer(store *registry.Store, cfg *Config, getSettings func() map[string]string) (*grpc.Server, net.Listener, error) {
	addr := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("port gRPC %d indisponible: %w", cfg.GRPCPort, err)
	}

	srv := grpc.NewServer()
	impl := newGRPCServer(store, cfg, getSettings)
	pb.RegisterOrchestratorServiceServer(srv, impl)
	reflection.Register(srv)

	go func() {
		log.Printf("[orchestrateur] gRPC d\u00e9marr\u00e9 sur %d", cfg.GRPCPort)
		if err := srv.Serve(lis); err != nil {
			log.Printf("[orchestrateur] gRPC serveur arr\u00eat\u00e9: %v", err)
		}
	}()

	return srv, lis, nil
}
