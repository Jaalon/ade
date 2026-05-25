# Tâche #002 - Story #006 : Module de validation Go et génération de rapports

## Objectif
Implémenter le module de validation Go (vérification de `go build`, `go test`) et les générateurs de rapports aux formats JSON structuré et JUnit XML.

## Contexte
- Story #006 : `docs/stories/story-006.md`
- Dépend de : Tâche #001 (package `internal/validation/`, interfaces, types, registry)
- Nécessaire pour : Tâche #003 (CLI + pipeline)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Deux composants doivent être implémentés :

#### 1. Module de validation Go (`GoValidator`)
Validateur pour les projets Go qui vérifie la configuration et la santé de l'environnement Go :

- **`go-version`** : Vérifie que `go` est disponible et que la version correspond à une version minimale configurable
- **`go-build`** : Exécute `go build ./...` pour vérifier que le projet compile
- **`go-test`** : Exécute `go test ./...` pour vérifier que les tests passent
- **`go-vet`** : Exécute `go vet ./...` pour vérifier la qualité du code (optionnel, activé par défaut)

Le module détecte automatiquement sa pertinence en vérifiant la présence d'un fichier `go.mod` dans le répertoire de travail.

#### 2. Générateurs de rapports
Deux formats de sortie pour les rapports de validation, lisibles depuis `ValidationReport` :

- **Rapport JSON** : Fichier structuré avec tous les résultats, métadonnées, horodatage
- **Rapport JUnit XML** : Format standard JUnit XML compatible avec les outils CI (Jenkins, GitLab, etc.)

**Cas nominaux :**
- `GoValidator.Detect(ctx, config)` retourne `true` si `go.mod` existe
- `GoValidator.Validate(ctx, config)` exécute les checks configurés, retourne `ModuleResult`
- `GoValidator.Name()` retourne `"golang"`, `Description()` retourne une description en français
- `GoValidator` s'enregistre automatiquement via `init()` dans le registre
- `JSONReporter.Write(report, writer)` produit un JSON valide et bien formaté
- `JUnitReporter.Write(report, writer)` produit un XML JUnit valide (schéma standard)
- Les reporters transforment `ModuleResult` en `CheckResult` → JUnit test cases
- Les deux reporters gèrent la sortie vers un `io.Writer`

**Cas limites :**
- `go.mod` absent → `Detect()` retourne `false`
- `go` non installé → check `go-version` retourne `StatusFailed` avec message explicite
- `go build` échoue → check retourne `StatusFailed` avec la sortie d'erreur
- `go test` sans tests → check retourne `StatusWarning` (pas d'erreur, mais aucun test)
- Projet avec `go.mod` mais pas de code Go → `go build` peut réussir (package vide)
- `JUnitReporter` avec 0 checks → fichier XML valide avec 0 testcases

**Gestion d'erreurs :**
- `go` introuvable dans le PATH → message d'installation (téléchargement depuis go.dev)
- Commande shell qui timeout → `StatusFailed` avec message "timeout dépassé"
- Répertoire de travail non accessible → `StatusError` avec le détail de l'erreur
- Report vide (aucun module exécuté) → JSON avec status "passed", JUnit avec 0 tests

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/validation/golang.go` | Créer | GoValidator implementation |
| `internal/validation/report.go` | Créer | Report interfaces et fonctions communes |
| `internal/validation/report_json.go` | Créer | JSON report generator |
| `internal/validation/report_junit.go` | Créer | JUnit XML report generator |
| `internal/validation/validation_test.go` | Modifier | Ajouter les tests pour GoValidator et reporters |

### Signatures

```go
// internal/validation/golang.go
package validation

import (
    "context"
    "os/exec"
)

// GoValidator validates a Go development environment.
type GoValidator struct {
    // goCmd is the path to the go binary (default: "go")
    goCmd func() (string, error)
    // execCmd is mockable command execution
    execCmd func(ctx context.Context, name string, args ...string) *exec.Cmd
}

func NewGoValidator() *GoValidator

func (v *GoValidator) Name() string        // "golang"
func (v *GoValidator) Description() string  // "Validation de l'environnement Go"

// Detect checks for go.mod in the working directory.
func (v *GoValidator) Detect(ctx context.Context) (bool, error)

// Validate runs the configured Go checks.
// Supported checks (ModuleConfig.Checks): "version", "build", "test", "vet".
// If Checks is empty, all checks are run.
func (v *GoValidator) Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error)

