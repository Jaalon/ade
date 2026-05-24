# Tâche #001 - Story #005 : Interface Pipeline CI et types fondamentaux

## Objectif
Définir l'interface abstraite du pipeline CI avec les types fondamentaux (StageType, StageStatus, StepConfig, PipelineConfig, résultats d'exécution), le moteur d'exécution PipelineRunner, et étendre l'interface docker.Client pour les futures exécutions conteneurisées.

## Contexte
- Story #005 : `docs/stories/story-005.md`
- Plan d'implémentation : Phase 3, après Story #001 et #004. Parallélisable avec Story #003.
- Dépend de : Story #001 (package `internal/command/`, patterns Go), Story #004 (package `internal/docker/`)
- Nécessaire pour : Tâche #002 (DryRunAdapter), Tâche #003 (CLI)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer le package `internal/ci/` avec les types et l'interface du pipeline CI. Le pipeline est agnostique du moteur sous-jacent : il définit un contrat que pourront implémenter un adaptateur dry-run (Tâche #002), ou un futur moteur réel.

Étendre l'interface `docker.Client` dans `internal/docker/` avec les méthodes nécessaires pour exécuter des conteneurs Docker (pull, create, start, wait, logs). Ces méthodes sont ajoutées maintenant pour définir le contrat, mais le `DockerStepExecutor` dans `internal/ci/` est un stub qui lance `ErrNotImplemented` — son implémentation complète sera faite dans une itération ultérieure (Story #007/008).

Le pipeline se compose de 6 étages (stages) prédéfinis, exécutés dans l'ordre :
1. **build** — Construction du projet
2. **unit-test** — Tests unitaires
3. **integration-test** — Tests d'intégration
4. **test-deploy** — Déploiement dans un environnement de test
5. **e2e** — Tests end-to-end
6. **preprod** — Déploiement en préproduction

Chaque étage contient une ou plusieurs étapes (steps). Chaque étape peut être exécutée localement (commande shell) ou via un conteneur Docker (plugin, futur). La configuration du pipeline est exprimée en YAML.

**Cas nominaux :**
- `PipelineConfig` se charge et se valide depuis un struct Go ou une map YAML
- Les 6 stages sont définis dans l'ordre correct
- `Pipeline.Run(ctx, config)` exécute chaque stage séquentiellement et retourne `PipelineResult`
- Chaque `StageType` a une description lisible et un ordre fixe
- `StageConfig.Enabled = false` permet de sauter un stage
- Un `StepConfig` avec `Command` renseigné est exécuté via commande shell ; `Image` est réservé pour le futur

**Cas limites :**
- Pipeline config vide → `PipelineConfig` valide avec tous les stages utilisant les valeurs par défaut
- Tous les stages désactivés → pipeline "réussi" mais vide (aucun stage exécuté)
- `PipelineConfig` avec des `StageType` inconnus → erreur de validation

