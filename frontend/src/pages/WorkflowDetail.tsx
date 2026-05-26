import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Workflow } from '../api/types'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

export default function WorkflowDetail() {
  const { id } = useParams<{ id: string }>()
  const [workflow, setWorkflow] = useState<Workflow | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    api.getWorkflow(id)
      .then(setWorkflow)
      .catch(() => setError("Workflow introuvable"))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) return <LoadingSpinner />
  if (error || !workflow) return <p className="error">{error || "Workflow introuvable"}</p>

  return (
    <div>
      <Link to="/workflows" className="back-link">&larr; Retour aux workflows</Link>
      <h1>{workflow.name}</h1>
      <div className="card">
        <p><strong>ID :</strong> {workflow.id}</p>
        <p><strong>Statut :</strong> <StatusBadge status={workflow.status} /></p>
        <p><strong>Projet :</strong> {workflow.project_name || '-'}</p>
        <p><strong>Démarré le :</strong> {workflow.started_at ? new Date(workflow.started_at).toLocaleString() : '-'}</p>
        <p><strong>Terminé le :</strong> {workflow.completed_at ? new Date(workflow.completed_at).toLocaleString() : '-'}</p>
      </div>

      <h2>Étapes</h2>
      {workflow.steps.length === 0 ? (
        <p className="empty">Aucune étape</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>Étape</th><th>Statut</th><th>Démarré</th><th>Terminé</th></tr>
          </thead>
          <tbody>
            {workflow.steps.map((s, i) => (
              <tr key={i}>
                <td>{s.name}</td>
                <td><StatusBadge status={s.status} /></td>
                <td className="text-muted">{s.started_at ? new Date(s.started_at).toLocaleString() : '-'}</td>
                <td className="text-muted">{s.completed_at ? new Date(s.completed_at).toLocaleString() : '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {workflow.steps.filter(s => s.output).length > 0 && (
        <>
          <h2>Sorties</h2>
          {workflow.steps.filter(s => s.output).map((s, i) => (
            <div key={i} className="card">
              <strong>{s.name}</strong>
              <pre className="output">{s.output}</pre>
            </div>
          ))}
        </>
      )}
    </div>
  )
}
