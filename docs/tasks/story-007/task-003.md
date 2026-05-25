# Tâche #003 - Story #007 : Découverte et enregistrement des plugins (côté orchestrateur)

## Objectif
Implémenter le mécanisme de découverte et d'enregistrement des plugins côté orchestrateur : endpoint d'enregistrement, découverte via labels Docker, health checking périodique, et API de gestion des plugins.

## Contexte
- Story #007 : `docs/stories/story-007.md`
- Dépend de : Tâche #001 (contrat plugin), Story #008 (orchestrateur)
- Nécessaire pour : Tâche #004, #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Ajouter à l'orchestrateur (Story #008) la capacité de découvrir, enregistrer, surveiller et gérer les plugins Docker. Deux mécanismes de découverte sont supportés : (1) enregistrement push (le plugin s'enregistre via API), (2) découverte passive (scan des conteneurs Docker avec labels spécifiques).

**Cas nominaux :**
- Endpoint `POST /api/v1/plugins/register` — un plugin s'enregistre (ou met à jour son enregistrement)
- Endpoint `GET /api/v1/plugins` — liste tous les plugins enregistrés avec leur statut
- Endpoint `GET /api/v1/plugins/{name}` — détails d'un plugin spécifique
- Endpoint `DELETE /api/v1/plugins/{name}` — désenregistre un plugin
- Découverte Docker : scanner les conteneurs avec le label `ade.plugin.name=...` à intervalle configurable
- Health check périodique : l'orchestrateur ping chaque plugin toutes les 30s
- Les plugins sont stockés en mémoire avec possibilité de persistance (fichier JSON)

**Cas limites :**
- Deux plugins avec le même nom → le plus récent remplace l'ancien (avec log d'avertissement)
- Un plugin qui ne répond plus au health check → `status = UNHEALTHY`, retiré après 3 échecs consécutifs
- Aucun plugin enregistré → les endpoints retournent une liste vide (pas d'erreur)
- Plugin qui se ré-enregistre avec des capacités modifiées → mise à jour des capacités
- La découverte Docker ne doit pas planter si le démon Docker est injoignable

**Gestion d'erreurs :**
- Plugin inconnu → 404 avec message explicite
- Corps de requête invalide → 400 Bad Request
- Docker injoignable → log warning, pas d'erreur remontée au client

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/plugins/registry/store.go` | Créer | Stockage en mémoire des plugins enregistrés |
| `internal/plugins/registry/discovery.go` | Créer | Découverte via labels Docker (via side-car proxy) |
| `internal/plugins/registry/discovery_sidecar.go` | Créer | Client HTTP vers le side-car Docker API proxy |
| `internal/plugins/registry/health.go` | Créer | Health check périodique des plugins |
| `internal/plugins/registry/api.go` | Créer | Handlers HTTP pour l'API de gestion des plugins |
| `internal/plugins/registry/registry.go` | Créer | Service de registry (coordination) |
| `internal/plugins/registry/registry_test.go` | Créer | Tests du registry |
| `internal/plugins/registry/server.go` | Créer | Serveur HTTP standalone du registry (port 8082) |

### Signatures

```go
// internal/plugins/registry/store.go
package registry

// PluginInstance représente un plugin enregistré et suivi.
type PluginInstance struct {
    Descriptor   contract.PluginDescriptor `json:"descriptor"`
    Status       contract.HealthStatusEnum `json:"status"`
    LastSeen     time.Time                 `json:"last_seen"`
    FailedChecks int                       `json:"failed_checks"`
    GrpcAddress  string                    `json:"grpc_address"`
    HttpAddress  string                    `json:"http_address"`
    Labels       map[string]string         `json:"labels,omitempty"`
}

// Store gère le stockage des instances de plugins.
type Store struct {
    plugins map[string]*PluginInstance
    mu      sync.RWMutex
}

func NewStore() *Store
func (s *Store) Register(instance *PluginInstance) error
func (s *Store) Unregister(name string) error
func (s *Store) Get(name string) (*PluginInstance, bool)
func (s *Store) List() []*PluginInstance
func (s *Store) UpdateHealth(name string, status contract.HealthStatusEnum)


// internal/plugins/registry/discovery.go
// DockerDiscoverer découvre les plugins via les labels des conteneurs Docker.
type DockerDiscoverer struct {
    dockerClient docker.Client
    store        *Store
}

func NewDockerDiscoverer(dockerClient docker.Client, store *Store) *DockerDiscoverer
func (d *DockerDiscoverer) Start(ctx context.Context)
func (d *DockerDiscoverer) discover(ctx context.Context) error


// internal/plugins/registry/health.go
// HealthChecker surveille l'état de santé des plugins.
type HealthChecker struct {
    store     *Store
    interval  time.Duration
    maxFails  int
}

func NewHealthChecker(store *Store, interval time.Duration, maxFails int) *HealthChecker
func (h *HealthChecker) Start(ctx context.Context)


// internal/plugins/registry/api.go
// API gère les endpoints REST pour la gestion des plugins.
type API struct {
    store    *Store
    discover *DockerDiscoverer
}

func NewAPI(store *Store, discover *DockerDiscoverer) *API
func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request)
func (a *API) ListHandler(w http.ResponseWriter, r *http.Request)
func (a *API) GetHandler(w http.ResponseWriter, r *http.Request)
func (a *API) DeleteHandler(w http.ResponseWriter, r *http.Request)


