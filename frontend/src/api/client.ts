import type {
  Project, PluginInfo, Workflow, ValidationReport,
  DashboardStats, ActivityEvent, ConfigResponse,
} from './types'

const BASE = ''

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new ApiError(res.status, body || res.statusText)
  }
  return res.json()
}

export const api = {
  health: () => request<{ status: string; version: string }>('/health'),

  getConfig: () => request<ConfigResponse>('/api/v1/config'),
  updateConfig: (settings: Record<string, string>) =>
    request<{ message: string }>('/api/v1/config', {
      method: 'PUT',
      body: JSON.stringify({ settings }),
    }),

  listProjects: () =>
    request<{ projects: Project[] }>('/api/v1/projects').then(r => r.projects),
  getProject: (name: string) => request<Project>(`/api/v1/projects/${encodeURIComponent(name)}`),
  createProject: (project: { name: string; description?: string }) =>
    request<Project>('/api/v1/projects', {
      method: 'POST',
      body: JSON.stringify(project),
    }),
  updateProject: (name: string, data: Partial<Project>) =>
    request<Project>(`/api/v1/projects/${encodeURIComponent(name)}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteProject: (name: string) =>
    request<{ message: string }>(`/api/v1/projects/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    }),

  listPlugins: () =>
    request<{ plugins: PluginInfo[] }>('/api/v1/plugins').then(r => r.plugins),
  getPlugin: (name: string) => request<PluginInfo>(`/api/v1/plugins/${encodeURIComponent(name)}`),

  listWorkflows: (project?: string) => {
    const qs = project ? `?project=${encodeURIComponent(project)}` : ''
    return request<{ workflows: Workflow[] }>(`/api/v1/workflows${qs}`).then(r => r.workflows)
  },
  getWorkflow: (id: string) => request<Workflow>(`/api/v1/workflows/${encodeURIComponent(id)}`),

  listReports: () =>
    request<{ reports: { id: string; status: string }[] }>('/api/v1/reports').then(r => r.reports),
  getReport: (id: string) => request<ValidationReport>(`/api/v1/reports/${encodeURIComponent(id)}`),

  getDashboardStats: () => request<DashboardStats>('/api/v1/dashboard/stats'),
  getDashboardActivity: () =>
    request<{ events: ActivityEvent[] }>('/api/v1/dashboard/activity').then(r => r.events),
}
