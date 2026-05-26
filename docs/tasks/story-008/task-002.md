# Tâche #002 - Story #008 : API REST de l'orchestrateur (plugins, projets, workflows, configuration, rapports)

## Objectif

Implémenter l'API REST complète de l'orchestrateur en intégrant le registry plugins existant et en ajoutant les nouveaux endpoints pour la gestion des projets, workflows, configuration et rapports de validation.

## Contexte

- Story #008 : [docs/stories/story-008.md](../../stories/story-008.md)
- Dépend de : Tâche #001 (structure du serveur)
- Nécessaire pour : Tâche #004, Tâche #005, Tâche #006

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Étendre le serveur orchestrateur pour exposer une API REST complète sur le port 8080.

**Endpoints à implémenter :**

#### Plugins (délègue au registry existant)
- `GET /api/v1/plugins` — Liste des plugins enregistrés
- `GET /api/v1/plugins/{name}` — Détail d'un plugin
- `DELETE /api/v1/plugins/{name}` — Désenregistrer un plugin
- `POST /api/v1/plugins/register` — Enregistrement push d'un plugin
- `GET /api/v1/plugins/{name}/health` — Healthcheck spécifique d'un plugin

#### Projets (nouveau)
- `POST /api/v1/projects` — Créer un projet
- `GET /api/v1/projects` — Lister les projets
- `GET /api/v1/projects/{name}` — Détail d'un projet
- `PUT /api/v1/projects/{name}` — Modifier un projet
- `DELETE /api/v1/projects/{name}` — Supprimer un projet

#### Workflows (nouveau)
- `GET /api/v1/workflows` — Lister les workflows et leur historique
- `GET /api/v1/workflows/{id}` — Détail d'une exécution de workflow
- `POST /api/v1/workflows` — Déclencher un workflow

#### Configuration (nouveau)
- `GET /api/v1/config` — Récupérer la configuration globale de l'orchestrateur
- `PUT /api/v1/config` — Mettre à jour la configuration

#### Rapports de validation (nouveau)
- `GET /api/v1/reports` — Lister les rapports de validation disponibles
- `GET /api/v1/reports/{id}` — Détail d'un rapport

**Cas nominaux :**
- CRUD complet des projets via REST
- Délégation des endpoints plugins au registry existant (via `registry.API`)
- Récupération de la configuration globale de l'orchestrateur
- Consultation des workflows et de leur historique

**Cas limites :**
- Liste vide → `[]` ou objet vide selon le contexte, pas de null
- Projet déjà existant → 409 Conflict
- Plugin introuvable → 404 (déjà géré par registry)

**Gestion d'erreurs :**
- Erreur 400 → corps JSON `{"error": "message"}`
- Erreur 404 → corps JSON `{"error": "ressource introuvable"}`
- Erreur 409 → corps JSON `{"error": "conflit: ..."}`
- Erreur 500 → corps JSON `{"error": "erreur interne"}`
- Middleware CORS pour permettre les requêtes depuis le frontend web

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/orchestrator/router.go` | Créer | Routeur HTTP principal de l'orchestrateur |
| `internal/orchestrator/projects.go` | Créer | Handlers REST pour les projets |
| `internal/orchestrator/workflows.go` | Créer | Handlers REST pour les workflows |
| `internal/orchestrator/config.go` | Modifier | Ajouter les handlers pour la configuration |
| `internal/orchestrator/reports.go` | Créer | Handlers REST pour les rapports de validation |
| `internal/orchestrator/middleware.go` | Créer | Middleware CORS, logging, recovery |
| `internal/orchestrator/types.go` | Créer | Types de données pour projets, workflows |

### Signatures

```go
// internal/orchestrator/router.go
package orchestrator

func (s *Server) setupRoutes()

// internal/orchestrator/projects.go
type Project struct {
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Labels      map[string]string `json:"labels,omitempty"`
}

type ProjectStore struct {
    mu       sync.RWMutex
    projects map[string]*Project
}

func NewProjectStore() *ProjectStore
func (s *ProjectStore) Create(p *Project) error
func (s *ProjectStore) Get(name string) (*Project, bool)
func (s *ProjectStore) List() []*Project
func (s *ProjectStore) Update(p *Project) error
func (s *ProjectStore) Delete(name string) error

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request)
func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request)
func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request)
func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request)

