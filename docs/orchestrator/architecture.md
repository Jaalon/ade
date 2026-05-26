# Architecture de l'orchestrateur

## Vue d'ensemble

L'orchestrateur (`ade-config`) est un conteneur Docker léger qui sert de point central
de coordination pour l'environnement de développement. Il expose une API REST (HTTP)
et une API gRPC, embarque une interface web, et coordonne la découverte et la
surveillance des plugins Docker.

## Structure du binaire

```
cmd/ade-config/
└── main.go              # Point d'entrée

internal/orchestrator/
├── config.go            # Configuration (ports, intervalles, etc.)
├── server.go            # Serveur avec healthcheck et cycle de vie
├── client.go            # Client REST pour le CLI
├── router.go            # Routeur HTTP (tâche #002)
├── projects.go          # API des projets (tâche #002)
├── workflows.go         # API des workflows (tâche #002)
├── reports.go           # API des rapports (tâche #002)
├── middleware.go         # Middleware CORS, logging (tâche #002)
├── types.go             # Types partagés (tâche #002)
├── grpc.go              # Serveur gRPC (tâche #003)
├── dashboard.go         # API dashboard web (tâche #004)
├── embed.go             # Embed du frontend (tâche #004)
├── websocket.go         # WebSocket (tâche #004)
├── static_dev.go        # Mode développement frontend (tâche #004)
└── server_test.go       # Tests
```

## Cycle de vie

1. **Démarrage** : `main.go` charge la configuration via `ConfigFromEnv()`,
   crée un `Server`, puis appelle `Start()`.
2. **Signal handling** : Le serveur écoute `SIGINT` et `SIGTERM` pour un arrêt
   gracieux.
3. **Healthcheck** : L'endpoint `GET /health` retourne `{"status":"healthy"}`.
   Utilisé par Docker HEALTHCHECK et par le CLI.
4. **Arrêt** : `Shutdown()` arrête le serveur HTTP avec un timeout de 30s.

## Ports

| Port | Protocole | Usage |
|------|-----------|-------|
| 8080 | HTTP/REST | API REST + interface web |
| 9090 | gRPC | API gRPC pour les plugins |

## Configuration

Variables d'environnement :

| Variable | Défaut | Description |
|----------|--------|-------------|
| `ADE_CONFIG_REST_PORT` | `8080` | Port HTTP REST |
| `ADE_CONFIG_GRPC_PORT` | `9090` | Port gRPC |
| `ADE_CONFIG_DATA_DIR` | `/data` | Répertoire de données |
| `ADE_CONFIG_DISCOVERY_INTERVAL` | `30s` | Intervalle de découverte des plugins |
| `ADE_CONFIG_HEALTH_INTERVAL` | `15s` | Intervalle de healthcheck des plugins |
| `ADE_CONFIG_MAX_HEALTH_FAILS` | `3` | Échecs max avant désenregistrement |
| `ADE_LOG_LEVEL` | `info` | Niveau de log |
