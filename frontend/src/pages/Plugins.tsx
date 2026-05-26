import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { PluginInfo } from '../api/types'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

export default function Plugins() {
  const [plugins, setPlugins] = useState<PluginInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.listPlugins()
      .then(setPlugins)
      .catch(() => setError("Impossible de charger les plugins"))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <LoadingSpinner />

  return (
    <div>
      <h1>Plugins</h1>
      {error && <p className="error">{error}</p>}
      {plugins.length === 0 ? (
        <p className="empty">Aucun plugin enregistr\u00e9</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>Nom</th><th>Version</th><th>Statut</th><th>Adresse HTTP</th><th>Capacit\u00e9s</th></tr>
          </thead>
          <tbody>
            {plugins.map(p => (
              <tr key={p.name}>
                <td><Link to={`/plugins/${p.name}`}>{p.name}</Link></td>
                <td>{p.version || '-'}</td>
                <td><StatusBadge status={p.status} /></td>
                <td className="text-muted">{p.http_address || '-'}</td>
                <td>{p.capabilities.map(c => c.name).join(', ') || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
