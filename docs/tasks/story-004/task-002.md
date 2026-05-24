# Tâche #002 - Story #004 : Commande `ade init ci` — logique de déploiement

## Objectif
Implémenter la commande `ade init ci` complète : détection de Docker/Podman, génération des fichiers docker-compose.yml et .env, déploiement via docker-compose, et affichage du statut des conteneurs.

## Contexte
- Story #004 : `docs/stories/story-004.md`
- Dépend de : Tâche #001 (templates docker-compose.yml et .env)
- Nécessaire pour : Tâche #003 (documentation)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Remplacer le stub actuel de `internal/command/init_ci.go` par une implémentation complète.

**Algorithme de `ade init ci` :**
1. **Détection de l'environnement conteneurisé**
   - Appeler `docker.Check()` pour détecter Docker ou Podman
   - Si ni l'un ni l'autre n'est trouvé, afficher une erreur explicative et des instructions d'installation (Windows : Docker Desktop, winget install, etc.)
   - Si trouvé, afficher le nom du binaire détecté (ex: "Docker" ou "Podman")
2. **Vérification du démon**
   - Créer un client Docker/Podman via `docker.NewClient()`
   - Pinger le démon via `client.Ping(ctx)`
   - Si le démon n'est pas accessible, afficher une erreur avec instructions (démarrer Docker Desktop, vérifier que le service tourne)
3. **Génération des fichiers de configuration**
   - Utiliser le package `generator` avec `TemplatesFilter: []string{"docker-compose", "env"}`
   - Construire le `TemplateData` avec :
     - `ProjectName` : depuis le flag `--name` ou le nom du répertoire de sortie
     - `ConfigPort` : depuis le flag `--port` (défaut: `8080`)
     - `ComposeNetwork` : depuis le flag `--network` (défaut: `ade-network`)
   - Appliquer le flag `--force` pour écraser sans confirmation
4. **Déploiement**
   - Déterminer la commande compose appropriée :
     - Docker → `docker compose` d'abord, fallback `docker-compose` si indisponible
     - Podman → `podman-compose` (vérifier disponibilité)
   - Exécuter `[compose_cmd] up -d` dans le répertoire de sortie
   - Si la commande compose n'est pas disponible, afficher un avertissement mais ne pas échouer
   - Si le déploiement échoue (ex: image introuvable), afficher un avertissement et proposer à l'utilisateur de déployer manuellement une fois l'image configurée
5. **Affichage du statut**
   - Exécuter `[compose_cmd] ps` pour lister les conteneurs
   - Afficher un tableau formaté avec : nom du service, état, ports

**Cas nominaux :**
- `ade init ci` avec Docker disponible : détecte Docker, génère les fichiers, déploie, affiche le statut
- `ade init ci --output ./preprod` : génère les fichiers dans `./preprod/`
- `ade init ci --port 9090` : génère docker-compose.yml avec le port 9090
- `ade init ci --force` : écrase les fichiers existants sans confirmation

**Cas limites :**
- `ade init ci` avec Podman : détecte Podman, tente `podman-compose`
- `ade init ci` sans Docker ni Podman : affiche une erreur claire avec instructions
- Docker Desktop installé mais arrêté (démon injoignable) : message explicite pour démarrer Docker Desktop
- Commande compose absente (docker sans plugin compose) : avertissement, génération des fichiers quand même
- Projet sans répertoire de sortie : crée le répertoire si nécessaire

**Gestion d'erreurs :**
- Docker/Podman non trouvé → erreur `"Docker ou Podman requis. Veuillez installer Docker Desktop (https://www.docker.com/products/docker-desktop/) ou Podman (https://podman.io/)."`
- Démon injoignable → erreur `"Le démon {docker/podman} n'est pas en cours d'exécution. Démarrez Docker Desktop ou vérifiez que le service Podman tourne."`
- Échec de `docker compose up -d` → ne pas retourner l'erreur fatale. Afficher un avertissement : `"⚠ Déploiement impossible. Les fichiers sont générés. Déployez manuellement avec 'docker compose up -d' une fois l'image configurée."` (les fichiers sont la priorité)
- Tous les chemins doivent fonctionner sur Windows (antislashs, `USERPROFILE`, etc.)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/init_ci.go` | Modifier | Implémentation complète de la commande `ade init ci` |
| `internal/command/init_ci_test.go` | Créer | Tests unitaires de la commande |
| `internal/config/config.go` | Modifier | Ajouter une section `Deployment` à `AgenticConfig` si non existante |

### Signatures

```go
// internal/command/init_ci.go
package command

// Variables mockables pour les tests
var (
    dockerCheckFn       = docker.Check
    dockerNewClientFn   = docker.NewClient
    execLookPathFn      = exec.LookPath
    execCommandFn       = exec.Command  // pour les tests, capturer les appels
)