**Gestion d'erreurs :**
- Stage inconnu → `ErrUnknownStage` avec le nom du stage
- Étape avec `Command` vide → erreur de validation (sauf si `Image` est défini pour usage futur, mais celui-ci n'est pas encore implémenté)
- Résultat d'étape en échec → le stage est marqué `StatusFailed`, les stages suivants sont `StatusSkipped`
- `context.Context` annulé pendant l'exécution → arrêt immédiat, statut `StatusCancelled`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/ci/types.go` | Créer | Types fondamentaux : StageType, StageStatus, StepConfig, StageConfig, PipelineConfig |
| `internal/ci/result.go` | Créer | Types de résultat : StepResult, StageResult, PipelineResult |
| `internal/ci/pipeline.go` | Créer | Interface Pipeline, PipelineRunner, fonctions de validation et d'exécution |
| `internal/ci/executor.go` | Créer | Executor interface (DryRunExecutor dans Tâche #002, LocalExecutor dans Tâche #003) |
| `internal/ci/config.go` | Créer | Chargement et validation de PipelineConfig (YAML) |
| `internal/ci/errors.go` | Créer | Erreurs sentinelles du package |
| `internal/ci/ci_test.go` | Créer | Tests unitaires |
| `internal/docker/docker.go` | Modifier | Étendre l'interface Client avec des méthodes d'exécution de conteneurs |
| `internal/docker/docker_test.go` | Modifier | Tests pour les nouvelles méthodes (mock) |

### Signatures

```go
// internal/ci/types.go
package ci

type StageType string

const (
    StageBuild           StageType = "build"
    StageUnitTest        StageType = "unit-test"
    StageIntegrationTest StageType = "integration-test"
    StageTestDeploy      StageType = "test-deploy"
    StageE2E             StageType = "e2e"
    StagePreprod         StageType = "preprod"
)

// AllStages returns all stages in execution order
func AllStages() []StageType

// StageDescription returns a human-readable French description
func StageDescription(t StageType) string

type StageStatus string

const (
    StatusPending    StageStatus = "pending"
    StatusRunning    StageStatus = "running"
    StatusSucceeded  StageStatus = "succeeded"
    StatusFailed     StageStatus = "failed"
    StatusSkipped    StageStatus = "skipped"
    StatusCancelled  StageStatus = "cancelled"
)

// StepConfig defines a single pipeline step
type StepConfig struct {
    Name    string            `yaml:"name" json:"name"`
    Image   string            `yaml:"image,omitempty" json:"image,omitempty"`
    Command []string          `yaml:"command,omitempty" json:"command,omitempty"`
    Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
    WorkDir string            `yaml:"workdir,omitempty" json:"workdir,omitempty"`
}

// StageConfig defines a pipeline stage
type StageConfig struct {
    Type    StageType    `yaml:"type" json:"type"`
    Name    string       `yaml:"name,omitempty" json:"name,omitempty"`
    Steps   []StepConfig `yaml:"steps" json:"steps"`
    Enabled bool         `yaml:"enabled" json:"enabled"`
}

// PipelineConfig defines the full pipeline configuration
type PipelineConfig struct {
    Stages []StageConfig `yaml:"stages" json:"stages"`
}
```

```go
// internal/ci/result.go
type StepResult struct {
    Name     string
    Status   StageStatus
    Output   string
    Duration time.Duration
    Err      error
}

type StageResult struct {
    Type     StageType
    Name     string
    Status   StageStatus
    Steps    []StepResult
    Duration time.Duration
    Started  time.Time
}

type PipelineResult struct {
    Stages    []StageResult
    Status    StageStatus
    Duration  time.Duration
    StartedAt time.Time
    CompletedAt time.Time
}

func (r *PipelineResult) Failed() bool
func (r *PipelineResult) Succeeded() bool
```

```go
// internal/ci/pipeline.go

type Pipeline interface {
    Run(ctx context.Context, config PipelineConfig) (*PipelineResult, error)
    Validate(config PipelineConfig) error
}

type PipelineRunner struct {
    executor Executor
}

func NewPipelineRunner(executor Executor) *PipelineRunner
func (r *PipelineRunner) Run(ctx context.Context, config PipelineConfig) (*PipelineResult, error)
func (r *PipelineRunner) Validate(config PipelineConfig) error
```

```go
// internal/ci/executor.go

type Executor interface {
    Execute(ctx context.Context, step StepConfig) (*StepResult, error)
}

// DockerStepExecutor est un stub — l'implantation réelle viendra avec la Story #007/008.
// Pour l'instant, toute exécution retourne ErrNotImplemented.
type DockerStepExecutor struct{}

func NewDockerStepExecutor() *DockerStepExecutor
func (e *DockerStepExecutor) Execute(ctx context.Context, step StepConfig) (*StepResult, error)
// Retourne toujours un résultat StatusFailed avec ErrNotImplemented
```

```go
// internal/ci/config.go

func DefaultConfig() PipelineConfig
func LoadConfig(path string) (PipelineConfig, error)
func ValidateConfig(cfg PipelineConfig) error
```

```go
// internal/ci/errors.go
package ci

var (
    ErrUnknownStage    = errors.New("stage inconnu")
    ErrInvalidStep     = errors.New("étape invalide")
    ErrConfigInvalid   = errors.New("configuration invalide")
    ErrStageOrder      = errors.New("ordre des stages invalide")
    ErrNotImplemented  = errors.New("non implémenté — sera disponible dans une version ultérieure")
)
```

```go
// internal/docker/docker.go — Extension de l'interface Client

type ContainerExecConfig struct {
    Image   string
    Command []string
    Env     map[string]string
    WorkDir string
}

type ContainerExecResult struct {
    Output   string
    ExitCode int64
}

type Client interface {
    Ping(ctx context.Context) error
    IsContainerRunning(ctx context.Context, containerName string) (bool, error)
    Close() error

    // Nouvelles méthodes pour l'exécution de conteneurs (utilisées par DockerStepExecutor dans le futur)
    PullImage(ctx context.Context, image string) error
    RunContainer(ctx context.Context, cfg ContainerExecConfig) (*ContainerExecResult, error)
}

// Les méthodes PullImage et RunContainer sont ajoutées à realClient.
// Pour l'instant, elles peuvent lever une erreur "not implemented" ou être
// implémentées avec le SDK Docker (github.com/docker/docker/client).
// PullImage : utilise client.ImagePull
// RunContainer : combine ContainerCreate, ContainerStart, ContainerWait, ContainerLogs
```

### Structure des données

```go
// Exemple de configuration YAML générique avec commentaires
// ade-pipeline.yaml

stages:
  - type: build
    name: "Compilation du projet"
    enabled: true
    steps:
      - name: "Compiler le projet"
        # Commande à personnaliser selon votre langage
        # Exemples :
        # Go:  ["go", "build", "./..."]
        # Java/Maven:  ["mvn", "clean", "compile"]
        # Java/Gradle: ["gradle", "build", "-x", "test"]
        # Node: ["npm", "run", "build"]
        command: []

  - type: unit-test
    name: "Tests unitaires"
    enabled: true
    steps:
      - name: "Exécuter les tests unitaires"
        # Commande à personnaliser
        # Go:  ["go", "test", "./..."]
        # Java/Maven:  ["mvn", "test"]
        # Java/Gradle: ["gradle", "test"]
        # Node: ["npm", "test"]
        command: []

  - type: integration-test
    name: "Tests d'intégration"
    enabled: true
    steps:
      - name: "Exécuter les tests d'intégration"
        # Go:  ["go", "test", "-tags=integration", "./..."]
        # Java/Maven:  ["mvn", "verify"]
        # Java/Gradle: ["gradle", "integrationTest"]
        command: []

  - type: test-deploy
    name: "Déploiement environnement de test"
    enabled: false
    steps:
      - name: "Déployer l'environnement de test"
        command: ["docker", "compose", "-f", "docker-compose.test.yml", "up", "-d"]

  - type: e2e
    name: "Tests E2E"
    enabled: true
    steps:
      - name: "Exécuter les tests E2E"
        # Tags E2E à définir selon le langage
        command: []

  - type: preprod
    name: "Déploiement préproduction"
    enabled: false
    steps:
      - name: "Déployer en préproduction"
        command: ["./deploy-preprod.sh"]
```

### Contraintes techniques
- **Package** : `internal/ci/` — nouveau package, bien séparé des autres
- **docker.Client** : Étendre l'interface existante — ne pas casser la compatibilité ascendante. Les nouvelles méthodes (`PullImage`, `RunContainer`) ont une implémentation réelle dans `realClient` via le SDK Docker. Les mocks existants doivent être mis à jour.
- **DockerStepExecutor** : C'est un **stub**. Il implémente `Executor` mais retourne `ErrNotImplemented` pour toute exécution. Il sert de placeholder pour l'intégration future.
- **Dépendances** : Le package `internal/ci/` n'importe PAS `internal/docker/` directement — l'interface Executor est indépendante.
- **YAML** : Utiliser `gopkg.in/yaml.v3` déjà dans go.mod pour le parsing de config
- **Pas de dépendance Cobra** : Le package `internal/ci/` ne doit pas importer Cobra
- **`PipelineResult.Failed()`** : Retourne true si au moins un stage est `StatusFailed`
- **`PipelineResult.Succeeded()`** : Retourne true seulement si tous les stages exécutés ont le statut `StatusSucceeded`
- **Ordre des stages** : `AllStages()` définit l'ordre canonique. `ValidateConfig` réordonne automatiquement les stages
- **Stage sans steps** : Si une `StageConfig` a `Steps` vide et `Enabled = true`, c'est une erreur de validation
- **Tests** : Coverage minimum 80% pour ce package

### Tests à implémenter

#### Tests unitaires — `internal/ci/ci_test.go`

- **Scénario 1 : `TestStageTypeOrder`** — AllStages() retourne les 6 stages dans le bon ordre
- **Scénario 2 : `TestStageDescription`** — descriptions françaises non vides pour chaque stage
- **Scénario 3 : `TestDefaultConfig`** — 6 stages tous Enabled = true, chaque stage a ≥ 1 step
- **Scénario 4 : `TestValidateConfig_Valid`** — DefaultConfig() valide retourne nil
- **Scénario 5 : `TestValidateConfig_UnknownStage`** — erreur ErrUnknownStage
- **Scénario 6 : `TestValidateConfig_EmptyCommand`** — step sans Command → erreur (Image ignoré pour l'instant)
- **Scénario 7 : `TestValidateConfig_DisabledNoSteps`** — stage désactivé sans steps → pas d'erreur
- **Scénario 8 : `TestValidateConfig_Reordering`** — stages dans le désordre → réordonnés
- **Scénario 9 : `TestPipelineResult_Failed`** — Failed() = true si un stage a échoué
- **Scénario 10 : `TestPipelineResult_Succeeded`** — Succeeded() = true si tous réussis
- **Scénario 11 : `TestPipelineResult_SucceededWithSkipped`** — stages skipped → Succeeded() = true
- **Scénario 12 : `TestPipelineRunner_AllSuccessful`** — mock Executor, 6 stages tous réussis
- **Scénario 13 : `TestPipelineRunner_StageFailure`** — 3e étape échoue, les suivantes sont skipped
- **Scénario 14 : `TestPipelineRunner_AllDisabled`** — tous désactivés, pipeline vide mais réussi
- **Scénario 15 : `TestPipelineRunner_Cancellation`** — contexte annulé → StatusCancelled
- **Scénario 16 : `TestLoadConfig_FileNotFound`** — retourne DefaultConfig()
- **Scénario 17 : `TestLoadConfig_ValidYAML`** — fichier YAML temporaire parsé correctement
- **Scénario 18 : `TestDockerStepExecutor_NotImplemented`** — DockerStepExecutor.Execute retourne ErrNotImplemented
- **Scénario 19 : `TestPipelineRunner_ExecutorIntegration`** — vérifier que Execute est appelé pour chaque étape

#### Tests — `internal/docker/docker_test.go` (extension)

- **Scénario 20 : `TestClient_PullImage`** — mock du client SDK, vérifier que PullImage appelle ImagePull
- **Scénario 21 : `TestClient_RunContainer`** — mock du client SDK, vérifier le cycle create → start → wait → logs

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #004).

### Exemples d'utilisation

```go
// Création et exécution du pipeline
config := ci.DefaultConfig()
executor := ci.NewDryRunExecutor() // Tâche #002
pipeline := ci.NewPipelineRunner(executor)

result, err := pipeline.Run(ctx, config)
if err != nil {
    log.Fatal(err)
}

for _, stage := range result.Stages {
    fmt.Printf("%s (%s): %s\n", stage.Name, stage.Type, stage.Status)
}

// Chargement depuis YAML
cfg, err := ci.LoadConfig("ade-pipeline.yaml")
```
