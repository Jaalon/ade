# Architecture du pipeline CI local

## Vue d'ensemble
Le pipeline CI local est une architecture abstraite, agnostique du moteur
d'exécution. Il définit une interface `Pipeline` qui peut être implémentée
par différents exécuteurs (dry-run, local, ou futur moteur CI).

## Interface Pipeline

```go
type Pipeline interface {
    Run(ctx context.Context, config PipelineConfig) (*PipelineResult, error)
    Validate(config PipelineConfig) error
}
```

Toute implémentation doit respecter l'ordre des stages et retourner
des résultats standardisés.

## Stages du pipeline

Le pipeline se compose de 6 stages exécutés séquentiellement :

| Ordre | Stage | Description |
|-------|-------|-------------|
| 1 | `build` | Construction du projet (compilation Go, Java, etc.) |
| 2 | `unit-test` | Exécution des tests unitaires |
| 3 | `integration-test` | Exécution des tests d'intégration |
| 4 | `test-deploy` | Déploiement dans un environnement de test |
| 5 | `e2e` | Tests end-to-end |
| 6 | `preprod` | Déploiement en préproduction |

Si un stage échoue, les stages suivants sont ignorés (`skipped`).

## Exécuteurs disponibles

| Exécuteur | Description | Usage |
|-----------|-------------|-------|
| `DryRunExecutor` | Simule l'exécution sans effet réel | Défaut, tests, validation |
| `LocalExecutor` | Exécute les commandes sur la machine hôte | `ade pipeline run --local` |

> L'exécuteur Docker (`DockerStepExecutor`) sera disponible dans une version
> ultérieure (Story #007/008).

## Configuration

Le pipeline est configuré via un fichier YAML (`ade-pipeline.yaml`)

```yaml
stages:
  - type: build
    name: "Compilation"
    enabled: true
    steps:
      - name: "Compiler le projet"
        command: ["go", "build", "./..."]
```

## Commandes CLI

```bash
# Initialiser la configuration
ade pipeline init
ade pipeline init --template go
ade pipeline init --template java

# Exécuter le pipeline
ade pipeline run
ade pipeline run --local
ade pipeline run --verbose
ade pipeline run --dry-run
ade pipeline run --config ./mon-pipeline.yaml
```

## Extensibilité

Pour ajouter un nouvel exécuteur, implémenter l'interface `Executor` :

```go
type Executor interface {
    Execute(ctx context.Context, step StepConfig) (*StepResult, error)
}
```

Voir `docs/ci/plugins.md` pour l'extension via plugins Docker (vision future).
