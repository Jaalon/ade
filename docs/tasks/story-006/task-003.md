# Tâche #003 - Story #006 : Commande CLI `ade validate` et intégration pipeline

## Objectif
Implémenter la commande `ade validate` avec ses sous-commandes (`run`, `init`), l'intégration des rapports dans le pipeline CI, et le point d'intégration avec l'orchestrateur (Story #008 - stub).

## Contexte
- Story #006 : `docs/stories/story-006.md`
- Dépend de : Tâche #001 (package `internal/validation/`), Tâche #002 (GoValidator, reporters)
- Dépend de : Story #005 Tâche #001 (package `internal/ci/`, PipelineRunner)
- Nécessaire pour : Tâche #004 (documentation)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Trois composants à implémenter :

#### 1. Commande `ade validate`
Commande Cobra avec deux sous-commandes :

- **`ade validate run`** — Exécute les modules de validation :
  - Charge la configuration depuis `ade-validate.yaml` (ou fichier spécifié par `--config`)
  - Détecte les modules applicables au projet
  - Exécute chaque module et collecte les résultats
  - Génère les rapports dans les formats configurés (JSON, JUnit XML)
  - Affiche un résumé formaté dans la console
  - Sauvegarde les rapports dans le répertoire de sortie (`.ade/validation/` par défaut)

- **`ade validate init`** — Génère un fichier de configuration par défaut :
  - Crée `ade-validate.yaml` dans le répertoire courant
  - Template avec tous les modules disponibles commentés
  - N'écrase pas un fichier existant sans `--force`

**Flags pour `ade validate run` :**
| Flag | Court | Défaut | Description |
|------|-------|--------|-------------|
| `--config` | `-c` | `ade-validate.yaml` | Chemin du fichier de configuration |
| `--output` | `-o` | `.ade/validation` | Répertoire de sortie des rapports |
| `--format` | `-f` | `json` | Format de rapport (json, junit, ou les deux) |
| `--verbose` | `-v` | `false` | Afficher les détails de chaque check |

**Flags pour `ade validate init` :**
| Flag | Court | Défaut | Description |
|------|-------|--------|-------------|
| `--output` | `-o` | `.` | Répertoire de sortie |
| `--force` | `-f` | `false` | Écraser sans confirmation |

#### 2. Intégration pipeline CI
- Ajouter un stage `validate` au pipeline CI (dans `internal/ci/types.go`)
- Le stage `validate` est positionné entre `build` et `unit-test`
- Ajouter les descriptions, l'ordre et la configuration YAML pour ce nouveau stage
- Le stage `validate` peut être activé/désactivé via la configuration du pipeline

#### 3. Point d'intégration orchestrateur (stub)
- Créer une fonction `StoreReport(ctx, report)` qui prépare l'envoi du rapport à l'orchestrateur
- Pour V1 : stub qui écrit le rapport dans un fichier JSON dans un format prêt à être consommé par l'orchestrateur
- Un message log indique que l'intégration orchestrateur sera disponible dans Story #008
- Le rapport est stocké dans le répertoire de l'orchestrateur si disponible (détection via Docker)

**Cas nominaux :**
- `ade validate run` : charge la config, détecte les modules, exécute, génère les rapports
- `ade validate run --verbose` : idem + affichage détaillé des checks
- `ade validate run --format junit` : ne génère que le rapport JUnit XML
- `ade validate run --format json,junit` : génère les deux formats
- `ade validate run --output ./reports` : écrit les rapports dans `./reports/`
- `ade validate init` : crée `ade-validate.yaml` avec les modules disponibles commentés
- `ade validate init --force` : écrase le fichier existant
- Pipeline avec stage `validate` activé : le stage s'exécute entre `build` et `unit-test`
- Tous les stages du pipeline existant sont réordonnés automatiquement si nécessaire

**Cas limites :**
- Aucun module détecté pour le projet → message "Aucun module de validation détecté"
- Répertoire de sortie inexistant → créé automatiquement
- Fichier de config inexistant → `DefaultConfig()` chargée, tous les modules détectés sont exécutés
- Orchestrateur non disponible → stub log, rapport sauvegardé sur disque
- Pipeline avec `validate` désactivé → stage skipped, comportement normal
- `ade validate init` avec fichier existant et sans `--force` → message d'erreur

**Gestion d'erreurs :**
- Config invalide → afficher les erreurs, proposer d'utiliser la config par défaut
- Module échoué → afficher le détail, continuer les autres modules (non fatal)
- Tous les modules échouent → statut global "failed"
- Écriture des rapports impossible → avertir mais ne pas blocker (le résumé console a été affiché)
- Template inconnu pour `init` → erreur (réutiliser le pattern de `pipeline_init.go`)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/validate.go` | Créer | Commande `ade validate` racine |
| `internal/command/validate_run.go` | Créer | Sous-commande `ade validate run` |
| `internal/command/validate_init.go` | Créer | Sous-commande `ade validate init` |
| `internal/command/validate_test.go` | Créer | Tests des commandes validate |
| `internal/command/root.go` | Modifier | Ajouter `validateCmd` dans `init()` |
| `internal/ci/types.go` | Modifier | Ajouter `StageValidate` aux stages |
| `internal/ci/config.go` | Modifier | Mettre à jour DefaultConfig et ordre des stages |
| `internal/validation/orchestrator.go` | Créer | Point d'intégration avec l'orchestrateur (stub V1) |

### Signatures

```go
// internal/command/validate.go
package command

