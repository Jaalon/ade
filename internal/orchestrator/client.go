package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"automated_dev_environment/internal/plugins/contract"
	"automated_dev_environment/internal/plugins/registry"
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

var contractErrPluginNotFound = contract.ErrPluginNotFound

func IsNotFound(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "introuvable") || strings.Contains(err.Error(), "not found"))
}
