# Tâche #004 - Story #008 : Backend web UI (serveur de fichiers statiques et API dédiée)

## Objectif

Ajouter au serveur orchestrateur le service de fichiers statiques pour l'interface web et les endpoints API spécifiques au frontend (dashboard, métriques, WebSocket pour live updates).

## Contexte

- Story #008 : [docs/stories/story-008.md](../../stories/story-008.md)
- Dépend de : Tâche #002 (API REST existante)
- Nécessaire pour : Tâche #005 (frontend web)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter au serveur orchestrateur (`internal/orchestrator/server.go`) le service de fichiers statiques pour servir le frontend web (SPA React/Vue) et des endpoints API supplémentaires dédiés au tableau de bord.

**Fonctionnalités :**

1. **Serveur de fichiers statiques** : Servir les fichiers du frontend (HTML, JS, CSS) depuis un répertoire configurable ou depuis `embed.FS`
2. **SPA fallback** : Toutes les routes inconnues servent `index.html` (nécessaire pour le routing côté client du frontend)
3. **Endpoints dashboard** :
   - `GET /api/v1/dashboard/stats` — Statistiques globales (nb projets, plugins, workflows, rapports)
   - `GET /api/v1/dashboard/activity` — Activité récente (derniers événements)
4. **WebSocket (optionnel)** : Endpoint `/api/v1/ws` pour les notifications push (plugins up/down, workflow terminé)

**Cas nominaux :**
- Requête `GET /` → sert `index.html`
- Requête `GET /static/js/app.js` → sert le fichier JS correspondant
- Requête `GET /une-route-frontend` → sert `index.html` (SPA fallback)
- `GET /api/v1/dashboard/stats` → JSON avec les compteurs

**Cas limites :**
- Fichier statique inexistant → SPA fallback (pas 404), sauf pour `/api/*`
- Aucun frontend buildé → message clair dans les logs ("frontend not found, serving API only")
- Dashboard stats avec zéro projet/plugin → compteurs à 0, pas d'erreur

**Gestion d'erreurs :**
- Embed FS non disponible → fallback sur un système de fichiers local
- WebSocket non supporté → simplement pas d'erreur, la fonctionnalité est optionnelle

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/orchestrator/server.go` | Modifier | Ajouter le serveur de fichiers statiques et SPA fallback |
| `internal/orchestrator/dashboard.go` | Créer | Handlers REST pour le dashboard |
| `internal/orchestrator/websocket.go` | Créer | Gestion WebSocket optionnelle |
| `internal/orchestrator/embed.go` | Créer | Embed du frontend buildé (embed.FS) |
| `internal/orchestrator/static_dev.go` | Créer | Fallback fichiers statiques depuis disque (dev mode) |

### Signatures

```go
// internal/orchestrator/embed.go
package orchestrator

import "embed"

//go:embed frontend/dist/*
var FrontendFS embed.FS

// internal/orchestrator/static_dev.go
package orchestrator

// DevModeFS sert les fichiers statiques depuis le disque (mode développement)
// quand l'embed n'est pas disponible ou que ADE_DEV_MODE=true
type DevModeFS struct {
    root string
}

func NewDevModeFS(root string) *DevModeFS

// internal/orchestrator/dashboard.go
type DashboardStats struct {
    Projects      int `json:"projects"`
    Plugins       int `json:"plugins"`
    Workflows     int `json:"workflows"`
    Reports       int `json:"reports"`
    ActivePlugins int `json:"active_plugins"`
}

type ActivityEvent struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"` // plugin_register, plugin_unregister, workflow_start, workflow_complete, project_create, etc.
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}

func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request)
func (s *Server) handleDashboardActivity(w http.ResponseWriter, r *http.Request)

// internal/orchestrator/websocket.go
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request)
```

### Contraintes techniques

- **Embed** : Utiliser `//go:embed` pour embarquer le frontend buildé dans le binaire Go
- **SPA Fallback** : Intercepter les requêtes qui ne sont pas `/api/*` et ne sont pas des fichiers statiques existants → servir `index.html`
- **Ordre des routes** : `/api/` d'abord, puis fichiers statiques, puis SPA fallback
- **Dev mode** : Variable d'env `ADE_DEV_MODE=true` → charger les fichiers depuis le système de fichiers local (hot reload)
- **WebSocket** : Optionnel, utiliser `gorilla/websocket` si disponible dans go.mod (sinon, le rendre non-bloquant)
- **Activité** : Store d'activité en mémoire avec une limite de 100 événements (ring buffer)

### Tests à implémenter

#### Tests unitaires

- **Fichier** : `internal/orchestrator/dashboard_test.go`
- Scénario 1 : Stats avec données variées
  - Données : 3 projets, 2 plugins, 1 workflow
  - Résultat attendu : `{"projects":3, "plugins":2, "workflows":1, "reports":0, "active_plugins":2}`
- Scénario 2 : Stats vides
  - Données : Aucune donnée
  - Résultat attendu : Tous les compteurs à 0
- Scénario 3 : Activité récente
  - Données : 5 événements
  - Résultat attendu : 5 événements ordonnés par timestamp

- **Fichier** : `internal/orchestrator/static_test.go`
- Scénario 1 : Requête API → ne pas tomber dans SPA fallback
  - Données : GET `/api/v1/plugins`
  - Résultat attendu : Routé vers l'API, pas vers index.html
- Scénario 2 : SPA fallback pour route inconnue
  - Données : GET `/une-page-frontend`
  - Résultat attendu : `index.html` servi avec 200 OK

### Documentation

#### Documentation à créer
- `docs/orchestrator/web-ui.md` — Guide d'utilisation de l'interface web (à compléter dans tâche #005)

#### Documentation à mettre à jour
- `docs/orchestrator/api.md` — Ajouter les endpoints dashboard et WebSocket
