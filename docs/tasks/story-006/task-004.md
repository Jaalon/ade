# Tâche #004 - Story #006 : Documentation du système de validation

## Objectif
Créer la documentation complète pour le système de validation modulaire : architecture, création de modules, et formats de rapport.

## Contexte
- Story #006 : `docs/stories/story-006.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003 (toutes les implémentations terminées)
- Nécessaire pour : Aucune (documentation finale)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer trois fichiers de documentation dans `docs/validation/` :

1. **`docs/validation/architecture.md`** — Architecture du système de validation modulaire
2. **`docs/validation/modules.md`** — Guide de création de nouveaux modules de validation
3. **`docs/validation/report.md`** — Formats des rapports (JSON, JUnit XML) et visualisation web

**Cas nominaux :**
- Les trois fichiers sont en Markdown, en français, bien structurés
- `architecture.md` documente l'interface Validator, le registre, le runner, les types de résultats, la configuration
- `modules.md` explique comment créer un nouveau module de validation avec un exemple complet
- `report.md` décrit les formats JSON et JUnit XML, et la visualisation web via l'orchestrateur

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/validation/architecture.md` | Créer | Documentation de l'architecture |
| `docs/validation/modules.md` | Créer | Guide de création de modules |
| `docs/validation/report.md` | Créer | Formats de rapport et visualisation web |

### Contenu attendu

#### `docs/validation/architecture.md`

```markdown
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
```

#### `docs/validation/modules.md`

```markdown
# Création de modules de validation

## Principe
Un module de validation est un type Go qui implémente l'interface
`Validator`. Il peut être créé dans le package `internal/validation/`
ou dans un package externe.

## Structure d'un module

```go
package validation

import "context"

// Exemple : Validateur PostgreSQL
type PostgresValidator struct{}

func NewPostgresValidator() *PostgresValidator {
    return &PostgresValidator{}
}

func (v *PostgresValidator) Name() string {
    return "postgres"
}

func (v *PostgresValidator) Description() string {
    return "Validation de l'environnement PostgreSQL"
}

func (v *PostgresValidator) Detect(ctx context.Context) (bool, error) {
    // Vérifier si psql est disponible
    _, err := exec.LookPath("psql")
    return err == nil, nil
}

func (v *PostgresValidator) Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
    var checks []CheckResult

    // Check 1 : psql disponible
    checks = append(checks, v.checkPsqlAvailable(ctx))

    // Check 2 : connexion à la base
    checks = append(checks, v.checkConnection(ctx))

    return &ModuleResult{
        ModuleName: v.Name(),
        Status:     aggregateStatus(checks),
        Checks:     checks,
    }, nil
}
```

## Enregistrement

```go
func init() {
    Register(NewPostgresValidator())
}
```

## Points clés

1. **Nom unique** : `Name()` doit retourner un identifiant unique
2. **Détection pertinente** : `Detect()` ne doit pas être agressive
   (préférer faux négatif que faux positif)
3. **Checks granulaires** : Chaque vérification atomique est un `CheckResult`
4. **Messages en français** : Les messages utilisateur sont en français
5. **Gestion d'erreurs** : Ne pas paniquer, retourner des erreurs proprement
6. **Timeouts** : Chaque check doit avoir un timeout raisonnable

## Tests

```go
func TestPostgresValidator_Detect(t *testing.T) {
    v := NewPostgresValidator()
    assert.Equal(t, "postgres", v.Name())
}

func TestPostgresValidator_Validate(t *testing.T) {
    v := NewPostgresValidator()
    result, err := v.Validate(context.Background(), ModuleConfig{Name: "postgres"})
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Checks)
}
```
```

#### `docs/validation/report.md`

```markdown
# Formats de rapport de validation

## Rapport JSON

Le rapport JSON est généré par `JSONReporter` et contient la structure
complète des résultats de validation.

### Exemple

