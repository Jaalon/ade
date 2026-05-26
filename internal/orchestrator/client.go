package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	pb "automated_dev_environment/api/grpc"
	"automated_dev_environment/internal/plugins/contract"
	"automated_dev_environment/internal/plugins/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type orchestratorConfig struct {
	restURL string
	grpcURL string
}

func loadOrchestratorConfig() orchestratorConfig {
	cfg := orchestratorConfig{
		restURL: "http://localhost:8080",
		grpcURL: "grpc://localhost:9090",
	}

	if envURL := os.Getenv("ADE_ORCHESTRATOR_URL"); envURL != "" {
		cfg.restURL = envURL
		host := strings.TrimPrefix(envURL, "http://")
		host = strings.TrimPrefix(host, "https://")
		cfg.grpcURL = fmt.Sprintf("grpc://%s:9090", strings.Split(host, ":")[0])
	}

	return cfg
}

type Client struct {
	restURL    string
	grpcURL    string
	httpClient *http.Client
	timeout    time.Duration
}

func NewClient() *Client {
	cfg := loadOrchestratorConfig()
	return &Client{
		restURL: cfg.restURL,
		grpcURL: cfg.grpcURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		timeout: 5 * time.Second,
	}
}

func NewClientWithURL(restURL string) *Client {
	return &Client{
		restURL: restURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		timeout: 5 * time.Second,
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.restURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("création requête: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("orchestrateur non disponible: %w", err)
	}
	return resp, nil
}

func (c *Client) GetConfig(ctx context.Context) (*ConfigResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/config", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orchestrateur: %s", string(body))
	}

	var cfg ConfigResponse
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("erreur de communication avec l'orchestrateur: %w", err)
	}
	return &cfg, nil
}

func (c *Client) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/dashboard/stats", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orchestrateur: %s", string(body))
	}

	var stats DashboardStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("erreur de communication avec l'orchestrateur: %w", err)
	}
	return &stats, nil
}

func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/projects", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orchestrateur: %s", string(body))
	}

	var result struct {
		Projects []Project `json:"projects"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erreur de communication avec l'orchestrateur: %w", err)
	}
	return result.Projects, nil
}

func (c *Client) CreateProject(ctx context.Context, p Project) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(p); err != nil {
		return fmt.Errorf("encodage projet: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/projects", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("lecture réponse: %w", err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("orchestrateur: %s", string(body))
	}
	return nil
}

func (c *Client) DeleteProject(ctx context.Context, name string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/api/v1/projects/"+name, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("projet %q introuvable", name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("lecture réponse: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("orchestrateur: %s", string(body))
	}
	return nil
}

func (c *Client) ListPlugins(ctx context.Context) ([]*registry.PluginInstance, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.restURL+"/api/v1/plugins", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("orchestrateur non disponible: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orchestrateur: %s", string(body))
	}

	var result struct {
		Plugins []*registry.PluginInstance `json:"plugins"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erreur de communication avec l'orchestrateur: %w", err)
	}

	return result.Plugins, nil
}

func (c *Client) GetPlugin(ctx context.Context, name string) (*registry.PluginInstance, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.restURL+"/api/v1/plugins/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("orchestrateur non disponible: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("plugin %q introuvable", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orchestrateur: %s", string(body))
	}

	var plugin registry.PluginInstance
	if err := json.Unmarshal(body, &plugin); err != nil {
		return nil, fmt.Errorf("erreur de communication avec l'orchestrateur: %w", err)
	}

	return &plugin, nil
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.restURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("orchestrateur non disponible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("orchestrateur: status %d", resp.StatusCode)
	}

	return nil
}

func FormatPluginList(plugins []*registry.PluginInstance) string {
	if len(plugins) == 0 {
		return "Aucun plugin enregistré."
	}

	var b strings.Builder
	b.WriteString("NOM\tVERSION\tSTATUT\tCAPACITÉS\tADRESSE HTTP\n")

	for _, p := range plugins {
		caps := ""
		if p.Descriptor != nil {
			for i, cap := range p.Descriptor.Capabilities {
				if i > 0 {
					caps += ", "
				}
				caps += cap.Name
			}
		}
		addr := p.HttpAddress
		if addr == "" {
			addr = "-"
		}
		version := ""
		if p.Descriptor != nil {
			version = p.Descriptor.Version
		}
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
			p.Descriptor.Name, version, p.Status.String(), caps, addr))
	}

	return b.String()
}

func FormatPluginInfo(p *registry.PluginInstance) string {
	if p == nil {
		return ""
	}

	var b strings.Builder
	name := ""
	version := ""
	description := ""
	apiVersion := ""
	if p.Descriptor != nil {
		name = p.Descriptor.Name
		version = p.Descriptor.Version
		description = p.Descriptor.Description
		apiVersion = p.Descriptor.ApiVersion
	}

	b.WriteString(fmt.Sprintf("Nom         : %s\n", name))
	b.WriteString(fmt.Sprintf("Version     : %s\n", version))
	b.WriteString(fmt.Sprintf("Statut      : %s\n", p.Status.String()))
	b.WriteString(fmt.Sprintf("Description : %s\n", description))
	b.WriteString(fmt.Sprintf("API Version : %s\n", apiVersion))
	b.WriteString(fmt.Sprintf("HTTP        : %s\n", p.HttpAddress))
	b.WriteString(fmt.Sprintf("gRPC        : %s\n", p.GrpcAddress))
	b.WriteString("\nCapacités:\n")

	if p.Descriptor != nil {
		for _, cap := range p.Descriptor.Capabilities {
			b.WriteString(fmt.Sprintf("  %s %s - %s\n", cap.Name, cap.Version, cap.Description))
		}
	}

	b.WriteString("\nEndpoints:\n")
	if p.Descriptor != nil {
		for key, endpoint := range p.Descriptor.Endpoints {
			b.WriteString(fmt.Sprintf("  %s : %s\n", key, endpoint))
		}
	}

	return b.String()
}

func (c *Client) grpcAddr() string {
	addr := strings.TrimPrefix(c.grpcURL, "grpc://")
	return addr
}

func (c *Client) dialGRPC(ctx context.Context) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return grpc.DialContext(dialCtx, c.grpcAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
}

func (c *Client) RegisterPluginGRPC(ctx context.Context, name, version, httpAddr, grpcAddr string) (*pb.RegisterPluginResponse, error) {
	conn, err := c.dialGRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("connexion gRPC: %w", err)
	}
	defer conn.Close()

	client := pb.NewOrchestratorServiceClient(conn)
	return client.RegisterPlugin(ctx, &pb.RegisterPluginRequest{
		Name:        name,
		Version:     version,
		HttpAddress: httpAddr,
		GrpcAddress: grpcAddr,
	})
}

func (c *Client) ListPluginsGRPC(ctx context.Context) (*pb.ListPluginsResponse, error) {
	conn, err := c.dialGRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("connexion gRPC: %w", err)
	}
	defer conn.Close()

	client := pb.NewOrchestratorServiceClient(conn)
	return client.ListPlugins(ctx, &pb.ListPluginsRequest{})
}

func (c *Client) GetConfigGRPC(ctx context.Context) (*pb.GetConfigResponse, error) {
	conn, err := c.dialGRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("connexion gRPC: %w", err)
	}
	defer conn.Close()

	client := pb.NewOrchestratorServiceClient(conn)
	return client.GetConfig(ctx, &pb.GetConfigRequest{})
}

var contractErrPluginNotFound = contract.ErrPluginNotFound

func IsNotFound(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "introuvable") || strings.Contains(err.Error(), "not found"))
}
