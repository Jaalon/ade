# Tâche #003 - Story #008 : API gRPC et service de découverte/coordination des plugins

## Objectif

Implémenter l'API gRPC de l'orchestrateur pour permettre aux plugins de s'enregistrer, de découvrir d'autres plugins, et de récupérer la configuration. Intégrer le registry plugins existant avec ses mécanismes de découverte Docker et heartbeat.

## Contexte

- Story #008 : [docs/stories/story-008.md](../../stories/story-008.md)
- Dépend de : Tâche #001 (structure du serveur)
- Nécessaire pour : Tâche #006

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Ajouter un serveur gRPC au conteneur orchestrateur et intégrer le registry plugins existant (`internal/plugins/registry/`) découverte, enregistrement, heartbeat.

**Services gRPC à implémenter :**

#### OrchestratorService (nouveau proto)
- `RegisterPlugin(RegisterRequest) → RegisterResponse` — Enregistrement d'un plugin
- `GetConfig(Empty) → ConfigResponse` — Récupération de la configuration
- `ListPlugins(Empty) → ListPluginsResponse` — Liste des plugins enregistrés
- `Heartbeat(HeartbeatRequest) → HeartbeatResponse` — Heartbeat périodique des plugins

#### Intégration avec le registry existant
- Démarrer `registry.Registry` (Store + DockerDiscoverer + HealthChecker + SidecarClient) au lancement du serveur
- Le serveur gRPC partage le même `Store` que l'API REST
- Les plugins découverts par Docker (labels) sont automatiquement enregistrés dans le Store

**Cas nominaux :**
- Plugin s'enregistre via gRPC `RegisterPlugin` → ajouté au Store
- Plugin s'enregistre via REST `POST /api/v1/plugins/register` (handler existant) → même Store
- Plugin découvert via Docker labels → ajouté au Store
- HealthChecker vérifie périodiquement les plugins et marque comme unhealthy si échec

