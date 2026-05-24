# Tâche #002 - Story #005 : Adaptateur Dry-run du pipeline CI

## Objectif
Implémenter un adaptateur "dry-run" (`DryRunExecutor`) qui simule l'exécution du pipeline CI sans moteur réel, en produisant des résultats réalistes (logs, durées, statuts).

## Contexte
- Story #005 : `docs/stories/story-005.md`
- Dépend de : Tâche #001 (package `internal/ci/` avec types, interface Pipeline, Executor)
- Nécessaire pour : Tâche #003 (intégration CLI)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer un `DryRunExecutor` qui implémente `ci.Executor`. Il simule chaque étape du pipeline en :
1. Attendant un délai configurable (simule le temps d'exécution réel)
2. Produisant des logs réalistes selon le type d'étape
3. Retournant un `StepResult` avec un statut `StatusSucceeded` ou `StatusFailed` (configurable)

Le dry-run permet de tester l'intégration CLI, valider la configuration du pipeline, et démontrer le fonctionnement sans dépendre de Docker ou d'un moteur CI réel.

**Cas nominaux :**
- `DryRunExecutor.Execute(ctx, step)` avec `step.Name = "Build Go"` retourne un résultat après ~500ms avec des logs simulant `go build ./...`
- `DryRunExecutor.Execute(ctx, step)` avec `step.Image != ""` simule un pull et run de conteneur Docker
- Le délai par défaut pour chaque étape est configurable via `DryRunExecutor.Delay`
- Le taux de succès est configurable via `DryRunExecutor.SuccessRate` (0.0 = toujours échouer, 1.0 = toujours réussir)

**Cas limites :**
- Contexte annulé pendant la simulation → arrêt immédiat, `StatusCancelled`
- `SuccessRate = 0.0` → toutes les étapes échouent avec `StatusFailed`
- `Delay = 0` → simulation instantanée (utile pour les tests)
- Step avec `Command` vide et `Image` vide → retourne une erreur (conforme à la validation de Tâche #001)

**Gestion d'erreurs :**
- Contexte annulé → `StatusCancelled`, pas d'erreur fatale
- Simulation d'échec → `StatusFailed` avec `Err = errors.New("dry-run: étape simulée en échec")`
- Step invalide → `StatusFailed` avec `Err` contenant `ci.ErrInvalidStep`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/ci/dryrun.go` | Créer | Implémentation de DryRunExecutor |
| `internal/ci/dryrun_test.go` | Créer | Tests du DryRunExecutor |

### Signatures

```go
// internal/ci/dryrun.go
package ci

import (
    "context"
    "math/rand"
    "time"
)

// DryRunExecutor simulates pipeline step execution without a real engine.
// It implements the Executor interface.
type DryRunExecutor struct {
    // Delay is the simulated execution time per step. Default: 500ms.
    Delay time.Duration

    // SuccessRate controls the probability of success (0.0–1.0). Default: 1.0.
    SuccessRate float64

    // Random source for deterministic tests
    rng *rand.Rand
}

// NewDryRunExecutor creates a dry-run executor with default settings.
// Delay = 500ms, SuccessRate = 1.0
func NewDryRunExecutor() *DryRunExecutor

// Execute simulates running a pipeline step.
// It sleeps for Delay (with context cancellation support),
// then returns a StepResult based on SuccessRate.
func (e *DryRunExecutor) Execute(ctx context.Context, step StepConfig) (*StepResult, error)

// SimulatedOutput returns mock logs based on the step configuration.
// For steps with Command, it simulates command execution output.
// For steps with Image, it simulates container pull + run output.
func SimulatedOutput(step StepConfig) string
```

```go
// Signatures additionnelles pour les helpers
// dryrun_logs.go (optionnel, peut être intégré dans dryrun.go)

// buildLog simulates a Go build output
func buildLog() string

// testLog simulates a Go test output with passed/failed tests
func testLog(passed, failed, skipped int) string

// deployLog simulates a docker-compose deployment output
func deployLog(serviceName string) string

// containerRunLog simulates running a container step
func containerRunLog(image string, command []string) string
```

### Structure des données

```go
// Les logs simulés doivent ressembler à de vraies sorties d'outils :

// Build Go
// > go build ./...
// go: downloading github.com/foo/bar v1.0.0
// go: downloading github.com/baz/qux v2.0.0
// ✓ build succeeded (3.24s)

// Tests unitaires
// > go test ./...
// ok  	myproject/pkg/foo	0.342s
// ok  	myproject/pkg/bar	0.567s
// --- FAIL: TestSomething (0.01s)
//     something_test.go:42: expected true, got false
// FAIL	myproject/pkg/baz	0.891s
// ✓ 12 passed, 1 failed, 2 skipped (3.45s total)

// Déploiement Docker
// > docker compose up -d
// [+] Running 2/2
//  ✔ Container ade-config  Started
//  ✔ Network ade-network   Created
// ✓ Deployment completed (5.12s)

// Conteneur Docker
// > docker run --rm golang:1.26-alpine go test ./...
// Unable to find image 'golang:1.26-alpine' locally
// 1.26-alpine: Pulling from library/golang
// Digest: sha256:abc123...
// Status: Downloaded newer image for golang:1.26-alpine
// ✓ Container execution completed (12.34s)
```

### Contraintes techniques
- **Package** : Même package `internal/ci/` (pas de sous-package)
- **Simulation réaliste** : Les logs doivent ressembler à de vraies sorties d'outils de build/test/déploiement
- **Contexte** : Vérifier `ctx.Done()` pendant le sleep pour supporter l'annulation
- **RNG déterministe** : `rng` initialisé avec `rand.New(rand.NewSource(time.Now().UnixNano()))` par défaut, mais accessible pour les tests avec une seed fixe
- **SuccessRate** : Comparer `e.rng.Float64() < e.SuccessRate` pour décider succès/échec
- **Pas de dépendance externe** : Le dry-run n'utilise que la stdlib (sauf les types du package `ci`)
- **Messages en français** : Les logs et messages d'erreur sont en français (comme le reste du projet)
- **Pas de vrais appels shell** : Le dry-run ne doit JAMAIS exécuter de commandes réelles

### Tests à implémenter

#### Tests unitaires — `internal/ci/dryrun_test.go`

- **Scénario 1 : `TestDryRunExecutor_DefaultConfig`**
  - `NewDryRunExecutor()` a `Delay = 500ms`, `SuccessRate = 1.0`
  - Exécuter un step → `result.Status` = `StatusSucceeded`

- **Scénario 2 : `TestDryRunExecutor_AlwaysFails`**
  - `SuccessRate = 0.0`
  - Exécuter un step → `result.Status` = `StatusFailed`, `result.Err` non nil

- **Scénario 3 : `TestDryRunExecutor_DeterministicFailure`**
  - Utiliser un RNG seedé
  - `SuccessRate = 0.5`, seed fixe → toujours le même résultat

- **Scénario 4 : `TestDryRunExecutor_CancelledContext`**
  - Créer un contexte annulé
  - `Execute(ctx, step)` → `result.Status` = `StatusCancelled`

- **Scénario 5 : `TestDryRunExecutor_ZeroDelay`**
  - `Delay = 0`
  - L'exécution est instantanée (vérifier que `Duration` ≈ 0)

- **Scénario 6 : `TestDryRunExecutor_Delay`**
  - `Delay = 100ms`
  - `Duration` ≥ 100ms

- **Scénario 7 : `TestDryRunExecutor_InvalidStep`**
  - Step avec `Command` vide et `Image` vide → `StatusFailed`, `Err` contient `ErrInvalidStep`

- **Scénario 8 : `TestSimulatedOutput_Command`**
  - `SimulatedOutput(StepConfig{Command: []string{"go", "build", "./..."}})` contient "go build"
  - `SimulatedOutput(StepConfig{Command: []string{"go", "test", "./..."}})` contient "go test"

- **Scénario 9 : `TestSimulatedOutput_Image`**
  - `SimulatedOutput(StepConfig{Image: "golang:1.26-alpine"})` contient "docker run" ou "Pulling from"

- **Scénario 10 : `TestSimulatedOutput_DeployStep`**
  - Step avec commande contenant "compose" → output contient des logs de déploiement réalistes

- **Scénario 11 : `TestDryRunExecutor_WithPipelineRunner`**
  - Créer un `PipelineRunner` avec `DryRunExecutor` (Delay = 0, SuccessRate = 1.0)
  - Exécuter `DefaultConfig()` → 6 stages, tous `StatusSucceeded`
  - Durée totale < 1s (grâce à Delay = 0)

- **Scénario 12 : `TestDryRunExecutor_WithPipelineRunnerPartialFailure`**
  - `DryRunExecutor` avec `SuccessRate = 0.5`
  - Exécuter un pipeline de 6 étapes → au moins un succès et au moins un échec (probabiliste, avec seed fixe)

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #004).

### Exemples d'utilisation

```go
// Test du pipeline complet avec dry-run
executor := ci.NewDryRunExecutor()
executor.Delay = 100 * time.Millisecond
executor.SuccessRate = 1.0

pipeline := ci.NewPipelineRunner(executor)
result, _ := pipeline.Run(ctx, ci.DefaultConfig())

fmt.Printf("Pipeline: %s (%v)\n", result.Status, result.Duration)
for _, stage := range result.Stages {
    fmt.Printf("  %s: %s (%v)\n", stage.Type, stage.Status, stage.Duration)
    for _, step := range stage.Steps {
        fmt.Printf("    %s: %s\n", step.Name, step.Status)
        fmt.Println(step.Output)
    }
}
```
