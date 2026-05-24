# Tâche #003 - Story #005 : Commandes CLI `ade pipeline`

## Objectif
Ajouter la commande `ade pipeline` avec ses sous-commandes `run` et `init`, le chargement de configuration YAML, la détection de l'orchestrateur, et l'implémentation de `LocalExecutor`.

## Contexte
- Story #005 : `docs/stories/story-005.md`
- Dépend de : Tâche #001 (package `internal/ci/`), Tâche #002 (DryRunExecutor)
- Nécessaire pour : Tâche #004 (documentation)
- Note : L'orchestrateur (Story #008) n'est pas encore implémenté. La configuration se fait via fichier YAML local.

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer la commande `ade pipeline` avec deux sous-commandes :

1. **`ade pipeline run`** — Exécute le pipeline CI :
   - Charge la configuration depuis `ade-pipeline.yaml` (ou fichier spécifié par `--config`)
   - Détecte l'orchestrateur (conteneur `ade-config`) via `docker.EnsureConfigContainer()` — simple vérification de disponibilité pour l'instant (l'intégration réelle via API viendra avec Story #008)
   - Utilise `DryRunExecutor` par défaut (simulation), ou `LocalExecutor` avec le flag `--local`
   - Affiche les résultats dans un tableau formaté
   - Affiche les logs détaillés avec `--verbose` (ou `-v`)

2. **`ade pipeline init`** — Génère un fichier de configuration pipeline par défaut :
   - Crée `ade-pipeline.yaml` dans le répertoire courant (ou celui spécifié par `--output`)
   - Propose différents templates : `--template generic` (défaut, avec commentaires), `--template go`, `--template java`
   - N'écrase pas un fichier existant sans le flag `--force`

**Cas nominaux :**
- `ade pipeline run` : charge `ade-pipeline.yaml`, utilise DryRunExecutor, affiche le résumé
- `ade pipeline run --verbose` : idem + logs détaillés
- `ade pipeline run --config ./mon-pipeline.yaml` : charge la config depuis le chemin spécifié
- `ade pipeline run --local` : utilise LocalExecutor
- `ade pipeline run --dry-run` : force l'utilisation du DryRunExecutor
- `ade pipeline init` : crée `ade-pipeline.yaml` avec le template générique commenté
- `ade pipeline init --template go` : crée avec un template prêt pour Go
- `ade pipeline init --template java` : crée avec un template prêt pour Java/Maven
- `ade pipeline init --output ./ci/` : crée dans le répertoire `./ci/`

**Cas limites :**
- Fichier de config inexistant → `LoadConfig` retourne `DefaultConfig()` (charge la config par défaut)
- Config invalide → afficher les erreurs de validation et proposer de continuer avec les valeurs par défaut
- Aucun stage activé → message "Aucun stage activé dans la configuration"
- Pipeline annulé (Ctrl+C) → message "Pipeline annulé" et affichage des résultats partiels
- `ade pipeline init` avec fichier existant et sans `--force` → message "Le fichier existe déjà. Utilisez --force pour écraser."
- `ade pipeline init --template inconnu` → erreur avec la liste des templates disponibles

**Gestion d'erreurs :**
- Config vide mais explicite → pas d'erreur, utiliser `DefaultConfig()`
- Validation échouée → afficher les erreurs, demander confirmation avant de continuer
- Toutes les étapes échouent → afficher un récapitulatif avec le statut de chaque stage
- Template inconnu pour `init` → erreur listant les templates disponibles

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/pipeline.go` | Créer | Commande `ade pipeline` racine |
| `internal/command/pipeline_run.go` | Créer | Sous-commande `ade pipeline run` |
| `internal/command/pipeline_init.go` | Créer | Sous-commande `ade pipeline init` |
| `internal/command/pipeline_test.go` | Créer | Tests des commandes pipeline |
| `internal/ci/local_executor.go` | Créer | LocalExecutor (exécution de commandes shell) |
| `internal/ci/local_executor_test.go` | Créer | Tests du LocalExecutor |
| `internal/command/root.go` | Modifier | Enregistrer la commande `pipeline` dans `init()` |

### Signatures

```go
// internal/command/pipeline.go
package command

import "github.com/spf13/cobra"

var pipelineCmd = &cobra.Command{
    Use:   "pipeline",
    Short: "Gère le pipeline d'intégration continue",
    Long: `Gère le pipeline d'intégration continue locale.

Sous-commandes :
  run    Exécute le pipeline CI
  init   Génère la configuration par défaut du pipeline`,
}

func init() {
    pipelineCmd.AddCommand(pipelineRunCmd)
    pipelineCmd.AddCommand(pipelineInitCmd)
    rootCmd.AddCommand(pipelineCmd)
}
```

```go
// internal/command/pipeline_run.go
package command

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/spf13/cobra"
    "automated_dev_environment/internal/ci"
)

