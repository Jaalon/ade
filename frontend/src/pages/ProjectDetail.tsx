import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Project, Workflow } from '../api/types'
import StatusBadge from '../components/StatusBadge'
import LoadingSpinner from '../components/LoadingSpinner'

export default function ProjectDetail() {
  const { name } = useParams<{ name: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!name) return
    Promise.all([
      api.getProject(name),
      api.listWorkflows(name),
    ])
      .then(([p, w]) => {
        setProject(p)
        setWorkflows(w)
      })
      .catch(() => setError("Projet introuvable"))
      .finally(() => setLoading(false))
  }, [name])

  if (loading) return <LoadingSpinner />
  if (error || !project) return <p className="error">{error || "Projet introuvable"}</p>

  return (
    <div>
      <Link to="/projects" className="back-link">&larr; Retour aux projets</Link>
      <h1>{project.name}</h1>
      <div className="card">
        <p><strong>Description :</strong> {project.description || '-'}</p>
        <p><strong>Créé le :</strong> {new Date(project.created_at).toLocaleString()}</p>
        <p><strong>Mis à jour le :</strong> {new Date(project.updated_at).toLocaleString()}</p>
      </div>

      <h2>Workflows</h2>
      {workflows.length === 0 ? (
        <p className="empty">Aucun workflow pour ce projet</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>ID</th><th>Statut</th><th>Démarré le</th></tr>
          </thead>
          <tbody>
            {workflows.map(w => (
              <tr key={w.id}>
                <td><Link to={`/workflows/${w.id}`}>{w.id.slice(0, 12)}...</Link></td>
                <td><StatusBadge status={w.status} /></td>
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
