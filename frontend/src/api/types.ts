export interface Project {
  name: string
  description?: string
  created_at: string
  updated_at: string
  labels?: Record<string, string>
}

export interface Capability {
  name: string
  version: string
  description: string
}

export interface PluginInfo {
  name: string
  version: string
  status: 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY' | 'UNKNOWN'
  http_address: string
  grpc_address: string
  capabilities: Capability[]
}

export interface WorkflowStep {
  name: string
  status: string
  started_at?: string
  completed_at?: string
  output?: string
}

export interface Workflow {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  started_at?: string
  completed_at?: string
  steps: WorkflowStep[]
  project_name?: string
}

export interface ValidationModule {
  name: string
  status: string
  checks: number
  passed: number
  failed: number
  details?: string
}

export interface ValidationReport {
  status: string
  duration: string
  num_modules: number
  num_checks: number
  num_passed: number
  num_failed: number
  modules: ValidationModule[]
  format_version: string
}

export interface DashboardStats {
  projects: number
  plugins: number
  workflows: number
  reports: number
  active_plugins: number
}

export interface ActivityEvent {
  id: string
  type: string
  message: string
  timestamp: string
}

export interface ConfigResponse {
  project_name: string
  orchestrator_version: string
  rest_port: number
  grpc_port: number
  settings?: Record<string, string>
}
