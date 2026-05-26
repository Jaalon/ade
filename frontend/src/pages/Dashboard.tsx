import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { DashboardStats, ActivityEvent } from '../api/types'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [events, setEvents] = useState<ActivityEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      api.getDashboardStats(),
      api.getDashboardActivity(),
    ])
      .then(([s, e]) => {
        setStats(s)
        setEvents(e)
      })
      .catch(() => setError("Orchestrateur non joignable"))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <LoadingSpinner />
  if (error) return <div className="card"><p className="error">{error}</p></div>

  const cards = [
    { label: 'Projets', value: stats?.projects ?? 0 },
    { label: 'Plugins actifs', value: `${stats?.active_plugins ?? 0}/${stats?.plugins ?? 0}` },
    { label: 'Workflows', value: stats?.workflows ?? 0 },
    { label: 'Rapports', value: stats?.reports ?? 0 },
  ]

  return (
    <div>
      <h1>Tableau de bord</h1>
      <div className="stats-grid">
        {cards.map(c => (
          <div key={c.label} className="stat-card">
            <span className="stat-value">{c.value}</span>
            <span className="stat-label">{c.label}</span>
          </div>
        ))}
      </div>
      <h2>Activité récente</h2>
      {events.length === 0 ? (
        <p className="empty">Aucune activité</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>Type</th><th>Message</th><th>Date</th></tr>
          </thead>
          <tbody>
            {events.map(e => (
              <tr key={e.id}>
                <td><StatusBadge status={e.type} /></td>
                <td>{e.message}</td>
                <td className="text-muted">{new Date(e.timestamp).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
