# Architecture du système de validation modulaire

## Vue d'ensemble

Le système de validation modulaire permet de valider l'environnement de
développement et de préproduction via des modules spécialisés (Go, Quarkus,
etc.). Chaque module implémente une interface commune et est découvert
automatiquement.

## Interface Validator

```go
type Validator interface {
    Name() string
    Description() string
    Detect(ctx context.Context) (bool, error)
    Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error)
}
```

Tout module de validation doit implémenter cette interface pour être
compatible avec le système.

## Types de résultats

| Type | Description |
|------|-------------|
| `CheckResult` | Résultat d'une vérification élémentaire |
| `ModuleResult` | Agrégation des checks d'un module |
| `ValidationReport` | Rapport complet de l'exécution |

### Statuts

| Statut | Signification |
|--------|---------------|
| `passed` | La vérification a réussi |
| `failed` | La vérification a échoué |
| `warning` | Succès avec avertissement |
| `skipped` | Vérification non exécutée |
| `error` | Erreur technique lors de la vérification |

## Registre de modules

Les modules s'enregistrent automatiquement via `init()` :

```go
func init() {
    validation.Register(NewGoValidator())
}
```

Le runner utilise le registre pour détecter les modules applicables au
projet courant via `DetectModules()`.

## Moteur d'exécution

`ValidationRunner` orchestre l'exécution :

1. Charger la configuration (YAML ou défaut)
2. Détecter les modules applicables
3. Filtrer selon la configuration (modules activés/désactivés)
4. Exécuter chaque module séquentiellement
5. Agréger les résultats dans un `ValidationReport`
6. Générer les rapports aux formats configurés

## Commandes CLI

```bash
# Exécuter la validation
ade validate run

# Exécuter avec détails
ade validate run --verbose

# Générer la configuration
ade validate init

# Formats de rapport
ade validate run --format json
ade validate run --format junit
ade validate run --format json,junit
```

## Configuration

Fichier `ade-validate.yaml` :

```yaml
output_dir: .ade/validation
formats:
  - json
  - junit

modules:
  - name: golang
    enabled: true
    checks:
      - version
      - build
      - test
      - vet
```

## Intégration pipeline

La validation est intégrée comme un stage du pipeline CI, entre `build`
et `unit-test` :

```yaml
stages:
  - type: build
    ...
  - type: validate
    name: "Validation de l'environnement"
    enabled: true
    steps:
      - name: "Validation modulaire"
        command: ["ade", "validate", "run", "--format", "junit"]
  - type: unit-test
    ...
```

## Intégration orchestrateur (Story #008)

> ⚠ L'intégration complète avec l'orchestrateur sera disponible dans
> Story #008. En V1, les rapports sont sauvegardés sur le disque dans
> un format prêt à être consommé par l'orchestrateur.