// runCommand executes a command and returns stdout+stderr combined.
func (v *GoValidator) runCommand(ctx context.Context, name string, args ...string) (string, error)

// checkGoVersion runs "go version" and validates the version.
func (v *GoValidator) checkGoVersion(ctx context.Context) CheckResult

// checkGoBuild runs "go build ./..." in the project directory.
func (v *GoValidator) checkGoBuild(ctx context.Context) CheckResult

// checkGoTest runs "go test ./..." in the project directory.
func (v *GoValidator) checkGoTest(ctx context.Context) CheckResult

// checkGoVet runs "go vet ./..." in the project directory.
func (v *GoValidator) checkGoVet(ctx context.Context) CheckResult

// parseGoVersion parses "go version" output to extract the version string.
// Input: "go version go1.26 windows/amd64"
// Output: "1.26"
func parseGoVersion(output string) string

// minGoVersion returns the minimum Go version required.
// Configurable via ModuleConfig.Options["go_version"], default "1.21".
func (v *GoValidator) minGoVersion(cfg ModuleConfig) string
```

```go
// internal/validation/report.go
package validation

import "io"

// ReportWriter defines the interface for report output formats.
type ReportWriter interface {
    // Format returns the format name (e.g. "json", "junit").
    Format() string
    // Write writes the validation report to the given writer.
    Write(report *ValidationReport, w io.Writer) error
}

// AvailableReporters returns the registered report writers.
func AvailableReporters() []ReportWriter

// NewReportWriter creates a ReportWriter by format name.
// Supported: "json", "junit". Returns nil if unknown.
func NewReportWriter(format string) ReportWriter
```

```go
// internal/validation/report_json.go
package validation

type JSONReporter struct{}

func NewJSONReporter() *JSONReporter
func (r *JSONReporter) Format() string // "json"

// JSONReport is the serializable structure for JSON output.
type JSONReport struct {
    Status      string            `json:"status"`
    Duration    string            `json:"duration"`
    StartedAt   string            `json:"started_at"`
    CompletedAt string            `json:"completed_at"`
    NumChecks   int               `json:"num_checks"`
    NumPassed   int               `json:"num_passed"`
    NumFailed   int               `json:"num_failed"`
    Modules     []JSONModuleResult `json:"modules"`
}

type JSONModuleResult struct {
    ModuleName string        `json:"module_name"`
    Status     string        `json:"status"`
    Duration   string        `json:"duration"`
    Checks     []JSONCheckResult `json:"checks"`
}

type JSONCheckResult struct {
    Name     string `json:"name"`
    Status   string `json:"status"`
    Message  string `json:"message"`
    Duration string `json:"duration"`
    Details  string `json:"details,omitempty"`
}
```

```go
// internal/validation/report_junit.go
package validation

// JUnitReporter generates JUnit XML format reports.
type JUnitReporter struct{}

func NewJUnitReporter() *JUnitReporter
func (r *JUnitReporter) Format() string // "junit"