// Variables mockables
var (
    plLoadConfigFn     = ci.LoadConfig
    plValidateConfigFn = ci.ValidateConfig
    plNewDryRunFn      = func() ci.Executor { return ci.NewDryRunExecutor() }
    plNewLocalFn       = func() (ci.Executor, error) { return ci.NewLocalExecutor() }
    plNewPipelineFn    = func(executor ci.Executor) ci.Pipeline {
        return ci.NewPipelineRunner(executor)
    }
    plDetectOrchFn     = docker.EnsureConfigContainer
)

var (
    plRunConfigPath string
    plRunVerbose    bool
    plRunLocal      bool
    plRunDryRun     bool
)

var pipelineRunCmd = &cobra.Command{
    Use:   "run",
    Short: "Exécute le pipeline d'intégration continue",
    Long: `Exécute le pipeline d'intégration continue locale.

Le pipeline se compose de 6 étages exécutés séquentiellement :
build → tests unitaires → tests d'intégration → déploiement test → E2E → préproduction.

Par défaut, le pipeline s'exécute en mode dry-run (simulation).
Utilisez --local pour exécuter les commandes directement sur la machine.`,
    RunE: runPipeline,
}

type pipelineRunOptions struct {
    ConfigPath string
    Verbose    bool
    Local      bool
    DryRun     bool
}

func runPipeline(cmd *cobra.Command, args []string) error
func selectPipelineExecutor(ctx context.Context, opts pipelineRunOptions) (ci.Executor, error)
func loadPipelineConfig(path string) (ci.PipelineConfig, error)
func displayPipelineResult(out io.Writer, result *ci.PipelineResult, verbose bool)
func formatDuration(d time.Duration) string
```

```go
// internal/command/pipeline_init.go
package command

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
)

var (
    plInitOutput  string
    plInitForce   bool
    plInitTemplate string
)

var pipelineInitCmd = &cobra.Command{
    Use:   "init",
    Short: "Génère la configuration par défaut du pipeline",
    Long: `Génère un fichier ade-pipeline.yaml avec la configuration par défaut.

Templates disponibles :
  generic   Configuration générique avec commentaires (défaut)
  go        Configuration prête pour un projet Go
  java      Configuration prête pour un projet Java/Maven`,
    RunE: runPipelineInit,
}

func runPipelineInit(cmd *cobra.Command, args []string) error

// pipelineTemplates retourne le contenu YAML pour chaque template
func pipelineTemplateContent(template string) (string, error)

// availableTemplates liste les noms de templates disponibles
func availableTemplates() []string
```

```go
// internal/ci/local_executor.go
package ci

import "os/exec"

type LocalExecutor struct{}

func NewLocalExecutor() *LocalExecutor

// Execute runs step.Command as a shell command on the host machine.
// Captures combined stdout+stderr into StepResult.Output.
// Supports workdir and env from StepConfig.
func (e *LocalExecutor) Execute(ctx context.Context, step StepConfig) (*StepResult, error)
```

```go
// internal/command/root.go — Modification
func init() {
    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(pipelineCmd)
}
```

### Structure des données

```go
// Templates de configuration pipeline

