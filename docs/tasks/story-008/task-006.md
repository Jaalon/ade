# Tâche #006 - Story #008 : Intégration CLI et mise à jour docker-compose

## Objectif

Implémenter la détection/démarrage automatique de l'orchestrateur par le CLI (`ade`), la récupération de la configuration depuis l'orchestrateur, et mettre à jour le template `docker-compose.yml.tmpl` pour utiliser l'image orchestrateur au lieu de `nginx:alpine`.

## Contexte

- Story #008 : [docs/stories/story-008.md](../../stories/story-008.md)
- Dépend de : Tâche #001, Tâche #002, Tâche #003
- Nécessaire pour : Story #004 (déploiement docker-compose)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**1. Détection et démarrage de l'orchestrateur par le CLI**

Améliorer la fonction `EnsureConfigContainer` dans `internal/docker/docker.go` pour :
- Détecter si le conteneur `ade-config` est en cours d'exécution (déjà fait)
- S'il n'est pas en cours d'exécution mais que l'image est disponible, le démarrer automatiquement
- Si l'image n'est pas disponible, afficher un message invitant à exécuter `ade init ci`
- Ajouter l'état de l'orchestrateur dans la commande `ade version` ou `ade status`

**2. Récupération de la configuration depuis l'orchestrateur**

Ajouter des méthodes au `Client` (`internal/orchestrator/client.go`) pour :
- `GetConfig(ctx) → ConfigResponse` — Récupérer la configuration globale depuis l'API REST
- `GetDashboardStats(ctx) → DashboardStats` — Récupérer les statistiques
- `ListProjects(ctx) → []Project` — Lister les projets
- `CreateProject(ctx, Project) → error` — Créer un projet
- `DeleteProject(ctx, name) → error` — Supprimer un projet

**3. Mise à jour du template docker-compose**

Remplacer `nginx:alpine` dans `internal/templates/embed/docker-compose.yml.tmpl` par l'image orchestrateur publiée `ade/ade-config:latest`. Ajouter le port gRPC (9090) en exposé. Ajouter la variable `ADE_CONFIG_IMAGE` dans `.env` pour surcharger l'image. L'image est buildée et publiée via le script de build dédié, puis distribuée via le registre d'images. Ajouter les volumes nécessaires et la configuration d'environnement.

**4. Script de build de l'image orchestrateur**

Créer un script PowerShell `scripts/build-orchestrator.ps1` pour builder l'image Docker de l'orchestrateur, la tague (`ade/ade-config:latest`) et la publie sur le registre d'images. Le registre d'images (service Docker séparé) maintient la liste des images disponibles et est mis à jour à chaque build.

**Cas nominaux :**
- `ade init ci` génère un docker-compose.yml avec le service `ade-config` utilisant l'image orchestrateur
- Le CLI détecte que l'orchestrateur tourne via le healthcheck HTTP
- La commande `ade plugin list` utilise l'orchestrateur s'il est disponible
- Si l'orchestrateur n'est pas disponible, le CLI continue en mode dégradé

**Cas limites :**
- Docker indisponible → comportement dégradé (pas de panique)
- Orchestrateur non démarré → message clair avec instruction de déploiement
- Conflit de ports (8080 déjà utilisé) → message d'erreur explicite

**Gestion d'erreurs :**
- Timeout de connexion à l'orchestrateur → message "orchestrateur non disponible" en français
- Toutes les fonctions CLI continuent de fonctionner sans orchestrateur (mode dégradé)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/docker/docker.go` | Modifier | Améliorer `EnsureConfigContainer` avec auto-start |
| `internal/orchestrator/client.go` | Modifier | Ajouter les nouvelles méthodes REST |
| `internal/orchestrator/client_test.go` | Modifier | Ajouter les tests des nouvelles méthodes |
| `internal/templates/embed/docker-compose.yml.tmpl` | Modifier | Remplacer nginx:alpine par ade-config |
| `internal/templates/embed/env.tmpl` | Modifier | Ajouter variables d'env pour l'orchestrateur |
| `internal/templates/template.go` | Modifier | Ajouter `ConfigImage` à `ComposeConfig` |
| `internal/command/init_ci.go` | Modifier | Ajouter `ConfigImage` à `toTemplateData` et option `--image` |
| `internal/command/plugin_list.go` | Modifier | Utiliser l'orchestrateur si disponible |
| `internal/command/version.go` | Créer/Modifier | Afficher le statut de l'orchestrateur |
| `scripts/build-orchestrator.ps1` | Créer | Script pour builder l'image Docker orchestrateur |

### Signatures

```go
// Nouvelles méthodes sur internal/orchestrator/client.go

