# Développement d'un plugin

## Prérequis

- Go 1.26
- Docker
- Compréhension de base de REST et gRPC
- `buf` installé pour la compilation protobuf

## SDK

Le SDK est accessible de deux manières :

### Plugins first-party (dans ce dépôt)

```go
import "automated_dev_environment/internal/plugins/sdk"
import "automated_dev_environment/internal/plugins/contract"
```

### Plugins externes

```go
import "github.com/ade/plugins-sdk"
import "github.com/ade/plugins-sdk/contract"
```

Le module SDK externe est extrait via le script `scripts/extract-sdk.ps1`.

## Étapes de création d'un plugin

### 1. Initialiser un module Go

```bash
go mod init mon-plugin
```

### 2. Importer le SDK

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/ade/plugins-sdk"
    "github.com/ade/plugins-sdk/contract"
)
```

### 3. Créer le PluginServer

```go
server, err := sdk.NewPlugin()
if err != nil {
    panic(err)
}
```

Configuration via variables d'environnement :

| Variable | Obligatoire | Défaut | Description |
|----------|-------------|--------|-------------|
| `PLUGIN_NAME` | Oui | - | Nom du plugin |
| `PLUGIN_VERSION` | Oui | - | Version du plugin |
| `PLUGIN_DESCRIPTION` | Non | - | Description |
| `PLUGIN_ORCHESTRATOR_URL` | Oui | - | URL du registry |
| `PLUGIN_HTTP_PORT` | Non | 8081 | Port HTTP |
| `PLUGIN_GRPC_PORT` | Non | 50051 | Port gRPC |
| `PLUGIN_REGISTER_INTERVAL` | Non | 30s | Intervalle de ré-enregistrement |

### 4. Ajouter des capacités

```go
server.AddCapability(&contract.Capability{
    Name:        "template_provider",
    Description: "Provides project templates",
    Version:     "1.0.0",
})
```

### 5. Enregistrer des handlers REST

```go
server.HandleFunc("/api/v1/templates", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"templates": ["go-api", "go-cli"]}`))
})
```

### 6. Démarrer le serveur

```go
if err := server.Start(context.Background()); err != nil {
    panic(err)
}
```

`Start` bloque jusqu'à SIGTERM/SIGINT et gère l'arrêt gracieux.

## Exemple complet : Plugin hello-world

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/ade/plugins-sdk"
    "github.com/ade/plugins-sdk/contract"
)

func main() {
    server, _ := sdk.NewPlugin(
        sdk.WithConfig(sdk.Config{
            Name:            "hello-world",
            Version:         "0.1.0",
            Description:     "Mon premier plugin",
            OrchestratorURL: "http://registry:8082",
        }),
    )

    server.AddCapability(&contract.Capability{
        Name:        "greeter",
        Description: "Dis bonjour",
        Version:     "1.0.0",
    })

    server.HandleFunc("/api/v1/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"message": "Hello from plugin!"}`))
    })

    if err := server.Start(context.Background()); err != nil {
        panic(err)
    }
}
```

## Dockerfile

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /plugin .

FROM gcr.io/distroless/base:latest
COPY --from=builder /plugin /plugin

LABEL ade.plugin.name=hello-world
LABEL ade.plugin.version=0.1.0
LABEL ade.plugin.http-port=8081
LABEL ade.plugin.grpc-port=50051

EXPOSE 8081 50051
ENTRYPOINT ["/plugin"]
```

## Test avec Docker Compose

```yaml
services:
  hello-plugin:
    build: ./plugins/hello-world
    environment:
      PLUGIN_NAME: hello-world
      PLUGIN_VERSION: "0.1.0"
      PLUGIN_ORCHESTRATOR_URL: "http://registry:8082"
      PLUGIN_HTTP_PORT: "8081"
    labels:
      ade.plugin.name: "hello-world"
      ade.plugin.version: "0.1.0"
    networks:
      - ade-network
```

## Contrat protobuf

Défini dans `api/grpc/plugin.proto`, généré avec :

```bash
buf generate
```

## Configuration de l'orchestrateur

Le plugin s'enregistre auprès du registry sur le port 8082 (pas le port public 8080 de l'orchestrateur). L'orchestrateur proxyfie les appels `/api/v1/plugins/*` depuis le port 8080 vers 8082.

En environnement Docker Compose, `PLUGIN_ORCHESTRATOR_URL=http://registry:8082`.