// JUnitTestSuites is the root XML element for JUnit output.
type JUnitTestSuites struct {
    XMLName    struct{}         `xml:"testsuites"`
    Name       string           `xml:"name,attr"`
    Tests      int              `xml:"tests,attr"`
    Failures   int              `xml:"failures,attr"`
    Errors     int              `xml:"errors,attr"`
    Time       string           `xml:"time,attr"`
    TestSuites []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
    XMLName   struct{}      `xml:"testsuite"`
    Name      string        `xml:"name,attr"`
    Tests     int           `xml:"tests,attr"`
    Failures  int           `xml:"failures,attr"`
    Errors    int           `xml:"errors,attr"`
    Skipped   int           `xml:"skipped,attr"`
    Time      string        `xml:"time,attr"`
    TestCases []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
    XMLName   struct{}       `xml:"testcase"`
    Name      string         `xml:"name,attr"`
    Classname string         `xml:"classname,attr"`
    Time      string         `xml:"time,attr"`
    Failure   *JUnitFailure  `xml:"failure,omitempty"`
    Error     *JUnitError    `xml:"error,omitempty"`
    Skipped   *JUnitSkipped  `xml:"skipped,omitempty"`
}

type JUnitFailure struct {
    Message string `xml:"message,attr"`
    Type    string `xml:"type,attr"`
    Content string `xml:",chardata"`
}

type JUnitError struct {
    Message string `xml:"message,attr"`
    Type    string `xml:"type,attr"`
    Content string `xml:",chardata"`
}

