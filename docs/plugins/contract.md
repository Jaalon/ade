# Contrat d'interface plugin

## Définition protobuf

Le contrat d'interface est défini dans le fichier protobuf :

```
api/grpc/plugin.proto
```

### Service gRPC

```protobuf
service PluginService {
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc HealthCheck(Empty) returns (HealthCheckResponse);
    rpc GetCapabilities(Empty) returns (CapabilitiesResponse);
}
```

### Messages

**PluginDescriptor** — Description d'un plugin :

| Champ | Type | Description |
|-------|------|-------------|
| `name` | string | Nom unique du plugin |
| `version` | string | Version semver |
| `description` | string | Description courte |
| `api_version` | string | Version du contrat d'API |
| `capabilities` | Capability[] | Capacités fonctionnelles |
| `endpoints` | map<string,string> | Endpoints additionnels |

**Capability** — Capacité fonctionnelle :

| Champ | Type | Description |
|-------|------|-------------|
| `name` | string | Nom de la capacité |
| `description` | string | Description |
| `version` | string | Version de la capacité |
| `config_schema` | string | JSON Schema optionnel |

**HealthStatus** — État de santé :

| Champ | Type | Description |
|-------|------|-------------|
| `status` | enum | UNKNOWN, HEALTHY, DEGRADED, UNHEALTHY |
| `message` | string | Message optionnel |
| `timestamp` | int64 | Timestamp Unix |

## Endpoints REST

Les endpoints REST suivants sont équivalents aux services gRPC :

| Endpoint | Méthode | Équivalent gRPC |
|----------|---------|-----------------|
| `POST /api/v1/plugins/register` | Push | `Register` |
| `GET /api/v1/plugins/health` | - | `HealthCheck` |
| `GET /api/v1/plugins/capabilities` | - | `GetCapabilities` |

## Génération de code

Le code Go est généré à partir du protobuf avec `buf` :

```bash
buf generate
```

Fichiers générés :
- `internal/plugins/contract/plugin.pb.go` — Messages
- `internal/plugins/contract/plugin_grpc.pb.go` — Service gRPC
- `internal/plugins/contract/types.go` — Types Go REST (HealthStatus, erreurs)

## Versionnement

Le champ `api_version` permet de versionner le contrat plugin/orchestrateur :

- `v1` — Version initiale (actuelle)
- Les plugins et l'orchestrateur doivent déclarer la même version pour être compatibles
- Le registry refuse l'enregistrement si les versions API diffèrent
