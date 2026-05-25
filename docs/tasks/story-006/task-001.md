# Tâche #001 - Story #006 : Infrastructure de validation modulaire

## Objectif
Créer le package `internal/validation/` avec les interfaces, types, registre de modules, configuration YAML et moteur d'exécution permettant la validation modulaire de l'environnement.

## Contexte
- Story #006 : `docs/stories/story-006.md`
- Plan d'implémentation : Phase 3, après Story #004 et #005
- Dépend de : Story #001 (conventions Go, patterns de package), Story #004 (docker-compose, déploiement), Story #005 (package `internal/ci/`)
- Nécessaire pour : Tâche #002 (modules de validation + rapports), Tâche #003 (CLI + pipeline)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer le package `internal/validation/` qui constitue le cœur du système de validation modulaire. Ce package définit :

1. **Interface `Validator`** : contrat que chaque module de validation doit implémenter
2. **Types de résultats** : `CheckResult`, `ModuleResult`, `ValidationReport`
3. **Registre de modules** : système de découverte et d'enregistrement des validateurs
4. **Configuration YAML** : chargement et validation de la configuration de validation
5. **Moteur d'exécution** : `ValidationRunner` qui orchestre l'exécution des validateurs

**Cas nominaux :**
- `Validator` interface définit `Name()`, `Description()`, `Detect()`, `Validate(ctx, config)`
- `CheckResult` contient le nom du check, statut, message, durée, erreur éventuelle
- `ModuleResult` agrège les `CheckResult` d'un module avec le nom du module, statut global, durée
- `ValidationReport` agrège tous les `ModuleResult` avec statut global, durée, date
- `Registry` permet d'enregistrer (`Register`) et de détecter (`DetectModules`) les validateurs
- `ValidationRunner.Run(ctx, config)` exécute les modules sélectionnés et retourne un rapport
- La configuration se charge depuis un fichier YAML (`ade-validate.yaml`)

**Cas limites :**
- Aucun module enregistré → rapport vide avec statut "success" (rien à valider)
- Aucun module détecté pour le projet → message d'avertissement, rapport vide
- `ValidationRunner.Run` avec config vide → exécute tous les modules enregistrés et détectés
- Contexte annulé pendant l'exécution → arrêt immédiat, statut "cancelled" sur les modules non exécutés

**Gestion d'erreurs :**
- Module qui panique → catch avec `recover`, statut "failed" + message d'erreur
- `Detect()` d'un module qui échoue → le module est ignoré (non fatal)
- Configuration invalide → erreur de validation avec détails
- Fichier de config inexistant → `DefaultConfig()` chargée (tous les modules détectés activés)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/validation/types.go` | Créer | Types fondamentaux : CheckStatus, CheckResult, ModuleResult, ValidationReport |
| `internal/validation/validator.go` | Créer | Interface Validator, Registry, Register, DetectModules, Modules |
| `internal/validation/runner.go` | Créer | ValidationRunner, Run, execution engine |
| `internal/validation/config.go` | Créer | ValidationConfig, DefaultConfig, LoadConfig, ValidateConfig |
| `internal/validation/errors.go` | Créer | Erreurs sentinelles du package |
| `internal/validation/validation_test.go` | Créer | Tests unitaires |

### Signatures

```go
// internal/validation/types.go
package validation

import "time"

// CheckStatus represents the status of a single validation check.
type CheckStatus string

const (
    StatusPassed  CheckStatus = "passed"
    StatusFailed  CheckStatus = "failed"
    StatusWarning CheckStatus = "warning"
    StatusSkipped CheckStatus = "skipped"
    StatusError   CheckStatus = "error"
)

// CheckResult holds the result of a single validation check.
type CheckResult struct {
    Name      string        `json:"name"`
    Status    CheckStatus   `json:"status"`
    Message   string        `json:"message"`
    Duration  time.Duration `json:"duration"`
    Err       error         `json:"-"`
    Details   string        `json:"details,omitempty"`
}

// ModuleResult aggregates CheckResults for one validation module.
type ModuleResult struct {
    ModuleName string        `json:"module_name"`
    Status     CheckStatus   `json:"status"`
    Checks     []CheckResult `json:"checks"`
    Duration   time.Duration `json:"duration"`
    StartedAt  time.Time     `json:"started_at"`
    Error      error         `json:"-"`
}

