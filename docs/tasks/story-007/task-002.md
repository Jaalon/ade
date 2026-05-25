# Tâche #002 - Story #007 : Plugin SDK (bibliothèque Go pour développeurs de plugins)

## Objectif
Créer une bibliothèque Go réutilisable (Plugin SDK) qui encapsule la configuration du serveur REST+ gRPC, l'enregistrement auprès de l'orchestrateur, le health check, la gestion des capacités, et le cleanup, afin de simplifier le développement de nouveaux plugins.

## Contexte
- Story #007 : `docs/stories/story-007.md`
- Dépend de : Tâche #001 (contrat protobuf + types)
- Nécessaire pour : Tâche #004 (plugin exemple)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer un SDK Go (package `internal/plugins/sdk`) qui fournit une structure `PluginServer` prête à l'emploi. Le développeur d'un plugin instancie un `PluginServer`, enregistre ses gestionnaires métier, puis appelle `Start()`.

**Cas nominaux :**
- `PluginServer` démarre un serveur gRPC et un serveur REST/HTTP sur des ports configurables
- `PluginServer` s'enregistre automatiquement auprès de l'orchestrateur via l'API REST de l'orchestrateur
- `PluginServer` expose les endpoints obligatoires (`/health`, `/capabilities`, `/register`)
- Le développeur ajoute ses propres endpoints métier via `HandleFunc(path, handler)` (REST) et `RegisterService(desc, impl)` (gRPC)
- `PluginServer` répond au health check de l'orchestrateur automatiquement
- `PluginServer` supporte la configuration via variables d'environnement :
  - `PLUGIN_NAME` (obligatoire) — nom du plugin
  - `PLUGIN_VERSION` (obligatoire) — version du plugin
  - `PLUGIN_DESCRIPTION` (optionnel) — description
  - `PLUGIN_ORCHESTRATOR_URL` (obligatoire) — URL de l'orchestrateur (ex: `http://orchestrator:8080`)
  - `PLUGIN_GRPC_PORT` (défaut: 50051) — port gRPC
  - `PLUGIN_HTTP_PORT` (défaut: 8081) — port HTTP/REST
  - `PLUGIN_REGISTER_INTERVAL` (défaut: 30s) — intervalle de ré-enregistrement

**Cas limites :**
- Si l'orchestrateur est injoignable au démarrage, le plugin continue à tourner et réessaie de s'enregistrer périodiquement
- Si `PLUGIN_NAME` n'est pas défini → erreur fatale au démarrage
- Si un handler REST panique → le SDK doit catcher la panic et retourner une erreur 500
- Le serveur doit supporter un délai de grâce pour le shutdown graceful (SIGTERM/SIGINT)

**Gestion d'erreurs :**
- Orchestrateur injoignable → log warning + réessai dans `PLUGIN_REGISTER_INTERVAL`
- Port déjà utilisé → erreur fatale avec le message explicite
- Handler métier qui retourne une erreur → le SDK doit la traduire en réponse HTTP appropriée (4xx/5xx)
- Panique dans un handler → récupérée, logguée, retour 500

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/plugins/sdk/server.go` | Créer | Serveur principal (gRPC + REST) |
| `internal/plugins/sdk/config.go` | Créer | Configuration depuis l'environnement |
| `internal/plugins/sdk/registration.go` | Créer | Enregistrement auprès de l'orchestrateur |
| `internal/plugins/sdk/health.go` | Créer | Health check handler (REST + gRPC) |
| `internal/plugins/sdk/capabilities.go` | Créer | Gestion des capacités déclarées |
| `internal/plugins/sdk/middleware.go` | Créer | Middleware (recovery, logging, CORS) |
| `internal/plugins/sdk/errors.go` | Créer | Erreurs du SDK |
| `internal/plugins/sdk/sdk_test.go` | Créer | Tests du SDK |
| `internal/plugins/sdk/extract.go` | Créer | Marqueur pour l'extraction du SDK (fichier avec `//go:build extract`) |
| `plugins/sdk/go.mod` | Créer | Module Go indépendant du SDK (pour publication externe) |
| `plugins/sdk/` | Créer | Copie du SDK extraite (générée par script) |
| `scripts/extract-sdk.ps1` | Créer | Script PowerShell d'extraction du SDK |
| `internal/plugins/contract/` | Dépend de | Types déjà définis dans la Tâche #001 |

### Signatures

