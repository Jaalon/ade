package contract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPluginDescriptor_JSON(t *testing.T) {
	d := &PluginDescriptor{
		Name:        "templates",
		Version:     "1.0.0",
		Description: "Fournit des templates de projets",
		ApiVersion:  "v1",
		Capabilities: []*Capability{
			{Name: "template_provider", Description: "Provides templates", Version: "1.0.0"},
		},
		Endpoints: map[string]string{"templates": "/api/v1/templates"},
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got PluginDescriptor
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.Name != d.Name {
		t.Errorf("Name = %q, want %q", got.Name, d.Name)
	}
	if got.Version != d.Version {
		t.Errorf("Version = %q, want %q", got.Version, d.Version)
	}
	if got.ApiVersion != d.ApiVersion {
		t.Errorf("ApiVersion = %q, want %q", got.ApiVersion, d.ApiVersion)
	}
	if len(got.Capabilities) != 1 {
		t.Fatalf("len(Capabilities) = %d, want 1", len(got.Capabilities))
	}
	if got.Capabilities[0].Name != "template_provider" {
		t.Errorf("Capability[0].Name = %q, want %q", got.Capabilities[0].Name, "template_provider")
	}
	if got.Endpoints["templates"] != "/api/v1/templates" {
		t.Errorf("Endpoints[templates] = %q, want %q", got.Endpoints["templates"], "/api/v1/templates")
	}
}

func TestPluginDescriptor_EmptyCapabilities(t *testing.T) {
	d := &PluginDescriptor{
		Name:    "utility",
		Version: "0.1.0",
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got PluginDescriptor
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(got.Capabilities) != 0 {
		t.Errorf("len(Capabilities) = %d, want 0", len(got.Capabilities))
	}
}

func TestHealthStatus_AllEnums(t *testing.T) {
	tests := []struct {
		status HealthStatusEnum
		want   string
	}{
		{HealthUnknown, "UNKNOWN"},
		{HealthHealthy, "HEALTHY"},
		{HealthDegraded, "DEGRADED"},
		{HealthUnhealthy, "UNHEALTHY"},
	}

	for _, tt := range tests {
		h := HealthStatus{Status: tt.status, Timestamp: 1000}
		data, err := json.Marshal(h)
		if err != nil {
			t.Fatalf("json.Marshal(%d) failed: %v", tt.status, err)
		}

		var got HealthStatus
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("json.Unmarshal(%s) failed: %v", string(data), err)
		}

		if got.Status != tt.status {
			t.Errorf("Status = %d, want %d", got.Status, tt.status)
		}
		if got.Status.String() != tt.want {
			t.Errorf("Status.String() = %q, want %q", got.Status.String(), tt.want)
		}
	}
}

func TestHealthStatusFromProto(t *testing.T) {
	tests := []struct {
		proto HealthCheckResponse_Status
		want  HealthStatusEnum
	}{
		{HealthCheckResponse_HEALTHY, HealthHealthy},
		{HealthCheckResponse_DEGRADED, HealthDegraded},
		{HealthCheckResponse_UNHEALTHY, HealthUnhealthy},
		{HealthCheckResponse_UNKNOWN, HealthUnknown},
		{HealthCheckResponse_Status(99), HealthUnknown},
	}

	for _, tt := range tests {
		got := HealthStatusFromProto(tt.proto)
		if got != tt.want {
			t.Errorf("HealthStatusFromProto(%d) = %d, want %d", tt.proto, got, tt.want)
		}
	}
}

func TestCapability_EmptyConfigSchemaOmitted(t *testing.T) {
	c := &Capability{
		Name:        "test",
		Description: "test capability",
		Version:     "1.0.0",
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	if strings.Contains(string(data), "config_schema") {
		t.Error("config_schema field should be omitted when empty, got:", string(data))
	}
}

func TestCapability_WithConfigSchema(t *testing.T) {
	c := &Capability{
		Name:         "configurable",
		Description:  "Has config schema",
		Version:      "1.0.0",
		ConfigSchema: `{"type": "object", "properties": {}}`,
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got Capability
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.ConfigSchema != c.ConfigSchema {
		t.Errorf("ConfigSchema = %q, want %q", got.ConfigSchema, c.ConfigSchema)
	}
}

func TestRegisterRequest_JSON(t *testing.T) {
	req := &RegisterRequest{
		Name:        "templates",
		Version:     "1.0.0",
		ApiVersion:  "v1",
		Description: "Template provider",
		GrpcAddress: "localhost:50051",
		HttpAddress: "localhost:8081",
		Capabilities: []*Capability{
			{Name: "template_provider", Description: "Provides templates", Version: "1.0.0"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got RegisterRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.Name != req.Name {
		t.Errorf("Name = %q, want %q", got.Name, req.Name)
	}
	if len(got.Capabilities) != 1 {
		t.Errorf("len(Capabilities) = %d, want 1", len(got.Capabilities))
	}
}

func TestRegisterResponse_Accepted(t *testing.T) {
	resp := &RegisterResponse{
		Accepted: true,
		Message:  "plugin registered successfully",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got RegisterResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.Accepted != true {
		t.Errorf("Accepted = %v, want true", got.Accepted)
	}
	if got.Message != resp.Message {
		t.Errorf("Message = %q, want %q", got.Message, resp.Message)
	}
}

func TestHealthCheckResponse_JSON(t *testing.T) {
	h := &HealthCheckResponse{
		Status:    HealthCheckResponse_HEALTHY,
		Message:   "all good",
		Timestamp: 1700000000,
	}

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got HealthCheckResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.Status != HealthCheckResponse_HEALTHY {
		t.Errorf("Status = %v, want HEALTHY", got.Status)
	}
	if got.Message != "all good" {
		t.Errorf("Message = %q, want %q", got.Message, "all good")
	}
}

func TestPluginDescriptor_EmptyEndpointsNil(t *testing.T) {
	d := &PluginDescriptor{
		Name:    "no-endpoints",
		Version: "0.1.0",
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := raw["endpoints"]; ok {
		t.Error("endpoints field should be omitted when nil")
	}
}