type JUnitSkipped struct {
    Message string `xml:"message,attr"`
}
```

### Contraintes techniques
- **Package** : Même package `internal/validation/` (pas de sous-packages). Les fichiers `golang.go`, `report.go`, `report_json.go`, `report_junit.go` font tous partie du package `validation`.
- **Auto-enregistrement** : `GoValidator` s'enregistre via `func init() { validation.Register(NewGoValidator()) }` dans le fichier `golang.go`.
- **Commandes shell** : Utiliser `os/exec` avec `exec.CommandContext` pour le support d'annulation et timeout. Capturer stdout+stderr combinés.
- **PATH** : Utiliser `exec.LookPath("go")` pour trouver le binaire Go. Si non trouvé, retourner un message d'erreur clair avec instructions d'installation.
- **Répertoire de travail** : Les commandes Go sont exécutées dans le répertoire de travail courant du processus (où se trouve `go.mod`).
- **Timeout** : Chaque check Go a un timeout de 5 minutes. Configurable via `ModuleConfig.Options["timeout"]` en secondes.
- **JUnit XML** : Généré manuellement avec `encoding/xml` (pas de dépendance externe). Le format doit être compatible avec les rapports JUnit standard utilisés par Jenkins, GitLab CI, GitHub Actions.
- **JSON** : Généré avec `encoding/json`. Le JSON est formaté avec `json.MarshalIndent` pour la lisibilité humaine.
- **Messages en français** : Messages d'erreur, descriptions, logs en français. Les statuts (passed, failed, etc.) restent en anglais.
- **Mockable** : Les appels à `exec.Command` et `exec.LookPath` sont mockables pour les tests. Voir le pattern utilisé dans `internal/command/init.go`.
- **Tous les chemins fonctionnent sur Windows** : Utiliser les commandes Go standards qui sont multi-plateforme.

### Tests à implémenter

#### Tests — `internal/validation/validation_test.go` (ajouts)

**GoValidator :**
- **Scénario 1 : `TestGoValidator_Name`** — Name() retourne "golang"
- **Scénario 2 : `TestGoValidator_Description`** — Description() non vide, en français
- **Scénario 3 : `TestGoValidator_Detect_GoModFound`** — go.mod existe → Detect() = true
- **Scénario 4 : `TestGoValidator_Detect_GoModNotFound`** — go.mod absent → Detect() = false
- **Scénario 5 : `TestGoValidator_Detect_Error`** — erreur de lecture → Detect() = false, pas d'erreur fatale
- **Scénario 6 : `TestGoValidator_CheckGoVersion_Success`** — "go version go1.26 windows/amd64" → StatusPassed
- **Scénario 7 : `TestGoValidator_CheckGoVersion_TooOld`** — version < minimum → StatusFailed
- **Scénario 8 : `TestGoValidator_CheckGoVersion_NotFound`** — go introuvable → StatusFailed
- **Scénario 9 : `TestGoValidator_CheckGoBuild_Success`** — build réussit → StatusPassed
- **Scénario 10 : `TestGoValidator_CheckGoBuild_Failure`** — build échoue → StatusFailed
- **Scénario 11 : `TestGoValidator_CheckGoTest_Success`** — tests réussissent → StatusPassed
- **Scénario 12 : `TestGoValidator_CheckGoTest_Failure`** — tests échouent → StatusFailed
- **Scénario 13 : `TestGoValidator_CheckGoTest_NoTests`** — "no test files" → StatusWarning
- **Scénario 14 : `TestGoValidator_CheckGoVet_Success`** — vet réussit → StatusPassed
- **Scénario 15 : `TestGoValidator_CheckGoVet_Failure`** — vet échoue → StatusFailed
- **Scénario 16 : `TestGoValidator_Validate_AllChecks`** — tous les checks activés → 4 checks dans ModuleResult
- **Scénario 17 : `TestGoValidator_Validate_FilteredChecks`** — Checks=["build","test"] → seulement 2 checks
- **Scénario 18 : `TestGoValidator_Validate_Timeout`** — commande bloquée → timeout, StatusFailed
- **Scénario 19 : `TestGoValidator_ParseGoVersion`** — parseGoVersion("go version go1.26 windows/amd64") = "1.26"
- **Scénario 20 : `TestGoValidator_Validate_MinVersionOption`** — go_version option lue depuis ModuleConfig

**Reporters :**
- **Scénario 21 : `TestJSONReporter_Format`** — Format() = "json"
- **Scénario 22 : `TestJSONReporter_Write_AllPassed`** — rapport avec tous passed → JSON valide, status "passed"
- **Scénario 23 : `TestJSONReporter_Write_SomeFailed`** — rapport avec échecs → JSON contient "failed"
- **Scénario 24 : `TestJSONReporter_Write_EmptyReport`** — rapport vide (0 modules) → JSON valide
- **Scénario 25 : `TestJSONReporter_Write_Indented`** — vérifier que la sortie est bien indentée
- **Scénario 26 : `TestJUnitReporter_Format`** — Format() = "junit"
- **Scénario 27 : `TestJUnitReporter_Write_AllPassed`** — rapport avec tous passed → XML valide, 0 failures
- **Scénario 28 : `TestJUnitReporter_Write_SomeFailed`** — rapport avec échecs → failures > 0
- **Scénario 29 : `TestJUnitReporter_Write_EmptyReport`** — rapport vide → XML valide, tests="0"
- **Scénario 30 : `TestJUnitReporter_Write_WithWarnings`** — checks warning → skipped dans JUnit
- **Scénario 31 : `TestNewReportWriter_Valid`** — NewReportWriter("json") != nil
- **Scénario 32 : `TestNewReportWriter_Invalid`** — NewReportWriter("unknown") == nil
- **Scénario 33 : `TestAvailableReporters`** — AvailableReporters() contient "json" et "junit"

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #004).

### Exemples d'utilisation

```go
// Exécution du validateur Go seul
goVal := validation.NewGoValidator()
detected, _ := goVal.Detect(ctx)
if detected {
    result, _ := goVal.Validate(ctx, validation.ModuleConfig{
        Name:    "golang",
        Enabled: true,
        Checks:  []string{"version", "build", "test", "vet"},
    })
    fmt.Printf("Go: %s (%d checks)\n", result.Status, len(result.Checks))
}

// Génération des rapports
report, _ := runner.Run(ctx, cfg)

jsonReporter := validation.NewJSONReporter()
jsonFile, _ := os.Create(".ade/validation/report.json")
jsonReporter.Write(report, jsonFile)

junitReporter := validation.NewJUnitReporter()
junitFile, _ := os.Create(".ade/validation/report.xml")
junitReporter.Write(report, junitFile)
```
