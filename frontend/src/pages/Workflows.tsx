import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Workflow } from '../api/types'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

export default function Workflows() {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.listWorkflows()
      .then(setWorkflows)
      .catch(() => setError("Impossible de charger les workflows"))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <LoadingSpinner />

  return (
    <div>
      <h1>Workflows</h1>
      {error && <p className="error">{error}</p>}
      {workflows.length === 0 ? (
        <p className="empty">Aucun workflow</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>ID</th><th>Nom</th><th>Statut</th><th>Projet</th><th>D\u00e9marr\u00e9 le</th></tr>
          </thead>
          <tbody>
            {workflows.map(w => (
              <tr key={w.id}>
                <td><Link to={`/workflows/${w.id}`}>{w.id.slice(0, 12)}...</Link></td>
                <td>{w.name}</td>
                <td><StatusBadge status={w.status} /></td>
                <td>{w.project_name || '-'}</td>
                <td className="text-muted">
                  {w.started_at ? new Date(w.started_at).toLocaleString() : '-'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