// internal/plugins/registry/server.go
// Server encapsule le serveur HTTP du registry (port 8082).
type Server struct {
    api    *API
    server *http.Server
    port   int
}

func NewServer(api *API, port int) *Server
func (s *Server) Start(ctx context.Context) error
func (s *Server) Shutdown(ctx context.Context) error
```

### Contraintes techniques
- **Port séparé** : Le registry expose son API HTTP sur le port **8082** (pas sur le routeur principal de l'orchestrateur). L'orchestrateur proxyfie les appels `GET/POST/DELETE /api/v1/plugins/*` depuis son port 8080 vers `localhost:8082`.
- **Labels Docker** : Les conteneurs plugins doivent avoir le label `ade.plugin.name=<name>`. Labels additionnels : `ade.plugin.version`, `ade.plugin.http-port`, `ade.plugin.grpc-port`.
- **Découverte via side-car** : Le `DockerDiscoverer` ne parle pas directement au socket Docker. Il utilise un side-car proxy HTTP (conteneur dédié dans docker-compose) qui expose une API REST restreinte pour lister les conteneurs avec leurs labels. Le side-car est le seul conteneur avec accès au socket Docker.
- **Découverte push** : Les plugins s'enregistrent via `POST /api/v1/plugins/register` sur le port 8082 (le SDK le fait automatiquement)
- **Découverte pull** : Le registry interroge le side-car Docker proxy toutes les 30s (`ADE_PLUGIN_DISCOVERY_INTERVAL`)
- **Intervalle** : Découverte Docker par défaut toutes les 30s, configurable via variable d'environnement `ADE_PLUGIN_DISCOVERY_INTERVAL`
- **Health check** : Timeout de 5s par requête HTTP vers le plugin, intervalle de 30s
- **Max failures** : 3 échecs consécutifs → désenregistrement automatique
- **Sécurité** : Pas d'authentification pour les plugins en réseau Docker interne (réseau isolé). Le side-car proxy expose uniquement `GET /containers?label=ade.plugin.*` en lecture seule.

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/plugins/registry/registry_test.go`
- Scénario 1 : `Store.Register` ajoute un plugin → `Get` le retourne
- Scénario 2 : `Store.Register` avec le même nom → remplacement
- Scénario 3 : `Store.Unregister` → `Get` retourne false
- Scénario 4 : `Store.List` avec 3 plugins → retourne 3 instances
- Scénario 5 : `Store.UpdateHealth` → statut mis à jour, `FailedChecks` reset ou incrémente selon le statut
- Scénario 6 : HealthChecker marque un plugin comme UNHEALTHY après 3 échecs
- Scénario 7 : API ListHandler → JSON avec la liste
- Scénario 8 : API GetHandler sur plugin inconnu → 404

#### Tests d'intégration
- **Fichier** : `internal/plugins/registry/registry_test.go`
- Scénario 9 : DockerDiscoverer avec side-car proxy mock → découvre les conteneurs avec labels ade.plugin.name
- Scénario 10 : Server.Start + appel HTTP à `/api/v1/plugins/register` → enregistrement réussi
- Scénario 11 : Server.Start + appel HTTP à `/api/v1/plugins/list` → liste des plugins

### Documentation
- Mettre à jour `docs/plugins/discovery.md` : section "Découverte via labels Docker" et "API d'enregistrement"
- Mettre à jour `docs/orchestrator/api.md` : ajouter les endpoints `/api/v1/plugins/*`

### Exemples d'API

```http
### Enregistrer un plugin
POST /api/v1/plugins/register
Content-Type: application/json

{
    "name": "templates",
    "version": "1.0.0",
    "api_version": "v1",
    "description": "Fournit des templates de projets",
    "grpc_address": "templates-plugin:50051",
    "http_address": "templates-plugin:8081",
    "capabilities": [
        {"name": "template_provider", "description": "Fournit des templates", "version": "1.0.0"}
    ]
}

### Lister les plugins
GET /api/v1/plugins

### Détails d'un plugin
GET /api/v1/plugins/templates

### Désenregistrer un plugin
DELETE /api/v1/plugins/templates
```

```yaml
# docker-compose.yml (extrait avec side-car proxy)
services:
  orchestrator:
    image: ade-orchestrator:latest
    ports:
      - "8080:8080"
    networks:
      - ade-network

  registry:
    image: ade-orchestrator:latest  # même image, lance le registry sur 8082
    command: ["ade-orchestrator", "registry"]
    ports:
      - "8082:8082"
    networks:
      - ade-network
    depends_on:
      - docker-proxy

  docker-proxy:
    image: ade-docker-proxy:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - ade-network
    ports:
      - "8083:8083"  # API read-only pour lister les conteneurs

networks:
  ade-network:
    driver: bridge
```

```docker
# Exemple de label Docker pour la découverte
docker run -d \
  --label ade.plugin.name=templates \
  --label ade.plugin.version=1.0.0 \
  --label ade.plugin.http-port=8081 \
  --label ade.plugin.grpc-port=50051 \
  --network ade-network \
  templates-plugin:latest
```
