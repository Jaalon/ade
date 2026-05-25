# Tâche #006 - Story #007 : Documentation de l'architecture des plugins

## Objectif
Créer la documentation complète de l'architecture des plugins Docker (REST + gRPC), le guide de développement, et les exemples pour permettre aux développeurs de créer et intégrer leurs propres plugins.

## Contexte
- Story #007 : `docs/stories/story-007.md`
- Dépend de : Tâches #001 à #005
- Nécessaire pour : Rien (tâche finale)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer et mettre à jour les fichiers de documentation pour l'architecture des plugins Docker, couvrant la conception, le mécanisme de découverte, le guide de développement, et les exemples concrets.

**Cas nominaux :**
- `docs/plugins/architecture.md` — Vue d'ensemble de l'architecture des plugins
- `docs/plugins/contract.md` — Contrat d'interface plugin (protobuf + REST)
- `docs/plugins/discovery.md` — Mécanismes de découverte et d'enregistrement
- `docs/plugins/development.md` — Guide de développement d'un plugin
- `docs/plugins/examples.md` — Exemples de plugins
- La documentation couvre l'architecture et non l'implémentation détaillée
- Chaque document inclut des diagrammes Mermaid pour les flux

**Cas limites :**
- La documentation doit être compréhensible sans lire le code source
- Les exemples de code dans la documentation doivent être vérifiés contre l'implémentation réelle

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/plugins/architecture.md` | Créer | Architecture complète des plugins |
| `docs/plugins/contract.md` | Créer | Contrat d'interface plugin |
| `docs/plugins/discovery.md` | Créer | Mécanisme de découverte |
| `docs/plugins/development.md` | Créer | Guide de développement |
| `docs/plugins/examples.md` | Créer | Exemples de plugins |
| `docs/cli/commands.md` | Modifier | Ajouter `ade plugin` |

### Contenu détaillé des documents

#### `docs/plugins/architecture.md`
- **Vue d'ensemble** : Description du système de plugins Docker, rôle des plugins, de l'orchestrateur et du CLI
- **Diagramme d'architecture** (Mermaid) :
  ```mermaid
  flowchart LR
      CLI[ade CLI] -->|REST+gRPC\nfallback| ORCH[Orchestrateur:8080]
      ORCH -->|proxy| REG[Registry:8082]
      CLI -->|ade plugin install| DOCKER[docker]
      REG -->|découverte pull| SIDE[Docker Proxy:8083]
      SIDE -->|lecture labels| DOCKERD[Docker daemon]
      REG -->|REST/gRPC| P1[Plugin Templates:8081]
      REG -->|REST/gRPC| P2[Plugin GitHub]
      P1 -.->|push register| REG
  ```
- **Composants** :
  - **Registry (port 8082)** : Service autonome dans le conteneur orchestrateur, gère l'enregistrement et le suivi des plugins
  - **Docker Proxy (port 8083)** : Side-car avec accès Docker socket en lecture seule, expose une API REST limitée pour lister les conteneurs par labels
  - **Orchestrateur (port 8080)** : Proxyfie les appels `/api/v1/plugins/*` vers le registry
- **Flux d'enregistrement** : (1) Push : le plugin démarre et s'enregistre via `POST /api/v1/plugins/register` sur le registry → (2) Pull : le registry interroge périodiquement le Docker proxy pour détecter les nouveaux conteneurs avec labels `ade.plugin.*` → (3) Health check périodique toutes les 30s
- **Protocoles supportés** : REST/HTTP pour les opérations simples, gRPC pour les échanges structurés
- **Communication CLI** : Le CLI essaie d'abord gRPC (port 9090), fallback REST (port 8080) si gRPC indisponible
- **Dégradation** : CLI et orchestrateur restent fonctionnels sans plugins. Orchestrateur indisponible → messages informatifs au lieu d'erreurs fatales.
- **Installation** : `ade plugin install <image>` pour déploiement à la volée, `docker-compose.yml` pour plugins permanents

#### `docs/plugins/contract.md`
- **Définition protobuf** : Lien vers `api/grpc/plugin.proto` et explication des messages
- **Endpoints REST** : Tableau des endpoints obligatoires et optionnels
- **Structures de données** : `PluginDescriptor`, `Capability`, `HealthStatus`
- **Versionnement** : `api_version` et compatibilité entre versions

#### `docs/plugins/discovery.md`
- **Mécanisme push** : Le plugin s'enregistre via `POST /api/v1/plugins/register`
- **Mécanisme pull** : L'orchestrateur scanne les conteneurs Docker avec labels `ade.plugin.*`
- **Labels Docker supportés** : Tableau des labels et leur signification
- **Health check** : Périodicité, timeout, nombre d'échecs maximum
- **Diagramme de séquence** (Mermaid) du cycle de vie d'un plugin

#### `docs/plugins/development.md`
- **Prérequis** : Go 1.26, Docker, compréhension de base de REST et gRPC, `buf` installé pour la compilation protobuf
- **SDK** : Deux modes d'accès :
  - Pour les plugins du dépôt first-party : import `automated_dev_environment/internal/plugins/sdk`
  - Pour les plugins externes : utiliser le module SDK publié séparément `plugins/sdk` (avec `go get`)
- **Étapes de création d'un plugin** :
  1. Initialiser un module Go (`go mod init`)
  2. Importer le SDK
  3. Créer le `PluginServer` avec `sdk.NewPlugin()` et `net/http` standard pour les handlers REST
  4. Ajouter des capacités avec `AddCapability()`
  5. Enregistrer les handlers REST avec `HandleFunc()` et les services gRPC avec `RegisterService()`
  6. Démarrer le serveur avec `Start(ctx)` (gère SIGTERM/SIGINT)
  7. Créer un Dockerfile multi-stage avec labels `ade.plugin.*`
  8. Tester avec Docker Compose (registry + docker-proxy + plugin)
- **Exemple de code complet** : Plugin "hello-world" minimal
- **Contrat protobuf** : Défini dans `api/grpc/plugin.proto`, généré avec `buf generate`

#### `docs/plugins/examples.md`
- **Plugin Templates** (détaillé) : Description du plugin, endpoints, utilisation
- **Idées de plugins** : GitHub (issues, PRs), Slack (notifications), Database (migrations), CI (Jenkins/GitHub Actions)
- **Patterns avancés** : Configuration de plugins via l'interface web, agrégation de données entre plugins

### Documentation
- Cette tâche produit la documentation elle-même — pas de code supplémentaire

### Exemples de contenu

Exemple minimal de création de plugin (extrait du guide de développement) :

```go
    server, _ := sdk.NewPlugin(
        sdk.WithConfig(sdk.Config{
            Name:            "hello-world",
            Version:         "0.1.0",
            Description:     "Mon premier plugin",
            OrchestratorURL: "http://registry:8082", // port du service registry
        }),
    )
```

```yaml
services:
  hello-plugin:
    build: ./plugins/hello-world
    environment:
      PLUGIN_NAME: hello-world
      PLUGIN_ORCHESTRATOR_URL: "http://registry:8082"
      PLUGIN_HTTP_PORT: "8081"
    labels:
      ade.plugin.name: "hello-world"
      ade.plugin.version: "0.1.0"
    networks:
      - ade-network
```