**Cas limites :**
- Plugin déjà enregistré → mise à jour des infos (pas d'erreur)
- Plugin injoignable pendant heartbeat → incrémente `FailedChecks`, retiré après `maxFails`
- Aucun plugin → le serveur reste fonctionnel

**Gestion d'erreurs :**
- Enregistrement avec nom vide → gRPC error `InvalidArgument`
- Plugin introuvable → gRPC error `NotFound`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `api/grpc/orchestrator.proto` | Créer | Définition proto du service Orchestrator |
| `api/grpc/orchestrator.pb.go` | Générer | Code généré Go du proto (via buf) |
| `api/grpc/orchestrator_grpc.pb.go` | Générer | Code généré gRPC Go du proto |
| `internal/orchestrator/grpc.go` | Créer | Implémentation du serveur gRPC |
| `internal/orchestrator/server.go` | Modifier | Ajouter démarrage/arrêt du serveur gRPC |
| `internal/orchestrator/client.go` | Modifier | Ajouter méthodes gRPC au client existant |
| `buf.gen.yaml` | Modifier | Ajouter la génération pour le nouveau proto |
| `buf.yaml` | Modifier | Ajouter la dépendance si nécessaire |

### Signatures

```protobuf
// api/grpc/orchestrator.proto
syntax = "proto3";
package orchestrator;
option go_package = "automated_dev_environment/api/grpc;grpc";

service OrchestratorService {
  rpc RegisterPlugin(RegisterPluginRequest) returns (RegisterPluginResponse);
  rpc GetConfig(GetConfigRequest) returns (GetConfigResponse);
  rpc ListPlugins(ListPluginsRequest) returns (ListPluginsResponse);
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
}

message RegisterPluginRequest {
  string name = 1;
  string version = 2;
  string description = 3;
  string http_address = 4;
  string grpc_address = 5;
  string api_version = 6;
  repeated Capability capabilities = 7;
  map<string, string> endpoints = 8;
}

message Capability {
  string name = 1;
  string version = 2;
  string description = 3;
}

message RegisterPluginResponse {
  bool accepted = 1;
  string message = 2;
}

message GetConfigRequest {}

message GetConfigResponse {
  string project_name = 1;
  string orchestrator_version = 2;
  int32 rest_port = 3;
  int32 grpc_port = 4;
  map<string, string> settings = 5;
}

message ListPluginsRequest {}

message ListPluginsResponse {
  repeated PluginInfo plugins = 1;
}

message PluginInfo {
  string name = 1;
  string version = 2;
  string status = 3;
  string http_address = 4;
  string grpc_address = 5;
  repeated Capability capabilities = 6;
}

message HeartbeatRequest {
  string plugin_name = 1;
}

message HeartbeatResponse {
  bool accepted = 1;
}
```

```go
// internal/orchestrator/grpc.go
package orchestrator

type grpcServer struct {
    orchestrator.UnimplementedOrchestratorServiceServer
    store  *registry.Store
    config *Config
}

func newGRPCServer(store *registry.Store, cfg *Config) *grpcServer
func (s *grpcServer) RegisterPlugin(ctx context.Context, req *orchestrator.RegisterPluginRequest) (*orchestrator.RegisterPluginResponse, error)
func (s *grpcServer) GetConfig(ctx context.Context, req *orchestrator.GetConfigRequest) (*orchestrator.GetConfigResponse, error)
func (s *grpcServer) ListPlugins(ctx context.Context, req *orchestrator.ListPluginsRequest) (*orchestrator.ListPluginsResponse, error)
func (s *grpcServer) Heartbeat(ctx context.Context, req *orchestrator.HeartbeatRequest) (*orchestrator.HeartbeatResponse, error)
```

### Contraintes techniques

- **Framework** : `google.golang.org/grpc` v1.81+ (déjà dans go.mod)
- **Génération** : Utiliser `buf` (déjà configuré dans le projet) pour générer le code proto
- **Intégration** : Le gRPC Server partage le même `*registry.Store` que l'API REST — utilisation du store existant
- **Reflection** : Activer `grpc reflection` (déjà fait dans plugins/sdk)
- **Port** : gRPC sur le port 9090 par défaut
- **Déjà existant** : Le registry REST API (`internal/plugins/registry/api.go`) et sa configuration réseau Docker (sidecar) sont déjà fonctionnels — ne pas les recréer, les intégrer

### Tests à implémenter

#### Tests unitaires

- **Fichier** : `internal/orchestrator/grpc_test.go`
- Scénario 1 : Enregistrement d'un plugin via gRPC
  - Données : `RegisterPluginRequest{Name: "test-plugin", Version: "1.0.0", HttpAddress: "localhost:8081"}`
  - Résultat attendu : `accepted: true`, plugin visible dans le store
- Scénario 2 : Enregistrement avec nom vide
  - Données : `RegisterPluginRequest{Name: ""}`
  - Résultat attendu : Erreur gRPC `InvalidArgument`
- Scénario 3 : Liste des plugins (vide et avec plugins)
  - Données : Store vide → 0 plugins ; après enregistrement → 1 plugin
- Scénario 4 : Récupération de la configuration
  - Données : Config par défaut
  - Résultat attendu : Config valide avec ports et version
- Scénario 5 : Heartbeat d'un plugin enregistré
  - Données : Plugin enregistré, puis Heartbeat
  - Résultat attendu : `accepted: true`, `LastSeen` mis à jour

#### Tests d'intégration
- **Fichier** : `internal/orchestrator/grpc_integration_test.go`
- Scénario : Démarrer le serveur gRPC, enregistrer un plugin en gRPC, vérifier qu'il est listable via REST
- Scénario : Démarrer le registry complet (Store + HealthChecker), vérifier le cycle de vie

### Documentation

#### Documentation à mettre à jour
- `docs/orchestrator/api.md` — Ajouter la section API gRPC
- `docs/plugins/discovery.md` — Mettre à jour avec le mécanisme d'enregistrement gRPC

### Exemples d'utilisation

```go
// Connexion au serveur gRPC
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
if err != nil { log.Fatal(err) }
defer conn.Close()

client := orchestrator.NewOrchestratorServiceClient(conn)

// Enregistrement
resp, err := client.RegisterPlugin(ctx, &orchestrator.RegisterPluginRequest{
    Name:        "mon-plugin",
    Version:     "1.0.0",
    HttpAddress: "mon-plugin:8081",
    GrpcAddress: "mon-plugin:50051",
})
```