// Template generic — commenté, l'utilisateur remplit les commandes
const pipelineTemplateGeneric = `# Configuration du pipeline CI
# G\u00e9n\u00e9r\u00e9 par 'ade pipeline init'
# Personnalisez les commandes selon votre langage et vos besoins.

stages:
  - type: build
    name: "Compilation du projet"
    enabled: true
    steps:
      - name: "Compiler"
        # Exemples : ["go", "build", "./..."]
        #           ["mvn", "clean", "compile"]
        #           ["npm", "run", "build"]
        command: []

  - type: unit-test
    name: "Tests unitaires"
    enabled: true
    steps:
      - name: "Tests unitaires"
        command: []

  - type: integration-test
    name: "Tests d'int\u00e9gration"
    enabled: true
    steps:
      - name: "Tests d'int\u00e9gration"
        command: []

  - type: test-deploy
    name: "D\u00e9ploiement test"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["docker", "compose", "-f", "docker-compose.test.yml", "up", "-d"]

  - type: e2e
    name: "Tests E2E"
    enabled: true
    steps:
      - name: "Tests E2E"
        command: []

  - type: preprod
    name: "Pr\u00e9production"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["./deploy-preprod.sh"]
`

// Template Go
const pipelineTemplateGo = `# Configuration du pipeline CI pour un projet Go

stages:
  - type: build
    name: "Compilation du projet"
    enabled: true
    steps:
      - name: "Build Go"
        command: ["go", "build", "./..."]

  - type: unit-test
    name: "Tests unitaires"
    enabled: true
    steps:
      - name: "Tests unitaires Go"
        command: ["go", "test", "./..."]

  - type: integration-test
    name: "Tests d'int\u00e9gration"
    enabled: true
    steps:
      - name: "Tests d'int\u00e9gration Go"
        command: ["go", "test", "-tags=integration", "./..."]

  - type: test-deploy
    name: "D\u00e9ploiement test"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["docker", "compose", "-f", "docker-compose.test.yml", "up", "-d"]

  - type: e2e
    name: "Tests E2E"
    enabled: true
    steps:
      - name: "Tests E2E Go"
        command: ["go", "test", "-tags=e2e", "./..."]

  - type: preprod
    name: "Pr\u00e9production"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["./deploy-preprod.sh"]
`