import "github.com/spf13/cobra"

var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Valide l'environnement de développement",
    Long:  `Valide l'environnement de développement avec des modules de validation spécifiques.`,
}

func init() {
    validateCmd.AddCommand(validateRunCmd)
    validateCmd.AddCommand(validateInitCmd)
    rootCmd.AddCommand(validateCmd)
}
```

```go
// internal/command/validate_run.go
package command

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "automated_dev_environment/internal/validation"
)

// Variables mockables
var (
    valLoadConfigFn      = validation.LoadConfig
    valNewRunnerFn       = func() *validation.ValidationRunner { return validation.NewValidationRunner() }
    valDetectModulesFn   = validation.DetectModules
    valNewReportWriterFn = validation.NewReportWriter
    valStoreReportFn     = StoreValidationReport
)

var (
    valRunConfig  string
    valRunOutput  string
    valRunFormat  string
    valRunVerbose bool
)

var validateRunCmd = &cobra.Command{
    Use:   "run",
    Short: "Exécute les modules de validation",
    Long:  `Exécute les modules de validation de l'environnement et génère les rapports.`,
    RunE:  runValidate,
}

type validateRunOptions struct {
    ConfigPath string
    OutputDir  string
    Formats    []string
    Verbose    bool
}

func runValidate(cmd *cobra.Command, args []string) error
func loadValidateConfig(path string) (validation.ValidationConfig, error)
func displayValidationResult(out io.Writer, report *validation.ValidationReport, verbose bool)
func writeValidationReports(report *validation.ValidationReport, opts validateRunOptions) []error
func ensureOutputDir(path string) error
func parseFormats(formatFlag string) []string
```

```go
// internal/command/validate_init.go
package command

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
)

var (
    valInitOutput string
    valInitForce  bool
)

var validateInitCmd = &cobra.Command{
    Use:   "init",
    Short: "Génère la configuration par défaut de validation",
    Long:  `Génère un fichier ade-validate.yaml avec les modules disponibles.`,
    RunE:  runValidateInit,
}

func runValidateInit(cmd *cobra.Command, args []string) error

// validateTemplateContent retourne le YAML par défaut pour la validation
const validateTemplateYAML = `# Configuration de validation de l'environnement
# Généré par 'ade validate init'
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
`
```

```go
// internal/validation/orchestrator.go
package validation

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

// StoreReport saves the validation report for the orchestrator.
// V1: writes to a well-known location on disk.
// Future: will send to orchestrator REST API when available (Story #008).
func StoreReport(ctx context.Context, report *ValidationReport, outputDir string) error

// orchestratorReportPath returns the path where the orchestrator expects reports.
func orchestratorReportPath(outputDir string) string {
    return filepath.Join(outputDir, "orchestrator-report.json")
}

// writeOrchestratorFormat writes the report in a format consumable by the orchestrator.
// This is a superset of the JSON report with additional metadata.
func writeOrchestratorFormat(report *ValidationReport, path string) error
```

```go
// internal/ci/types.go — Modification : Ajouter StageValidate
const (
    StageValidate        StageType = "validate"
    StageBuild           StageType = "build"
    StageUnitTest        StageType = "unit-test"
    // ... existing stages ...
)

