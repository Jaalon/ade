import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '../api/client'
import type { ValidationReport } from '../api/types'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

export default function ReportDetail() {
  const { id } = useParams<{ id: string }>()
  const [report, setReport] = useState<ValidationReport | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    api.getReport(id)
      .then(setReport)
      .catch(() => setError("Rapport introuvable"))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <LoadingSpinner />
  if (error || !report) return <p className="error">{error || "Rapport introuvable"}</p>

  return (
    <div>
      <Link to="/reports" className="back-link">&larr; Retour aux rapports</Link>
      <h1>Rapport {id}</h1>
      <div className="stats-grid">
        <div className="stat-card">
          <span className="stat-value">{report.num_modules}</span>
          <span className="stat-label">Modules</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{report.num_checks}</span>
          <span className="stat-label">Checks</span>
        </div>
        <div className="stat-card stat-passed">
          <span className="stat-value">{report.num_passed}</span>
          <span className="stat-label">Passés</span>
        </div>
        <div className="stat-card stat-failed">
          <span className="stat-value">{report.num_failed}</span>
          <span className="stat-label">Échoués</span>
        </div>
      </div>

      <h2>Modules</h2>
      {report.modules.length === 0 ? (
        <p className="empty">Aucun module</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>Module</th><th>Statut</th><th>Checks</th><th>Passés</th><th>Échoués</th></tr>
          </thead>
          <tbody>
            {report.modules.map((m, i) => (
              <tr key={i}>
                <td>{m.name}</td>
                <td><StatusBadge status={m.status} /></td>
                <td>{m.checks}</td>
                <td>{m.passed}</td>
                <td>{m.failed}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