var initCiCmd = &cobra.Command{
    Use:   "ci",
    Short: "D\u00e9ploie l'environnement de pr\u00e9production",
    Long:  `D\u00e9tecte Docker/Podman, g\u00e9n\u00e8re docker-compose.yml et .env, d\u00e9ploie les conteneurs.`,
    RunE:  runInitCi,
}

// Flags
var (
    ciOutput    string
    ciForce     bool
    ciName      string
    ciPort      string
    ciNetwork   string
)

func runInitCi(cmd *cobra.Command, args []string) error
func detectAndPing(ctx context.Context, out io.Writer) (string, error)
func generateComposeFiles(ctx context.Context, out io.Writer, opts ciOptions) (*generator.Report, error)
func deployCompose(ctx context.Context, out io.Writer, binary string, workDir string) error
func showComposeStatus(ctx context.Context, out io.Writer, binary string, workDir string)

type ciOptions struct {
    OutputDir      string
    Force          bool
    ProjectName    string
    ConfigPort     string
    ComposeNetwork string
}

func (o ciOptions) toTemplateData() templates.TemplateData {
    return templates.TemplateData{
        ProjectName: o.ProjectName,
        Compose: templates.ComposeConfig{
            ConfigPort: o.ConfigPort,
            Network:    o.ComposeNetwork,
        },
    }
}
```

```go
// internal/config/config.go — Extension (optionnelle pour V1)
type DeploymentConfig struct {
    ComposePort    string `yaml:"compose_port"`
    ComposeNetwork string `yaml:"compose_network"`
}
```

### Structure des données

```go
// internal/command/init_ci.go

