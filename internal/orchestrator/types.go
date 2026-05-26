package orchestrator

import "time"

type Project struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Status      string         `json:"status"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Steps       []WorkflowStep `json:"steps,omitempty"`
	ProjectName string         `json:"project_name,omitempty"`
}

type WorkflowStep struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Output      string     `json:"output,omitempty"`
}

type DashboardStats struct {
	Projects      int `json:"projects"`
	Plugins       int `json:"plugins"`
	Workflows     int `json:"workflows"`
	Reports       int `json:"reports"`
	ActivePlugins int `json:"active_plugins"`
}

type ConfigResponse struct {
	ProjectName         string            `json:"project_name"`
	OrchestratorVersion string            `json:"orchestrator_version"`
	RESTPort            int               `json:"rest_port"`
	GRPCPort            int               `json:"grpc_port"`
	Settings            map[string]string `json:"settings,omitempty"`
}

type ActivityEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