```json
{
  "status": "failed",
  "duration": "4.23s",
  "started_at": "2026-05-24T10:30:00Z",
  "completed_at": "2026-05-24T10:30:04Z",
  "num_checks": 4,
  "num_passed": 3,
  "num_failed": 1,
  "modules": [
    {
      "module_name": "golang",
      "status": "failed",
      "duration": "3.84s",
      "checks": [
        {
          "name": "go-version",
          "status": "passed",
          "message": "Go 1.26 trouvé",
          "duration": "0.34s"
        },
        {
          "name": "go-build",
          "status": "passed",
          "message": "Build réussi",
          "duration": "1.23s"
        },
        {
          "name": "go-test",
          "status": "failed",
          "message": "2 tests échoués sur 15",
          "duration": "2.10s",
          "details": "--- FAIL: TestSomething (0.01s)\n    something_test.go:42: expected true, got false"
        },
        {
          "name": "go-vet",
          "status": "passed",
          "message": "Vet réussi",
          "duration": "0.17s"
        }
      ]
    }
  ]
}
```

## Rapport JUnit XML

Le rapport JUnit XML est généré par `JUnitReporter`. Il est compatible
avec les outils CI standards (Jenkins, GitLab CI, GitHub Actions).

### Exemple

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="ade-validate" tests="4" failures="1" errors="0" time="4.23">
  <testsuite name="golang" tests="4" failures="1" errors="0" skipped="0" time="3.84">
    <testcase name="go-version" classname="golang" time="0.34" />
    <testcase name="go-build" classname="golang" time="1.23" />
    <testcase name="go-test" classname="golang" time="2.10">
      <failure message="2 tests échoués sur 15" type="failure">--- FAIL: TestSomething (0.01s)
    something_test.go:42: expected true, got false</failure>
    </testcase>
    <testcase name="go-vet" classname="golang" time="0.17" />
  </testsuite>
</testsuites>
```

## Visualisation web (Story #008)

> ⚠ La visualisation web des rapports de validation sera disponible
> dans Story #008 via l'interface web de l'orchestrateur.

### Principe (vision future)

1. `ade validate run` génère les rapports sur le disque
2. L'orchestrateur (conteneur `ade-config`) expose une API REST
3. Les rapports sont envoyés à l'orchestrateur via `POST /api/v1/reports`
4. L'interface web affiche les rapports avec historique et tendances

### Format pour l'orchestrateur

En V1, les rapports sont sauvegardés dans un format JSON enrichi,
prêt à être consommé par l'orchestrateur :

```json
{
  "format": "ade-validation-report-v1",
  "generated_at": "2026-05-24T10:30:04Z",
  "project_name": "mon-projet",
  "report": { ... }
}
```

### État actuel

| Fonctionnalité | Statut |
|----------------|--------|
| Rapport JSON | ✅ Disponible |
| Rapport JUnit XML | ✅ Disponible |
| Sauvegarde disque | ✅ Disponible |
| Envoi orchestrateur | 📅 Story #008 |
| Visualisation web | 📅 Story #008 |
| Historique tendances | 📅 Story #008 |
```

### Contraintes techniques
- **Format** : Markdown, encodé en UTF-8
- **Langue** : Français
- **Style** : Clair, concis, avec des tableaux pour les références rapides
- **Liens** : Utiliser des liens relatifs entre fichiers de documentation
- **Exemples** : Les exemples Go et YAML doivent être valides et correspondre à l'implémentation réelle
- **Cohérence** : Vérifier que les noms de commandes utilisent `ade validate` (pas `ade check`)
- **Mentions "vision future"** : Bien indiquer ce qui est disponible maintenant vs. ce qui viendra plus tard (Story #008)
- **Aucun test automatisé** pour la documentation. Vérification manuelle recommandée.

### Documentation
Cette tâche **est** la documentation. Aucune documentation supplémentaire.

### Exemples d'utilisation
```bash
# Afficher la documentation
ade validate --help

# Lire la documentation
type docs\validation\architecture.md
type docs\validation\modules.md
type docs\validation\report.md
```