// ciResult regroupe les informations de d\u00e9ploiement pour l'affichage
type ciResult struct {
    Binary      string // "docker" ou "podman"
    ComposeCmd  string // Commande compose ex\u00e9cut\u00e9e
    Report      *generator.Report
    DeployErr   error
    DeployLog   string
}
```

### Contraintes techniques
- **Variables mockables** : Utiliser le même pattern que `init.go` (variables de package remplaçables dans les tests). Les appels à `docker.Check()`, `docker.NewClient()`, `exec.Command()` doivent être mockables.
- **Cobra** : Les flags `--output`, `--force`, `--name`, `--port`, `--network` sont ajoutés à `initCiCmd` dans `init()`. Le flag `--port` a un défaut de `"8080"`, `--network` a un défaut de `"ade-network"`.
- **Windows** : Tous les chemins de fichiers et commandes doivent fonctionner sous Windows. Utiliser `filepath.Join()` pour les chemins.
- **Generator** : Réutiliser `generator.Generate()` avec `TemplatesFilter` pour ne générer que les templates `docker-compose` et `env`. Ne pas recréer la logique de génération.
- **Compose commande** : Pour Docker, tenter `docker compose` (plugin v2) d'abord, fallback sur `docker-compose` (standalone). Pour Podman, vérifier `podman-compose` avec `exec.LookPath`, puis utiliser `exec.Command("podman-compose", ...)`. Si la commande compose n'est pas trouvée, afficher un avertissement et ne pas tenter le déploiement.
- **Déploiement non-fatal** : Si `up -d` échoue (ex: image `nginx:alpine` introuvable ou pas de Dockerfile), afficher un avertissement et continuer. Les fichiers générés sont la priorité. L'utilisateur pourra déployer manuellement.
- **Sortie utilisateur** : Afficher les messages en français. Utiliser `cmd.OutOrStdout()` et `cmd.ErrOrStderr()` pour les sorties. Respecter le style des messages de `init.go` (utiliser les symboles ✓, ✗, ∼ en UTF-8).
- **Image nginx:alpine** : Le template docker-compose.yml utilise `nginx:alpine` comme image placeholder. `docker compose up -d` tirera automatiquement l'image depuis Docker Hub. L'image réelle sera utilisée dans Story #008.

### Détail de l'implémentation

#### Fonction `runInitCi`
```go
func runInitCi(cmd *cobra.Command, args []string) error {
    out := cmd.OutOrStdout()
    ctx := cmd.Context()

    outputDir := ciOutput
    if outputDir == "" {
        outputDir = "."
    }

    // \u00c9tape 1 : D\u00e9tection et ping
    binary, err := detectAndPing(ctx, out)
    if err != nil {
        return err
    }

    // \u00c9tape 2 : Pr\u00e9paration des options
    projectName := ciName
    if projectName == "" {
        absDir, _ := filepath.Abs(outputDir)
        projectName = filepath.Base(absDir)
    }
    if projectName == "" || projectName == "." {
        projectName = "preprod"
    }

    opts := ciOptions{
        OutputDir:      outputDir,
        Force:          ciForce,
        ProjectName:    projectName,
        ConfigPort:     ciPort,
        ComposeNetwork: ciNetwork,
    }

    // \u00c9tape 3 : G\u00e9n\u00e9ration
    fmt.Fprintf(out, "  \u2713 G\u00e9n\u00e9ration des fichiers de d\u00e9ploiement...\n")
    report, err := generateComposeFiles(ctx, out, opts)
    if err != nil {
        return fmt.Errorf("g\u00e9n\u00e9ration des fichiers: %w", err)
    }
    for _, f := range report.Files {
        if f.Status == generator.StatusCreated {
            fmt.Fprintf(out, "    \u2713 %s cr\u00e9\u00e9\n", f.TargetPath)
        } else if f.Status == generator.StatusOverwritten {
            fmt.Fprintf(out, "    \u2713 %s mis \u00e0 jour\n", f.TargetPath)
        } else if f.Status == generator.StatusSkipped {
            fmt.Fprintf(out, "    \u223c %s ignor\u00e9\n", f.TargetPath)
        }
    }

    // \u00c9tape 4 : D\u00e9ploiement
    fmt.Fprintf(out, "\n  \u2192 D\u00e9ploiement avec %s...\n", binary)
    deployCompose(ctx, out, binary, outputDir)

    // \u00c9tape 5 : Statut
    fmt.Fprintf(out, "\n  \u2713 Statut des conteneurs:\n")
    showComposeStatus(ctx, out, binary, outputDir)

    return nil
}
```

#### Fonction `detectAndPing`
```go
func detectAndPing(ctx context.Context, out io.Writer) (string, error) {
    binary, err := dockerCheckFn()
    if err != nil {
        return "", fmt.Errorf("Docker ou Podman requis. "+
            "Veuillez installer Docker Desktop (https://www.docker.com/products/docker-desktop/) "+
            "ou Podman (https://podman.io/).")
    }
    fmt.Fprintf(out, "  \u2713 %s trouv\u00e9\n", binary)

    client, err := dockerNewClientFn()
    if err != nil {
        return binary, fmt.Errorf("connexion \u00e0 %s impossible: %v. "+
            "V\u00e9rifiez que le service est install\u00e9.", binary, err)
    }
    defer client.Close()

    if err := client.Ping(ctx); err != nil {
        return binary, fmt.Errorf("le d\u00e9mon %s n'est pas accessible: %v. "+
            "D\u00e9marrez Docker Desktop ou v\u00e9rifiez le service Podman.", binary, err)
    }
    fmt.Fprintf(out, "  \u2713 D\u00e9mon %s actif\n", binary)

    return binary, nil
}
```

#### Fonction `generateComposeFiles`
```go
func generateComposeFiles(ctx context.Context, out io.Writer, opts ciOptions) (*generator.Report, error) {
    absDir, err := filepath.Abs(opts.OutputDir)
    if err != nil {
        return nil, err
    }

    tmplData := opts.toTemplateData()

    genOpts := generator.Options{
        OutputDir:       absDir,
        Force:           opts.Force,
        TemplateData:    tmplData,
        TemplatesFilter: []string{"docker-compose", "env"},
        Prompter:        &generator.StdPrompter{},
    }

    return generator.Generate(ctx, genOpts)
}
```

#### Fonction `deployCompose`
```go
func deployCompose(ctx context.Context, out io.Writer, binary string, workDir string) error {
    composeCmd, composeArgs := getComposeCommand(binary)
    if composeCmd == "" {
        fmt.Fprintf(out, "  \u26a0 Commande compose non trouv\u00e9e pour %s. "+
            "Les fichiers sont g\u00e9n\u00e9r\u00e9s. D\u00e9ployez manuellement avec 'docker compose up -d'.\n", binary)
        return nil
    }

    args := append(composeArgs, "up", "-d")
    cmd := execCommandFn(composeCmd, args...)
    cmd.Dir = workDir
    cmd.Stdout = out
    cmd.Stderr = out
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(out, "  \u26a0 D\u00e9ploiement impossible. Les fichiers sont g\u00e9n\u00e9r\u00e9s. "+
            "D\u00e9ployez manuellement avec '%s up -d' une fois l'image configur\u00e9e.\n", composeCmd)
        return nil // non-fatal : les fichiers sont la priorit\u00e9
    }
    return nil
}