// internal/orchestrator/workflows.go
type Workflow struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    Status      string          `json:"status"` // pending, running, completed, failed
    StartedAt   *time.Time      `json:"started_at,omitempty"`
    CompletedAt *time.Time      `json:"completed_at,omitempty"`
    Steps       []WorkflowStep  `json:"steps,omitempty"`
    ProjectName string          `json:"project_name,omitempty"`
}

type WorkflowStep struct {
    Name        string     `json:"name"`
    Status      string     `json:"status"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    Output      string     `json:"output,omitempty"`
}

type WorkflowStore struct {
    mu        sync.RWMutex
    workflows map[string]*Workflow
}

func NewWorkflowStore() *WorkflowStore
func (s *WorkflowStore) Create(w *Workflow) error
func (s *WorkflowStore) Get(id string) (*Workflow, bool)
func (s *WorkflowStore) List() []*Workflow
func (s *WorkflowStore) ListByProject(projectName string) []*Workflow

// internal/orchestrator/reports.go
func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request)
```

### Contraintes techniques

- **Framework** : Go 1.26 `net/http` standard, pas de framework externe
- **Pattern** : Router central avec `http.ServeMux` (ou gorilla/mux si déjà présent — vérifier go.mod)
- **Intégration** : Réutiliser les handlers existants de `internal/plugins/registry/api.go` (déjà un `ServeHTTP` router)
- **CORS** : Middleware CORS permettant toutes les origines en développement
- **Format réponses** : Toujours du JSON avec `Content-Type: application/json`
- **Persistance** : Les stores sont en mémoire (pas de base de données)
- **Messages d'erreur** : En français (convention du projet)

### Tests à implémenter

#### Tests unitaires

- **Fichier** : `internal/orchestrator/projects_test.go`
- Scénario 1 : Création d'un projet
  - Données : POST `/api/v1/projects` avec `{"name":"mon-projet","description":"test"}`
  - Résultat attendu : 201 Created, body contient le projet créé
- Scénario 2 : Création d'un projet déjà existant
  - Données : Même nom deux fois
  - Résultat attendu : 409 Conflict
- Scénario 3 : Liste des projets (vide et non-vide)
  - Données : Aucun projet → `[]` ; après création → 1 élément
- Scénario 4 : Suppression d'un projet inexistant
  - Données : DELETE `/api/v1/projects/inconnu`
  - Résultat attendu : 404 Not Found
- Scénario 5 : Mise à jour d'un projet
  - Données : PUT avec nouveau `description`
  - Résultat attendu : 200 OK, projet mis à jour

- **Fichier** : `internal/orchestrator/workflows_test.go`
- Scénario 1 : Création et liste de workflows
- Scénario 2 : Workflow avec steps
- Scénario 3 : Liste par projet

- **Fichier** : `internal/orchestrator/router_test.go`
- Scénario 1 : Route inconnue → 404
- Scénario 2 : Méthode non autorisée → 405
- Scénario 3 : Header CORS présent dans les réponses

#### Tests d'intégration
- **Fichier** : `internal/orchestrator/integration_test.go`
- Scénario : Serveur complet démarré, enchaînement créer projet → lister → supprimer → vérifier

### Documentation

#### Documentation à créer
- `docs/orchestrator/api.md` — API REST de l'orchestrateur avec exemples de requêtes/réponses

#### Exemples d'API HTTP
- Fichier : `docs/tasks/story-008/task-002-examples.http`

```http
###
# Créer un projet
POST http://localhost:8080/api/v1/projects
Content-Type: application/json

{"name": "mon-app", "description": "Application de test"}

###
# Lister les projets
GET http://localhost:8080/api/v1/projects

###
# Récupérer un projet
GET http://localhost:8080/api/v1/projects/mon-app

###
# Modifier un projet
PUT http://localhost:8080/api/v1/projects/mon-app
Content-Type: application/json

{"description": "Description mise à jour"}

###
# Supprimer un projet
DELETE http://localhost:8080/api/v1/projects/mon-app

###
# Lister les plugins
GET http://localhost:8080/api/v1/plugins

###
# Récupérer la configuration
GET http://localhost:8080/api/v1/config

###
# Lister les workflows
GET http://localhost:8080/api/v1/workflows

###
# Lister les rapports
GET http://localhost:8080/api/v1/reports
```
