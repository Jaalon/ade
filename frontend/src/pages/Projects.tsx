import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Project } from '../api/types'
import LoadingSpinner from '../components/LoadingSpinner'

export default function Projects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const navigate = useNavigate()

  const load = () => {
    setLoading(true)
    api.listProjects()
      .then(setProjects)
      .catch(() => setError("Impossible de charger les projets"))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleCreate = async () => {
    if (!newName.trim()) return
    try {
      await api.createProject({ name: newName.trim(), description: newDesc.trim() })
      setShowModal(false)
      setNewName('')
      setNewDesc('')
      load()
    } catch {
      setError("Erreur lors de la création")
    }
  }

  const handleDelete = async (name: string) => {
    if (!confirm(`Supprimer le projet "${name}" ?`)) return
    try {
      await api.deleteProject(name)
      load()
    } catch {
      setError("Erreur lors de la suppression")
    }
  }

  if (loading) return <LoadingSpinner />

  return (
    <div>
      <div className="page-header">
        <h1>Projets</h1>
        <button className="btn" onClick={() => setShowModal(true)}>Nouveau projet</button>
      </div>

      {error && <p className="error">{error}</p>}

      {projects.length === 0 ? (
        <p className="empty">Aucun projet</p>
      ) : (
        <table className="table">
          <thead>
            <tr><th>Nom</th><th>Description</th><th>Créé le</th><th>Actions</th></tr>
          </thead>
          <tbody>
            {projects.map(p => (
              <tr key={p.name}>
                <td><Link to={`/projects/${p.name}`}>{p.name}</Link></td>
                <td className="text-muted">{p.description || '-'}</td>
                <td className="text-muted">{new Date(p.created_at).toLocaleDateString()}</td>
                <td>
                  <button className="btn btn-sm" onClick={() => navigate(`/projects/${p.name}`)}>
                    Voir
                  </button>
                  <button className="btn btn-sm btn-danger" onClick={() => handleDelete(p.name)}>
                    Suppr.
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <h2>Nouveau projet</h2>
            <input
              placeholder="Nom du projet"
              value={newName}
              onChange={e => setNewName(e.target.value)}
              autoFocus
            />
            <input
              placeholder="Description (optionnelle)"
              value={newDesc}
              onChange={e => setNewDesc(e.target.value)}
            />
            <div className="modal-actions">
              <button className="btn" onClick={() => setShowModal(false)}>Annuler</button>
              <button className="btn btn-primary" onClick={handleCreate}>Créer</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