func getComposeCommand(binary string) (string, []string) {
    switch binary {
    case "docker":
        if _, err := execLookPathFn("docker"); err == nil {
            // docker compose (plugin v2) — tenter avec espace
            if err := execCommandFn("docker", "compose", "version").Run(); err == nil {
                return "docker", []string{"compose"}
            }
        }
        // fallback : docker-compose standalone
        if _, err := execLookPathFn("docker-compose"); err == nil {
            return "docker-compose", nil
        }
    case "podman":
        if _, err := execLookPathFn("podman-compose"); err == nil {
            return "podman-compose", nil
        }
    }
    return "", nil
}
```

#### Fonction `showComposeStatus`
```go
func showComposeStatus(ctx context.Context, out io.Writer, binary string, workDir string) {
    composeCmd, composeArgs := getComposeCommand(binary)
    if composeCmd == "" {
        return
    }

    args := append(composeArgs, "ps")
    cmd := execCommandFn(composeCmd, args...)
    cmd.Dir = workDir
    cmd.Stdout = out
    cmd.Stderr = out
    cmd.Run() // non-fatal
}
```

### Tests à implémenter

#### Tests unitaires — `internal/command/init_ci_test.go`
Les tests doivent suivre le pattern de `init_test.go` (variables mockables, `execInit()` helper).

- **Scénario 1 : `TestInitCi_DetectDocker`**
  - Mocker `dockerCheckFn` pour retourner `"docker", nil`
  - Mocker `dockerNewClientFn` avec un mock qui réussit le Ping
  - Vérifier que la sortie contient "Docker trouvé"

- **Scénario 2 : `TestInitCi_DetectPodman`**
  - Mocker `dockerCheckFn` pour retourner `"podman", nil`
  - Vérifier que la sortie contient "Podman"

- **Scénario 3 : `TestInitCi_NoDockerFound`**
  - Mocker `dockerCheckFn` pour retourner une erreur
  - Vérifier que la commande retourne une erreur contenant "Docker ou Podman requis"

- **Scénario 4 : `TestInitCi_DaemonUnreachable`**
  - Mocker `dockerCheckFn` pour retourner `"docker", nil`
  - Mocker `dockerNewClientFn` avec un mock qui échoue au Ping
  - Vérifier que la commande retourne une erreur contenant "démon"

- **Scénario 5 : `TestInitCi_GeneratesFiles`**
  - Mocker toutes les dépendances docker pour réussir
  - Exécuter `ade init ci --output <tempdir> --force`
  - Vérifier que `docker-compose.yml` et `.env` existent dans le répertoire

- **Scénario 6 : `TestInitCi_ProjectNameDefaults`**
  - Exécuter avec `--name mon-projet`
  - Vérifier que le `.env` contient `ADE_PROJECT_NAME=mon-projet`

- **Scénario 7 : `TestInitCi_CustomPort`**
  - Exécuter avec `--port 9090`
  - Vérifier que `docker-compose.yml` contient `"9090:80"`

- **Scénario 8 : `TestInitCi_ComposeNotFoundContinues`**
  - Mocker `execLookPathFn` pour que compose ne soit pas trouvé
  - Vérifier que la commande continue (ne retourne pas d'erreur) et affiche un avertissement

- **Scénario 9 : `TestInitCi_HelpContainsNewFlags`**
  - Vérifier que `ade init ci --help` contient `--output`, `--force`, `--name`, `--port`, `--network`

- **Scénario 10 : `TestInitCi_DeployAndShowStatus`**
  - Mocker toutes les dépendances docker
  - Mocker `execLookPathFn` pour retourner `"docker", nil`
  - Mocker `execCommandFn` pour capturer les appels à `docker compose up -d` et `docker compose ps`
  - Vérifier que les commandes sont appelées dans le bon ordre (detect → ping → generate → up -d → ps)

- **Scénario 11 : `TestInitCi_FallbackToDockerCompose`**
  - Mocker `dockerCheckFn` pour retourner `"docker", nil`
  - Mocker `execCommandFn("docker", "compose", "version")` pour échouer (plugin compose absent)
  - Mocker `execLookPathFn` pour retourner `"docker-compose", nil` trouvé
  - Vérifier que la commande utilise `docker-compose` (standalone) au lieu de `docker compose`

- **Scénario 12 : `TestInitCi_DeployFailureNonFatal`**
  - Mocker toutes les dépendances pour réussir
  - Mocker `execCommandFn(... "up", "-d")` pour retourner une erreur
  - Vérifier que la commande ne retourne PAS d'erreur (non-fatal)
  - Vérifier que la sortie contient "Les fichiers sont générés"

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #003).

### Exemples d'utilisation
```bash
# D\u00e9ploiement de base
ade init ci

# Avec options personnalis\u00e9es
ade init ci --output ./preprod --port 9090 --network app-net --name mon-app

# \u00c9craser les fichiers existants
ade init ci --force

# Voir l'aide
ade init ci --help
```