// ValidationReport is the top-level result of a validation run.
type ValidationReport struct {
    Status      CheckStatus    `json:"status"`
    Modules     []ModuleResult `json:"modules"`
    Duration    time.Duration  `json:"duration"`
    StartedAt   time.Time      `json:"started_at"`
    CompletedAt time.Time      `json:"completed_at"`
    Config      ValidationConfig `json:"-"`
}

func (r *ValidationReport) Passed() bool
func (r *ValidationReport) Failed() bool
func (r *ValidationReport) HasWarnings() bool
// NumChecks returns total number of checks across all modules
func (r *ValidationReport) NumChecks() int
// NumPassed returns total number of passed checks
func (r *ValidationReport) NumPassed() int
// NumFailed returns total number of failed checks
func (r *ValidationReport) NumFailed() int
```

```go
// internal/validation/validator.go
package validation

import "context"

// Validator is the interface that all validation modules must implement.
type Validator interface {
    // Name returns the unique identifier for this validator (e.g. "golang").
    Name() string
    // Description returns a human-readable French description.
    Description() string
    // Detect checks whether this validator applies to the current project.
    // Returns true if the project environment matches this validator.
    Detect(ctx context.Context) (bool, error)
    // Validate runs all checks for this validator and returns the module result.
    Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error)
}

// Registry holds all registered validators.
var registry []Validator

// Register adds a validator to the global registry.
// Called by each validator's init() function.
func Register(v Validator)

// Modules returns a copy of all registered validators.
func Modules() []Validator

// DetectModules runs Detect() on all registered validators and returns
// those that match the current project environment.
func DetectModules(ctx context.Context) ([]Validator, error)
```

```go
// internal/validation/config.go
package validation

// ValidationConfig is the top-level configuration for a validation run.
type ValidationConfig struct {
    // Modules is the list of module-specific configurations.
    Modules []ModuleConfig `yaml:"modules" json:"modules"`
    // OutputDir is where report files are written.
    OutputDir string `yaml:"output_dir,omitempty" json:"output_dir,omitempty"`
    // Formats is the list of output formats (json, junit).
    Formats []string `yaml:"formats,omitempty" json:"formats,omitempty"`
}

// ModuleConfig configures a single validation module.
type ModuleConfig struct {
    // Name matches Validator.Name(). Empty means auto-detect.
    Name string `yaml:"name" json:"name"`
    // Enabled allows disabling a module without removing it.
    Enabled bool `yaml:"enabled" json:"enabled"`
    // Options are module-specific key-value settings.
    Options map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
    // Checks is an optional list of specific checks to run (module-dependent).
    // Empty means run all checks for this module.
    Checks []string `yaml:"checks,omitempty" json:"checks,omitempty"`
}

func DefaultConfig() ValidationConfig
func LoadConfig(path string) (ValidationConfig, error)
func ValidateConfig(cfg ValidationConfig) error
```

```go
// internal/validation/runner.go
package validation

import (
    "context"
    "time"
)

// ValidationRunner orchestrates the execution of validation modules.
type ValidationRunner struct {
    // AllowList restricts which modules can run. Empty means all.
    AllowList []string
}

func NewValidationRunner() *ValidationRunner

// Run executes all detected (or explicitly configured) validators
// and returns a complete ValidationReport.
func (r *ValidationRunner) Run(ctx context.Context, cfg ValidationConfig) (*ValidationReport, error)

// selectModules determines which modules to run based on config and detection.
func (r *ValidationRunner) selectModules(ctx context.Context, cfg ValidationConfig) ([]Validator, error)
```

```go
// internal/validation/errors.go
package validation

import "errors"

var (
    ErrConfigInvalid     = errors.New("configuration de validation invalide")
    ErrModuleNotFound    = errors.New("module de validation non trouvé")
    ErrModulePanic       = errors.New("le module a paniqué pendant l'exécution")
    ErrValidationFailed  = errors.New("la validation a échoué")
)
```

### Structure des données

Exemple de fichier de configuration `ade-validate.yaml` :

```yaml
# Configuration de validation de l'environnement
# Généré par 'ade validate init'
output_dir: .ade/validation
formats:
  - json
  - junit

modules:
  - name: golang
    enabled: true
    checks:
      - build
      - test
    options:
      go_version: "1.26"
      test_flags: "-count=1"

  - name: quarkus
    enabled: false
    checks:
      - build