// Template Java (Maven)
const pipelineTemplateJava = `# Configuration du pipeline CI pour un projet Java/Maven

stages:
  - type: build
    name: "Compilation du projet"
    enabled: true
    steps:
      - name: "Build Maven"
        command: ["mvn", "clean", "compile"]

  - type: unit-test
    name: "Tests unitaires"
    enabled: true
    steps:
      - name: "Tests unitaires Maven"
        command: ["mvn", "test"]

  - type: integration-test
    name: "Tests d'int\u00e9gration"
    enabled: true
    steps:
      - name: "Tests d'int\u00e9gration Maven"
        command: ["mvn", "verify"]

  - type: test-deploy
    name: "D\u00e9ploiement test"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["docker", "compose", "-f", "docker-compose.test.yml", "up", "-d"]

  - type: e2e
    name: "Tests E2E"
    enabled: false
    steps:
      - name: "Tests E2E"
        command: []

  - type: preprod
    name: "Pr\u00e9production"
    enabled: false
    steps:
      - name: "D\u00e9ployer"
        command: ["./deploy-preprod.sh"]
`
```

### Contraintes techniques
- **Pattern Cobra** : Suivre le même pattern que `init.go` et `init_ci.go` (variables mockable, initialisation dans `init()`, français dans les messages)
- **Enregistrement** : Ajouter `pipelineCmd` à la commande racine dans `root.go` (fonction `init()`)
- **Flags pour `ade pipeline run`** :
  - `--config` (string, défaut `"ade-pipeline.yaml"`)
  - `--verbose` / `-v` (bool, défaut `false`)
  - `--local` (bool, défaut `false`)
  - `--dry-run` (bool, défaut `false`)
- **Flags pour `ade pipeline init`** :
  - `--output` / `-o` (string, défaut `"."`)
  - `--force` / `-f` (bool, défaut `false`)
  - `--template` / `-t` (string, défaut `"generic"`)
- **Logger** : Utiliser `cmd.OutOrStdout()` et `cmd.ErrOrStderr()`
- **Priorité des flags** : `--dry-run` > `--local` > défaut (dry-run). Si plusieurs flags sont passés, le plus prioritaire l'emporte.
- **DockerStepExecutor** : Pas de flag `--docker` pour l'instant. L'exécution Docker sera disponible dans une version ultérieure.
- **Orchestrateur** : Vérifier la disponibilité du conteneur `ade-config` avec `docker.EnsureConfigContainer()` mais ne pas bloquer si absent (juste un log en verbose).
- **Tableau d'affichage** : Utiliser `fmt.Fprintf` avec des largeurs fixes. Pas de dépendance de bibliothèque de tableaux.
- **Templates init** : Les templates sont des constantes string dans `pipeline_init.go`, pas des fichiers embarqués (contrairement aux templates de génération de fichiers). La génération se fait par `os.WriteFile`.
- **LocalExecutor** : Utilise `os/exec` avec `exec.CommandContext` pour l'annulation. Capture stdout+stderr combinés.
- **Windows** : Chemins par défaut dans le répertoire courant. Utiliser `filepath.Join`.

### Tests à implémenter

#### Tests unitaires — `internal/command/pipeline_test.go`

- **Scénario 1 : `TestPipelineCmd_Registered`** — `ade pipeline --help` affiche la description
- **Scénario 2 : `TestPipelineRunCmd_DryRunDefault`** — `ade pipeline run` utilise DryRunExecutor
- **Scénario 3 : `TestPipelineRunCmd_WithConfigFlag`** — `--config ./mon-pipeline.yaml` charge le bon fichier
- **Scénario 4 : `TestPipelineRunCmd_VerboseOutput`** — `--verbose` affiche les logs détaillés
- **Scénario 5 : `TestPipelineRunCmd_LocalFlag`** — `--local` utilise LocalExecutor
- **Scénario 6 : `TestPipelineRunCmd_DisplayResults`** — affichage tableau correct
- **Scénario 7 : `TestPipelineRunCmd_DisplayFailure`** — échec affiché avec stages suivants ignorés
- **Scénario 8 : `TestPipelineRunCmd_PriorityDryRun`** — `--dry-run` prioritaire sur `--local`
- **Scénario 9 : `TestPipelineInitCmd_CreateGeneric`** — `ade pipeline init` crée ade-pipeline.yaml
- **Scénario 10 : `TestPipelineInitCmd_TemplateGo`** — `--template go` crée avec des commandes Go
- **Scénario 11 : `TestPipelineInitCmd_TemplateJava`** — `--template java` crée avec des commandes Maven
- **Scénario 12 : `TestPipelineInitCmd_ForceOverwrite`** — `--force` écrase le fichier existant
- **Scénario 13 : `TestPipelineInitCmd_NoOverwriteWithoutForce`** — pas d'écrasement sans `--force`
- **Scénario 14 : `TestPipelineInitCmd_InvalidTemplate`** — template inconnu → message d'erreur
- **Scénario 15 : `TestPipelineInitCmd_CustomOutput`** — `--output ./ci/` crée dans le bon répertoire

#### Tests unitaires — `internal/ci/local_executor_test.go`

- **Scénario 16 : `TestLocalExecutor_Success`** — commande simple réussit
- **Scénario 17 : `TestLocalExecutor_Failure`** — commande qui échoue
- **Scénario 18 : `TestLocalExecutor_Cancellation`** — contexte annulé → StatusCancelled
- **Scénario 19 : `TestLocalExecutor_WorkDir`** — WorkDir respecté
- **Scénario 20 : `TestLocalExecutor_Env`** — variables d'environnement passées correctement

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #004).

### Exemples d'utilisation

```bash
# Exécution en dry-run (mode par défaut)
ade pipeline run

# Exécution avec logs détaillés
ade pipeline run --verbose

# Exécution locale
ade pipeline run --local

# Configuration personnalisée
ade pipeline run --config ./mon-pipeline.yaml

# Init : template générique
ade pipeline init

# Init : template Go prêt à l'emploi
ade pipeline init --template go

# Init : template Java Maven
ade pipeline init --template java

# Init : forcer l'écrasement
ade pipeline init --force

# Aide
ade pipeline --help
ade pipeline run --help
ade pipeline init --help
```
