# API Orchestrateur

## API REST

L'API REST est documentée dans les exemples HTTP :
`docs/tasks/story-008/task-002-examples.http`

Port : 8080 par défaut

## API gRPC

L'API gRPC expose le service `OrchestratorService` sur le port 9090 par défaut.

### Service : OrchestratorService

#### RegisterPlugin

Enregistrement d'un plugin dans le registry.

- **Request** : `RegisterPluginRequest`
  - `name` (string, requis) — Nom du plugin
  - `version` (string) — Version du plugin
  - `description` (string) — Description
  - `http_address` (string) — Adresse HTTP du plugin
  - `grpc_address` (string) — Adresse gRPC du plugin
  - `api_version` (string) — Version de l'API supportée
  - `capabilities` (repeated Capability) — Capacités du plugin
  - `endpoints` (map<string, string>) — Endpoints additionnels
- **Response** : `RegisterPluginResponse` avec `accepted` et `message`
- **Erreurs** : `InvalidArgument` si le nom est vide

#### GetConfig

Récupération de la configuration de l'orchestrateur.

- **Request** : `GetConfigRequest` (vide)
- **Response** : `GetConfigResponse`
  - `project_name` (string) — Nom du projet
  - `orchestrator_version` (string) — Version de l'orchestrateur
  - `rest_port` (int32) — Port REST
  - `grpc_port` (int32) — Port gRPC
  - `settings` (map<string, string>) — Paramètres dynamiques

#### ListPlugins

Liste des plugins enregistrés.

- **Request** : `ListPluginsRequest` (vide)
- **Response** : `ListPluginsResponse`
  - `plugins` (repeated PluginInfo) — Liste des plugins
    - `name`, `version`, `status`, `http_address`, `grpc_address`, `capabilities`

#### Heartbeat

Heartbeat périodique d'un plugin.

- **Request** : `HeartbeatRequest`
  - `plugin_name` (string, requis) — Nom du plugin
- **Response** : `HeartbeatResponse` avec `accepted`
- **Erreurs** : `NotFound` si le plugin est inconnu

### Utilisation depuis un plugin

```go
import (
    "context"
    pb "automated_dev_environment/api/grpc"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

conn, err := grpc.Dial("orchestrateur:9090",
    grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil { /* gérer */ }
defer conn.Close()

client := pb.NewOrchestratorServiceClient(conn)

resp, err := client.RegisterPlugin(ctx, &pb.RegisterPluginRequest{
    Name:        "mon-plugin",
    Version:     "1.0.0",
    HttpAddress: "mon-plugin:8081",
    GrpcAddress: "mon-plugin:50051",
})
```

### Reflection

Le serveur gRPC expose le service de reflection, permettant l'utilisation d'outils
comme `grpcurl` :

```bash
grpcurl -plaintext localhost:9090 list
```