```

### Contraintes techniques
- **Package** : `internal/validation/` — nouveau package, pas de sous-packages dans cette tâche
- **Interfaces** : `Validator` est une interface Go publique dans le package. Les modules concrets (Go, Quarkus) seront dans des fichiers séparés du même package ou dans la Tâche #002.
- **Registre** : Pattern singleton avec `init()` pour l'enregistrement automatique. Le registre est une slice globale dans `validator.go`.
- **Goroutines** : Les modules sont exécutés séquentiellement. Pas de parallélisation en V1.
- **Context** : Vérifier `ctx.Done()` entre chaque module et chaque check pour supporter l'annulation.
- **Timeouts** : Chaque module a un timeout par défaut de 5 minutes. Configurable via `ModuleConfig.Options["timeout"]`.
- **Panic safety** : Chaque appel à `v.Validate()` est protégé par `defer recover()` pour éviter qu'un module instable ne fasse planter tout le runner.
- **Messages en français** : Tous les messages d'erreur, descriptions, et logs sont en français.
- **YAML** : Utiliser `gopkg.in/yaml.v3` déjà dans go.mod.
- **Nommage** : Les statuts sont en anglais (passed, failed, warning, skipped, error) pour la compatibilité avec les formats de rapport standards.
- **Pas de dépendance Cobra** : Le package `internal/validation/` ne doit pas importer Cobra.
- **Test coverage** : Coverage minimum 80% pour ce package.

### Tests à implémenter

#### Tests unitaires — `internal/validation/validation_test.go`

- **Scénario 1 : `TestCheckStatus_Values`** — Les constantes CheckStatus ont les bonnes valeurs
- **Scénario 2 : `TestValidationReport_Passed`** — Tous les modules passed → Passed() = true
- **Scénario 3 : `TestValidationReport_Failed`** — Un module failed → Failed() = true
- **Scénario 4 : `TestValidationReport_HasWarnings`** — Un module warning → HasWarnings() = true
- **Scénario 5 : `TestValidationReport_Counters`** — NumChecks(), NumPassed(), NumFailed() exacts
- **Scénario 6 : `TestRegisterAndModules`** — Register + Modules retourne les validateurs enregistrés
- **Scénario 7 : `TestRegisterAndDetect`** — Register + DetectModules avec un mock qui détecte true
- **Scénario 8 : `TestDefaultConfig`** — DefaultConfig() retourne une config valide avec OutputDir = ".ade/validation"
- **Scénario 9 : `TestValidateConfig_Valid`** — DefaultConfig() valide retourne nil
- **Scénario 10 : `TestValidateConfig_UnknownFormat`** — format inconnu → erreur
- **Scénario 11 : `TestLoadConfig_FileNotFound`** — retourne DefaultConfig()
- **Scénario 12 : `TestLoadConfig_ValidYAML`** — fichier YAML temporaire parsé correctement
- **Scénario 13 : `TestRunner_RunNoModules`** — aucun module enregistré → rapport vide, StatusPassed
- **Scénario 14 : `TestRunner_RunWithMockModule`** — mock Validator qui réussit → StatusPassed
- **Scénario 15 : `TestRunner_RunWithFailingModule`** — mock Validator qui échoue → StatusFailed
- **Scénario 16 : `TestRunner_RunCancelledContext`** — contexte annulé → Status "cancelled" pour modules restants
- **Scénario 17 : `TestRunner_RunPanickingModule`** — module qui panique → catch, StatusError
- **Scénario 18 : `TestRunner_ModuleFiltering`** — ModuleConfig avec Name spécifique → seul ce module tourne
- **Scénario 19 : `TestRunner_ModuleDisabled`** — ModuleConfig avec Enabled=false → module ignoré

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #004).

### Exemples d'utilisation

```go
// Enregistrement d'un validateur
func init() {
    validation.Register(NewGoValidator())
}

// Exécution de la validation
runner := validation.NewValidationRunner()
cfg := validation.DefaultConfig()
report, err := runner.Run(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Validation: %s (%d checks, %d passed, %d failed)\n",
    report.Status, report.NumChecks(), report.NumPassed(), report.NumFailed())

// Chargement depuis YAML
cfg, err := validation.LoadConfig("ade-validate.yaml")
```