func (c *Client) GetConfig(ctx context.Context) (*ConfigResponse, error)
func (c *Client) GetDashboardStats(ctx context.Context) (*DashboardStats, error)
func (c *Client) ListProjects(ctx context.Context) ([]Project, error)
func (c *Client) CreateProject(ctx context.Context, p *Project) error
func (c *Client) DeleteProject(ctx context.Context, name string) error

// Nouveau type interne
type ConfigResponse struct {
    ProjectName        string            `json:"project_name"`
    OrchestratorVersion string           `json:"orchestrator_version"`
    RESTPort           int               `json:"rest_port"`
    GRPCPort           int               `json:"grpc_port"`
    Settings           map[string]string `json:"settings,omitempty"`
}

// Docker enhance
func EnsureOrchestratorRunning(ctx context.Context) error
```

### Nouveau template docker-compose attendu

Le template référence l'image publiée de l'orchestrateur (pas de `build:` local). L'image est buildée et publiée via le script de build dédié, puis disponible depuis le registre d'images géré par le service registry.

```yaml
version: "3.8"
name: {{.ProjectName}}-preprod

services:
  ade-config:
    image: {{.Compose.ConfigImage}}
    container_name: ade-config
    ports:
      - "{{.Compose.ConfigPort}}:8080"
      - "9090:9090"
    environment:
      - ADE_CONFIG_REST_PORT=8080
      - ADE_CONFIG_GRPC_PORT=9090
      - ADE_PROJECT_NAME={{.ProjectName}}
    env_file:
      - .env
    networks:
      - {{.Compose.Network}}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

networks:
  {{.Compose.Network}}:
    driver: bridge
```

Pour le développement local, le script `scripts/build-orchestrator.ps1` build l'image localement et la tague `ade/ade-config:latest`.
Le service registry séparé gère la liste des images disponibles (plugins, etc.) et est mis à jour à chaque build de l'image registry.

**Note :** Le champ `image:` peut être surchargé via la variable `ADE_CONFIG_IMAGE` dans `.env` pour utiliser une image différente ou un tag spécifique.
Ajouter `ADE_CONFIG_IMAGE={{.Compose.ConfigImage}}` au template `env.tmpl`.

### Contraintes techniques

- **CLI dégradé** : Toute interaction avec l'orchestrateur doit être optionnelle — le CLI ne doit jamais planter si l'orchestrateur est absent
- **Template** : Ajouter `ConfigImage string` à `ComposeConfig` dans `internal/templates/template.go` (valeur par défaut `"ade/ade-config:latest"`)
- **Build image** : Le script `build-orchestrator.ps1` doit builder l'image docker depuis la racine du projet, la taguer `ade/ade-config:latest`, et la publier sur le registre
- **Distribution** : L'image orchestrateur est publiée sur un registre Docker, pas buildée localement par docker-compose. Le template compose référence l'image publiée.
- **Messages** : En français, conformément à la convention du projet
- **Testabilité** : Les fonctions doivent être testables avec des mocks (déjà le pattern avec `dockerCheckFn`, etc.)

### Tests à implémenter

#### Tests unitaires

- **Fichier** : `internal/orchestrator/client_test.go` (ajouts)
- Scénario 1 : `GetConfig` retourne la configuration
  - Données : Mock HTTP serveur retournant `{"project_name":"test","orchestrator_version":"1.0.0"}`
  - Résultat attendu : ConfigResponse valide
- Scénario 2 : `ListProjects` retourne la liste
  - Données : Mock avec 2 projets
  - Résultat attendu : 2 projets dans la liste
- Scénario 3 : `GetDashboardStats` retourne les stats
  - Données : Mock avec stats
  - Résultat attendu : Stats valides

- **Fichier** : `internal/command/init_ci_test.go` (ajouts)
- Scénario : Template render avec nouvelle config orchestrateur
  - Données : Options CI avec port personnalisé et image registry
  - Résultat : docker-compose.yml généré avec `image: ade/ade-config:latest` (pas de `build:` section)

- **Fichier** : `internal/docker/docker_test.go` (ajouts)
- Scénario 1 : `EnsureOrchestratorRunning` avec conteneur déjà en cours d'exécution → pas d'erreur
- Scénario 2 : `EnsureOrchestratorRunning` sans Docker → pas d'erreur (mode dégradé)

### Documentation

#### Documentation à mettre à jour
- `docs/commands/init-ci.md` — Mettre à jour avec la nouvelle configuration du service ade-config
- `docs/deployment/config-container.md` — Documenter le nouveau comportement d'auto-start
- `docs/deployment/preprod.md` — Mettre à jour avec l'architecture orchestrateur
