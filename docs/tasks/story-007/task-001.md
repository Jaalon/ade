# Tâche #001 - Story #007 : Contrat plugin (protobuf + spécification REST)

## Objectif
Définir le contrat d'interface commun que tous les plugins Docker doivent implémenter : définition protobuf du service gRPC, spécification des endpoints REST/HTTP équivalents, et structures de données partagées.

## Contexte
- Story #007 : `docs/stories/story-007.md`
- Dépend de : Story #008 (l'orchestrateur expose déjà une API REST + gRPC)
- Nécessaire pour : Tâches #002, #003, #004, #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer le contrat d'interface qu'un plugin Docker doit exposer. Chaque plugin expose deux protocoles strictement équivalents (REST et gRPC). Le contrat inclut le descripteur du plugin, les endpoints de découverte, health check, et le service métier générique.

**Cas nominaux :**
- Un fichier `.proto` définit le service `PluginService` avec :
  - `Register(request)` → `(PluginDescriptor, error)` — appelé par le plugin vers l'orchestrateur
  - `HealthCheck(Empty)` → `(HealthStatus, error)` — appelé par l'orchestrateur vers le plugin
  - `GetCapabilities(Empty)` → `(Capabilities, error)` — description des capacités du plugin
- Les endpoints REST correspondants sont définis :
  - `POST /api/v1/plugins/register` (appelé par le plugin)
  - `GET /api/v1/plugins/health` (appelé par l'orchestrateur)
  - `GET /api/v1/plugins/capabilities` (appelé par l'orchestrateur)
- Le descripteur `PluginDescriptor` contient :
  - `name` (string) — nom unique du plugin
  - `version` (string) — version semver
  - `description` (string) — description courte
  - `api_version` (string) — version du contrat d'API plugin (ex: "v1")
  - `capabilities` ([]Capability) — liste des capacités
  - `endpoints` (map<string,string>) — endpoints additionnels (ex: "templates" → "/api/v1/templates")
- La capacité `Capability` contient :
  - `name` (string) — nom de la capacité
  - `description` (string) — description
  - `version` (string) — version de cette capacité
  - `config_schema` (string) — JSON Schema optionnel de configuration
- Le statut `HealthStatus` contient :
  - `status` (enum: UNKNOWN, HEALTHY, DEGRADED, UNHEALTHY)
  - `message` (string) — message optionnel
  - `timestamp` (int64) — timestamp Unix
- La génération de code Go depuis le `.proto` produit :
  - `internal/plugins/contract/plugin_grpc.pb.go`
  - `internal/plugins/contract/plugin.pb.go`
- Les types Go pour REST sont dans `internal/plugins/contract/types.go` (sans protobuf, pour éviter la dépendance gRPC côté client REST simple)

**Cas limites :**
- Un plugin peut avoir zéro capacité (utile pour les plugins purement utilitaires)
- Un plugin peut annoncer des endpoints optionnels — l'orchestrateur ne doit pas échouer si un endpoint annoncé n'est pas joignable
- Le champ `config_schema` peut être vide si le plugin n'a pas de configuration
- L'`api_version` permet de versionner le contrat plugin/orchestrateur

**Gestion d'erreurs :**
- Si un plugin ne répond pas sur un endpoint annoncé → `status = DEGRADED` dans le health check
- Si le protobuf ne compile pas → erreur bloquante
- Si les endpoints REST renvoient autre chose que 200 → `status = UNHEALTHY`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `api/grpc/plugin.proto` | Créer | Définition protobuf du service plugin |
| `internal/plugins/contract/plugin.pb.go` | Créer | Code généré protobuf (messages) |
| `internal/plugins/contract/plugin_grpc.pb.go` | Créer | Code généré protobuf (service gRPC) |
| `internal/plugins/contract/types.go` | Créer | Types Go équivalents pour REST (DTO) |
| `internal/plugins/contract/errors.go` | Créer | Erreurs sentinelles du contrat |
| `internal/plugins/contract/contract_test.go` | Créer | Tests de validation du contrat |
| `buf.yaml` | Créer | Configuration buf du module protobuf |
| `buf.gen.yaml` | Créer | Configuration de génération Go |
| `api/grpc/buf.lock` | Créer | Lock file buf (généré) |
| `scripts/generate.ps1` | Modifier | Ajouter la section génération protobuf |
| `tools/tools.go` | Modifier | Ajouter `buf` comme outil Go |

### Signatures

```go
// internal/plugins/contract/types.go
package contract

// PluginDescriptor décrit un plugin enregistré.
type PluginDescriptor struct {
    Name         string              `json:"name"`
    Version      string              `json:"version"`
    Description  string              `json:"description"`
    APIVersion   string              `json:"api_version"`
    Capabilities []Capability        `json:"capabilities"`
    Endpoints    map[string]string   `json:"endpoints,omitempty"`
}

// Capability décrit une capacité fonctionnelle d'un plugin.
type Capability struct {
    Name         string `json:"name"`
    Description  string `json:"description"`
    Version      string `json:"version"`
    ConfigSchema string `json:"config_schema,omitempty"`
}

// HealthStatus représente l'état de santé d'un plugin.
type HealthStatus struct {
    Status    HealthStatusEnum `json:"status"`
    Message   string           `json:"message,omitempty"`
    Timestamp int64            `json:"timestamp"`
}

type HealthStatusEnum int

const (
    HealthUnknown   HealthStatusEnum = 0
    HealthHealthy   HealthStatusEnum = 1
    HealthDegraded  HealthStatusEnum = 2
    HealthUnhealthy HealthStatusEnum = 3
)
```

### Contraintes techniques
- **gRPC** : Utiliser `google.golang.org/grpc` et `google.golang.org/protobuf`
- **Protobuf** : Version proto3, fichier `api/grpc/plugin.proto`
- **Go generate** : Ajouter un `//go:generate` commentaire dans `internal/plugins/contract/` : `//go:generate buf generate`
- **Outil protobuf** : Utiliser `buf` (v1.50+) pour la génération. Ajouter les fichiers de configuration :
  - `buf.yaml` à la racine (définit le module protobuf)
  - `buf.gen.yaml` (configuration de génération Go)
  - `api/grpc/buf.lock` (lock file, généré)
- **Script** : Ajouter une section dans `scripts/generate.ps1` pour l'installation et l'exécution de `buf`
- **Dépendances buf** : Ajouter `buf` aux outils dans `tools/tools.go` via `go install github.com/bufbuild/buf/cmd/buf@latest`
- **Package naming** : `automated_dev_environment/internal/plugins/contract`
- **Version API** : `api_version = "v1"` pour cette première version

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/plugins/contract/contract_test.go`
- Scénario 1 : Création d'un PluginDescriptor valide → champs corrects
- Scénario 2 : Sérialisation JSON d'un PluginDescriptor → JSON valide avec tous les champs
- Scénario 3 : Désérialisation JSON → PluginDescriptor identique
- Scénario 4 : HealthStatus avec tous les statuts → valeurs enum correctes
- Scénario 5 : Capability avec ConfigSchema vide → champ omis en JSON (omitempty)
- Scénario 6 : PluginDescriptor avec endpoints vides → champ omis en JSON

### Documentation
- Mettre à jour `docs/plugins/architecture.md` : section "Contrat d'interface plugin" référençant le protobuf
- Créer `docs/plugins/contract.md` : Description du contrat, des endpoints REST et gRPC

### Exemples d'API

```yaml
# buf.yaml
version: v2
modules:
  - path: api/grpc
    name: ade/plugin
deps: []
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

```yaml
# buf.gen.yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: internal/plugins/contract
    opt: paths=source_relative
  - local: protoc-gen-go-grpc
    out: internal/plugins/contract
    opt: paths=source_relative
```

```protobuf
// api/grpc/plugin.proto
syntax = "proto3";
package plugin;

option go_package = "automated_dev_environment/internal/plugins/contract;contract";

service PluginService {
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc HealthCheck(Empty) returns (HealthCheckResponse);
    rpc GetCapabilities(Empty) returns (CapabilitiesResponse);
}

message RegisterRequest {
    string name = 1;
    string version = 2;
    string api_version = 3;
    repeated Capability capabilities = 4;
    map<string, string> endpoints = 5;
}
```

```http
### Enregistrement du plugin (REST)
POST /api/v1/plugins/register
Content-Type: application/json

{
    "name": "templates",
    "version": "1.0.0",
    "api_version": "v1",
    "capabilities": [
        {"name": "template_provider", "description": "Provides project templates", "version": "1.0.0"}
    ],
    "endpoints": {
        "templates": "/api/v1/templates"
    }
}

### Health check
GET /api/v1/plugins/health

### Capacités
GET /api/v1/plugins/capabilities
```
