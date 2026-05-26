import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

interface ReportSummary {
  id: string
  status: string
}

export default function Reports() {
  const [reports, setReports] = useState<ReportSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.listReports()
      .then(setReports)
      .catch(() => setError("Impossible de charger les rapports"))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <LoadingSpinner />

  return (
    <div>
      <h1>Rapports de validation</h1>
      {error && <p className="error">{error}</p>}
      {reports.length === 0 ? (
        <p className="empty">Aucun rapport disponible</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>ID</th><th>Statut</th></tr>
          </thead>
          <tbody>
            {reports.map(r => (
              <tr key={r.id}>
                <td><Link to={`/reports/${r.id}`}>{r.id}</Link></td>
                <td><StatusBadge status={r.status} /></td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