```go
// internal/plugins/sdk/server.go
package sdk

// PluginServer est le serveur principal d'un plugin.
type PluginServer struct {
    // champs internes (non exportés)
}

// NewPlugin crée une nouvelle instance de PluginServer.
// Les champs sont lus depuis l'environnement (voir config.go).
func NewPlugin(opts ...Option) (*PluginServer, error)

// HandleFunc enregistre un handler HTTP REST sur le chemin donné.
func (s *PluginServer) HandleFunc(path string, handler http.HandlerFunc)

// RegisterService enregistre un service gRPC.
func (s *PluginServer) RegisterService(desc *grpc.ServiceDesc, impl interface{})

// AddCapability déclare une capacité du plugin.
func (s *PluginServer) AddCapability(cap contract.Capability)

// Start démarre les serveurs gRPC et HTTP, puis s'enregistre auprès de l'orchestrateur.
// Bloque jusqu'à SIGTERM/SIGINT ou erreur fatale.
func (s *PluginServer) Start(ctx context.Context) error

// Shutdown arrête proprement les serveurs avec un timeout.
func (s *PluginServer) Shutdown(ctx context.Context) error


// internal/plugins/sdk/config.go
// Config représente la configuration d'un plugin lue depuis l'environnement.
type Config struct {
    Name              string
    Version           string
    Description       string
    OrchestratorURL   string
    GRPCPort          int
    HTTPPort          int
    RegisterInterval  time.Duration
}

func LoadConfigFromEnv() (Config, error)


// internal/plugins/sdk/registration.go
// Registration gère l'enregistrement périodique du plugin.
type Registration struct {
    descriptor *contract.PluginDescriptor
    orchURL    string
    interval   time.Duration
    client     *http.Client
}

func NewRegistration(descriptor *contract.PluginDescriptor, orchURL string, interval time.Duration) *Registration
func (r *Registration) Start(ctx context.Context)
func (r *Registration) RegisterNow(ctx context.Context) error
func (r *Registration) Stop()


// Options fonctionnelles pour PluginServer
type Option func(*PluginServer)

func WithConfig(cfg Config) Option
func WithLogger(logger *log.Logger) Option
func WithGRPCOptions(opts ...grpc.ServerOption) Option
```

### Contraintes techniques
- **Routeur HTTP** : Utiliser la `net/http` standard de Go + `http.ServeMux` (zéro dépendance externe)
- **gRPC** : `google.golang.org/grpc` pour le serveur gRPC
- **Contexte** : Toutes les opérations réseau prennent un `context.Context`
- **Logging** : Utiliser le package `log` standard de Go (pas de dépendance externe)
- **Shutdown** : Écouter `os.Signal` pour SIGTERM/SIGINT, puis appeler `Shutdown` avec un timeout de 10s
- **Serveur HTTP** : Démarrer dans une goroutine, gérer les erreurs de listen via un channel
- **CORS** : Middleware CORS permissif (permet à l'orchestrateur de contacter le plugin depuis son domaine)
- **Structure duale** : Le SDK est développé dans `internal/plugins/sdk/` pour les plugins first-party, avec un script d'extraction `scripts/extract-sdk.ps1` qui copie le package vers `plugins/sdk/` et génère un `go.mod` indépendant pour les développeurs externes
- **Enregistrement cible** : Le SDK s'enregistre auprès du service registry exposé sur le port 8082 de l'orchestrateur (`PLUGIN_ORCHESTRATOR_URL` = `http://orchestrator:8082` pour l'enregistrement ; l'orchestrateur proxyfie `/api/v1/plugins/*` depuis le port 8080 vers 8082)

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/plugins/sdk/sdk_test.go`
- Scénario 1 : `LoadConfigFromEnv` avec toutes les variables → Config correctement remplie
- Scénario 2 : `LoadConfigFromEnv` sans `PLUGIN_NAME` → erreur
- Scénario 3 : `LoadConfigFromEnv` avec valeurs par défaut → ports par défaut, intervalle par défaut
- Scénario 4 : `NewPlugin` avec options → PluginServer configuré correctement
- Scénario 5 : `AddCapability` → capacité listée dans `GetCapabilities`
- Scénario 6 : Handler REST enregistré via `HandleFunc` → répond sur le bon chemin
- Scénario 7 : Middleware recovery → handler qui panique retourne 500

#### Tests d'intégration
- **Fichier** : `internal/plugins/sdk/sdk_test.go`
- Scénario 8 : Démarrage et arrêt du serveur → pas d'erreur, port libéré après arrêt
- Scénario 9 : Endpoint `/health` répond avec 200 et status HEALTHY
- Scénario 10 : Endpoint `/capabilities` retourne la liste des capacités en JSON
- Scénario 11 : Enregistrement auprès d'un orchestrateur mock → appel POST effectué

### Documentation
- Créer `docs/plugins/development.md` : Guide de développement d'un plugin incluant les références au SDK
- Section "Utilisation du SDK" dans `docs/plugins/development.md` avec exemple de code

### Exemples d'utilisation

```go
package main

import (
    "context"
    "net/http"
    "automated_dev_environment/internal/plugins/contract"
    "automated_dev_environment/internal/plugins/sdk"
)

func main() {
    server, err := sdk.NewPlugin()
    if err != nil {
        panic(err)
    }

    server.AddCapability(contract.Capability{
        Name:        "template_provider",
        Description: "Fournit des templates de projets",
        Version:     "1.0.0",
    })

    server.HandleFunc("/api/v1/templates", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"templates": ["go-api", "go-cli"]}`))
    })

    if err := server.Start(context.Background()); err != nil {
        panic(err)
    }
}
```
