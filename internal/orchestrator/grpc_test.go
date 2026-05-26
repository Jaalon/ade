package orchestrator

import (
	"context"
	"net"
	"testing"
	"time"

	pb "automated_dev_environment/api/grpc"
	"automated_dev_environment/internal/plugins/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func setupGRPCTest(t *testing.T) (pb.OrchestratorServiceClient, *registry.Store, func()) {
	t.Helper()

	store := registry.NewStore()
	cfg := DefaultConfig()
	getSettings := func() map[string]string { return nil }

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	impl := newGRPCServer(store, &cfg, getSettings)
	pb.RegisterOrchestratorServiceServer(srv, impl)

	go srv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(2*time.Second))
	require.NoError(t, err)

	client := pb.NewOrchestratorServiceClient(conn)

	cleanup := func() {
		conn.Close()
		srv.Stop()
		lis.Close()
	}

	return client, store, cleanup
}

func TestGRPCRegisterPlugin(t *testing.T) {
	client, store, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.RegisterPlugin(ctx, &pb.RegisterPluginRequest{
		Name:        "test-plugin",
		Version:     "1.0.0",
		HttpAddress: "localhost:8081",
	})

	require.NoError(t, err)
	assert.True(t, resp.Accepted)
	assert.Equal(t, 1, store.Count())

	inst, ok := store.Get("test-plugin")
	require.True(t, ok)
	assert.Equal(t, "test-plugin", inst.Descriptor.Name)
}

func TestGRPCRegisterPlugin_InvalidArgument(t *testing.T) {
	client, _, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.RegisterPlugin(ctx, &pb.RegisterPluginRequest{Name: ""})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCListPlugins_Empty(t *testing.T) {
	client, _, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.ListPlugins(ctx, &pb.ListPluginsRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.Plugins)
}

func TestGRPCListPlugins_WithPlugins(t *testing.T) {
	client, _, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.RegisterPlugin(ctx, &pb.RegisterPluginRequest{
		Name:    "alpha",
		Version: "1.0",
	})
	require.NoError(t, err)

	_, err = client.RegisterPlugin(ctx, &pb.RegisterPluginRequest{
		Name:    "beta",
		Version: "2.0",
	})
	require.NoError(t, err)

	resp, err := client.ListPlugins(ctx, &pb.ListPluginsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Plugins, 2)
}

func TestGRPCGetConfig(t *testing.T) {
	client, _, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	resp, err := client.GetConfig(ctx, &pb.GetConfigRequest{})
	require.NoError(t, err)

	assert.Equal(t, int32(8080), resp.RestPort)
	assert.Equal(t, int32(9090), resp.GrpcPort)
	assert.NotEmpty(t, resp.OrchestratorVersion)
}

func TestGRPCHeartbeat(t *testing.T) {
	client, store, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.RegisterPlugin(ctx, &pb.RegisterPluginRequest{
		Name: "test-plugin",
	})
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	resp, err := client.Heartbeat(ctx, &pb.HeartbeatRequest{PluginName: "test-plugin"})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)

	inst, ok := store.Get("test-plugin")
	require.True(t, ok)
	assert.False(t, inst.LastSeen.IsZero())
}

func TestGRPCHeartbeat_NotFound(t *testing.T) {
	client, _, cleanup := setupGRPCTest(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.Heartbeat(ctx, &pb.HeartbeatRequest{PluginName: "unknown"})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
