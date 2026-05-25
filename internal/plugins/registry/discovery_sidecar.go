package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ContainerInfo struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	State  string            `json:"state"`
}

type SidecarClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSidecarClient(baseURL string) *SidecarClient {
	return &SidecarClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *SidecarClient) ListContainers(ctx context.Context, labelFilter string) ([]ContainerInfo, error) {
	url := c.baseURL + "/containers"
	if labelFilter != "" {
		url += "?label=" + labelFilter
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar returned status %d", resp.StatusCode)
	}

	var result struct {
		Containers []ContainerInfo `json:"containers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Containers, nil
}
