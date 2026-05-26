interface StatusBadgeProps {
  status: string
}

const colors: Record<string, string> = {
  HEALTHY: 'badge-green',
  DEGRADED: 'badge-yellow',
  UNHEALTHY: 'badge-red',
  UNKNOWN: 'badge-gray',
  healthy: 'badge-green',
  degraded: 'badge-yellow',
  unhealthy: 'badge-red',
  unknown: 'badge-gray',
  completed: 'badge-green',
  running: 'badge-blue',
  pending: 'badge-gray',
  failed: 'badge-red',
  passed: 'badge-green',
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  const cls = colors[status] || 'badge-gray'
  return <span className={`badge ${cls}`}>{status}</span>
}
