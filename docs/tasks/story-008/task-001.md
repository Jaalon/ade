# Tâche #001 - Story #008 : Dockerfile et point d'entrée du conteneur orchestrateur

## Objectif

Créer le point d'entrée Go (`cmd/ade-config/main.go`) et le `Dockerfile` du conteneur orchestrateur, incluant le healthcheck, la gestion du cycle de vie (démarrage/arrêt gracieux) et le lancement des services internes (API REST, gRPC, web UI).

## Contexte

- Story #008 : [docs/stories/story-008.md](../../stories/story-008.md)
- Dépend de : Aucune (première tâche)
- Nécessaire pour : Tâche #002, Tâche #003, Tâche #004, Tâche #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer le squelette du conteneur orchestrateur : un binaire Go qui s'exécute dans un conteneur Docker, expose un healthcheck HTTP, et orchestre le démarrage/arrêt des sous-systèmes (API REST, API gRPC, registry plugins, web UI).

**Cas nominaux :**
- Le binaire démarre et écoute sur les ports configurés (REST 8080, gRPC 9090)
- Le healthcheck HTTP `/health` retourne `200 OK` avec `{"status":"healthy"}`
- Le binaire gère les signaux SIGTERM/SIGINT pour un arrêt gracieux
- Le **multi-stage build** Docker produit une image légère (~20-50 Mo)

**Cas limites :**
- Ports déjà utilisés → message d'erreur clair et exit code non-zero
- Variables d'environnement manquantes → valeurs par défaut documentées
- Sous-système qui refuse de démarrer → échec rapide (fail fast)

**Gestion d'erreurs :**
- Échec de bind d'un port → erreur fatale avec message explicite
- Timeout d'arrêt gracieux → force kill après 30s

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `cmd/ade-config/main.go` | Créer | Point d'entrée du binaire orchestrateur |
| `internal/orchestrator/server.go` | Créer | Serveur orchestrateur (agrégateur des sous-systèmes) |
| `internal/orchestrator/config.go` | Créer | Configuration de l'orchestrateur (ports, etc.) |
| `Dockerfile` | Créer | Multi-stage Dockerfile (build Go + frontend) |
| `.dockerignore` | Créer | Fichier .dockerignore pour la racine |
| `scripts/build-orchestrator.ps1` | Créer | Script PowerShell pour build de l'image orchestrateur |

### Signatures

```go
// cmd/ade-config/main.go
package main

func main()

// internal/orchestrator/config.go
package orchestrator

type Config struct {
    RESTPort      int
    GRPCPort      int
    WebUIPort     int  // port interne pour le frontend
    DataDir       string
    RegistryPort  int  // port interne du registry plugins
    DiscoveryInterval time.Duration
    HealthInterval    time.Duration
    MaxHealthFails    int
}

func DefaultConfig() Config
func ConfigFromEnv() Config

// internal/orchestrator/server.go
package orchestrator

type Server struct {
    cfg      Config
    registry *registry.Registry
    // À compléter dans les tâches suivantes
}

func NewServer(cfg Config) *Server
func (s *Server) Start(ctx context.Context) error
func (s *Server) Shutdown(ctx context.Context) error
```

### Contraintes techniques

- **Framework** : Go 1.26 standard library (net/http), pas de framework web externe
- **Pattern** : Suivre le pattern de démarrage/arrêt avec context et signal handling déjà utilisé dans `plugins/sdk/server.go`
- **Style** : Respecter les conventions Go du projet (erreurs en français, testify pour les tests)
- **Sécurité** : Pas de secrets en dur, tout passe par variables d'environnement
- **Performance** : Healthcheck léger (pas de base de données)
- **Image** : Utiliser `golang:1.26-alpine` pour le build stage, `alpine:3.21` pour la cible
- **Frontend embed** : Laisser un placeholder dans le serveur de fichiers statiques pour le frontend qui sera ajouté dans la tâche #005

### Dockerfile structure attendue

```dockerfile
# Stage 1 : Build Go
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ade-config ./cmd/ade-config

# Stage 2 : Frontend (placeholder — sera complété dans tâche #005)
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
# Le frontend sera créé dans la tâche #005, le Dockerfile sera
# mis à jour à ce moment pour builder le frontend ici.

# Stage 3 : Image finale
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /ade-config /usr/local/bin/ade-config
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
EXPOSE 8080 9090
ENTRYPOINT ["ade-config"]
```

### Tests à implémenter

#### Tests unitaires

- **Fichier** : `internal/orchestrator/server_test.go`
- Scénario 1 : Création du serveur avec config par défaut
  - Données : `DefaultConfig()`
  - Résultat attendu : `Server` non nil, champs initialisés
- Scénario 2 : Healthcheck retourne 200
  - Données : Serveur démarré
  - Résultat attendu : GET `/health` → 200 OK, body `{"status":"healthy"}`
- Scénario 3 : `ConfigFromEnv` avec variables d'env
  - Données : `ADE_CONFIG_REST_PORT=9090`, `ADE_CONFIG_GRPC_PORT=9999`
  - Résultat attendu : `cfg.RESTPort == 9090`, `cfg.GRPCPort == 9999`
- Scénario 4 : Arrêt gracieux via context cancel
  - Données : Serveur démarré avec contexte
  - Résultat attendu : Après cancel, serveur arrêté sans erreur, ports libérés

### Documentation

#### Documentation à créer
- `docs/orchestrator/architecture.md` — Architecture de l'orchestrateur

#### Documentation à mettre à jour
- `docs/deployment/config-container.md` — Mettre à jour avec la nouvelle image orchestrateur
