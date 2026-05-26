import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '../api/client'
import type { PluginInfo } from '../api/types'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

export default function PluginDetail() {
  const { name } = useParams<{ name: string }>()
  const [plugin, setPlugin] = useState<PluginInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!name) return
    api.getPlugin(name)
      .then(setPlugin)
      .catch(() => setError("Plugin introuvable"))
      .finally(() => setLoading(false))
  }, [name])

  if (loading) return <LoadingSpinner />
  if (error || !plugin) return <p className="error">{error || "Plugin introuvable"}</p>

  return (
    <div>
      <Link to="/plugins" className="back-link">&larr; Retour aux plugins</Link>
      <h1>{plugin.name}</h1>
      <div className="card">
        <p><strong>Version :</strong> {plugin.version || '-'}</p>
        <p><strong>Statut :</strong> <StatusBadge status={plugin.status} /></p>
        <p><strong>Adresse HTTP :</strong> {plugin.http_address || '-'}</p>
        <p><strong>Adresse gRPC :</strong> {plugin.grpc_address || '-'}</p>
      </div>

      <h2>Capacités</h2>
      {plugin.capabilities.length === 0 ? (
        <p className="empty">Aucune capacité déclarée</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>Nom</th><th>Version</th><th>Description</th></tr>
          </thead>
          <tbody>
            {plugin.capabilities.map((c, i) => (
              <tr key={i}>
                <td>{c.name}</td>
                <td>{c.version || '-'}</td>
                <td>{c.description || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