// AllStages met à jour pour inclure validate en 2e position
func AllStages() []StageType {
    return []StageType{
        StageBuild,
        StageValidate,      // <-- NOUVEAU : après build, avant unit-test
        StageUnitTest,
        StageIntegrationTest,
        StageTestDeploy,
        StageE2E,
        StagePreprod,
    }
}
```

Ajouter dans les descriptions :
```go
stageDescriptions[StageValidate] = "Validation de l'environnement"
```

Ajouter dans `DefaultConfig()` : un step `validate` activé par défaut avec les commandes :
```go
{
    Type:    StageValidate,
    Name:    "Validation de l'environnement",
    Enabled: true,
    Steps: []StepConfig{
        {Name: "Validation modulaire", Command: []string{"ade", "validate", "run", "--format", "junit"}},
    },
},
```

### Contraintes techniques
- **Pattern Cobra** : Suivre le même pattern que `init.go`, `pipeline.go` (variables mockable, `init()` pour enregistrement, messages en français).
- **Enregistrement** : Ajouter `validateCmd` à `rootCmd` dans `root.go` et `validateRunCmd`, `validateInitCmd` à `validateCmd` dans `validate.go`.
- **Flags pour `ade validate run`** :
  - `--config` / `-c` (string, défaut `"ade-validate.yaml"`)
  - `--output` / `-o` (string, défaut `".ade/validation"`)
  - `--format` / `-f` (string, défaut `"json"`)
  - `--verbose` / `-v` (bool, défaut `false`)
- **Flags pour `ade validate init`** :
  - `--output` / `-o` (string, défaut `"."`)
  - `--force` / `-f` (bool, défaut `false`)
- **Écran de résumé** : Afficher un tableau formaté similaire à celui du pipeline :
  ```
  ✓ Validation terminée (3 modules, 8 checks, 7 passed, 1 failed)

    Module        Statut    Checks     Durée
    golang        passed    4/4        3.24s
    quarkus       skipped   0/0        0.00s

  Rapports générés :
    ✓ .ade/validation/report.json
    ✓ .ade/validation/report.xml
  ```
- **Pipeline stage** : Le nouveau `StageValidate` est inséré entre `build` et `unit-test`. Cela réordonne automatiquement les stages dans le pipeline. Les configurations YAML existantes avec l'ancien ordre seront automatiquement réordonnées par `ValidateConfig`.
- **Backward compatibility** : L'ajout de `StageValidate` ne doit pas casser les configs existantes. `DefaultConfig()` inclut le nouveau stage activé. Les configs YAML sans `validate` auront le stage skipped.
- **Orchestrator stub** : Simple écriture de fichier. Un log `fmt.Fprintf(out, "  ∼ Intégration orchestrateur : disponible dans Story #008\n")` est affiché.
- **Options de format** : Le flag `--format` accepte `json`, `junit`, ou `json,junit` (séparé par virgule).
- **Windows** : Tous les chemins utilisent `filepath.Join`. Le répertoire `.ade/validation` est créé si nécessaire.
- **Messages en français** : Tous les messages utilisateur sont en français.

### Tests à implémenter

#### Tests unitaires — `internal/command/validate_test.go`

- **Scénario 1 : `TestValidateCmd_Registered`** — `ade validate --help` affiche l'aide
- **Scénario 2 : `TestValidateRunCmd_Default`** — `ade validate run` avec config par défaut
- **Scénario 3 : `TestValidateRunCmd_WithConfigFlag`** — `--config ./custom.yaml`
- **Scénario 4 : `TestValidateRunCmd_VerboseOutput`** — `--verbose` affiche les détails
- **Scénario 5 : `TestValidateRunCmd_FormatJSON`** — `--format json` → rapport JSON seulement
- **Scénario 6 : `TestValidateRunCmd_FormatJUnit`** — `--format junit` → rapport JUnit seulement
- **Scénario 7 : `TestValidateRunCmd_FormatBoth`** — `--format json,junit` → les deux rapports
- **Scénario 8 : `TestValidateRunCmd_CustomOutput`** — `--output ./reports` → rapport dans ./reports/
- **Scénario 9 : `TestValidateRunCmd_OutputDirCreated`** — répertoire créé automatiquement
- **Scénario 10 : `TestValidateRunCmd_DisplayResults`** — tableau formaté correct
- **Scénario 11 : `TestValidateInitCmd_CreateFile`** — `ade validate init` crée ade-validate.yaml
- **Scénario 12 : `TestValidateInitCmd_ForceOverwrite`** — `--force` écrase
- **Scénario 13 : `TestValidateInitCmd_NoOverwriteWithoutForce`** — pas d'écrasement sans --force
- **Scénario 14 : `TestValidateInitCmd_CustomOutput`** — `--output ./custom/` crée au bon endroit
- **Scénario 15 : `TestParseFormats_Single`** — parseFormats("json") → ["json"]
- **Scénario 16 : `TestParseFormats_Multiple`** — parseFormats("json,junit") → ["json", "junit"]
- **Scénario 17 : `TestParseFormats_Empty`** — parseFormats("") → ["json"]

#### Tests unitaires — `internal/ci/ci_test.go` (ajouts)

- **Scénario 18 : `TestAllStages_IncludesValidate`** — AllStages() contient StageValidate
- **Scénario 19 : `TestStageOrder_ValidatePosition`** — StageValidate après build, avant unit-test
- **Scénario 20 : `TestDefaultConfig_IncludesValidate`** — DefaultConfig() a stage validate activé

#### Tests unitaires — `internal/validation/validation_test.go` (ajouts)

- **Scénario 21 : `TestStoreReport_CreatesFile`** — StoreReport crée le fichier de rapport
- **Scénario 22 : `TestStoreReport_ReadableByOrchestrator`** — le format est un JSON valide avec les champs attendus

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #004).

### Exemples d'utilisation

```bash
# Exécution de la validation
ade validate run

# Avec détails
ade validate run --verbose

# Formats spécifiques
ade validate run --format junit
ade validate run --format json,junit

# Répertoire personnalisé
ade validate run --output ./reports

# Configuration personnalisée
ade validate run --config ./ade-validate.yaml

# Génération de la configuration
ade validate init
ade validate init --force

# Dans le pipeline CI (ade-pipeline.yaml)
# La validation est déjà intégrée comme stage entre build et unit-test
```
