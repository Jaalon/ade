# Tâche #004 - Story #003 : Commande `ade init`, tests et documentation

## Objectif
Intégrer toutes les fonctionnalités agentic dans la commande `ade init` (détection des outils, skills, MCP), ajouter les tests complets, et créer la documentation utilisateur.

## Contexte
- Story #003 : `docs/stories/story-003.md`
- Dépend de : Tâches #001, #002, #003 (tous les composants agentic + `internal/config`)
- Nécessaire pour : Rien (dernière tâche de la story)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Modification de la commande `ade init`
Modifier la commande `ade init` (`internal/command/init.go`) pour qu'elle exécute le setup agentic complet quand elle est invoquée sans sous-commande.

**Comportement avec sous-commande :**
- `ade init specs` → génération des fichiers de config (inchangé, Story #002)
- `ade init ci` → initialisation CI (inchangé, Story #004 à venir)
- `ade init --help` → affiche l'aide (inchangé)

**Comportement sans sous-commande :**
- `ade init` exécute le setup agentic complet dans cet ordre :
  1. Détection des outils (OpenCode, Cursor) via `agentic.DetectTools()` + `config.LoadConfig()`
  2. Installation/référencement des skills via `agentic.EnsureSkills()`
  3. Configuration des serveurs MCP via `agentic.ConfigureMCPServers()`
  4. Rapport final

**Flags pour `ade init` :**
- `--force, -f` : Écraser les fichiers existants (bool, défaut: false)
- `--output, -o` : Répertoire du projet (string, défaut: ".")
- `--config` : Chemin du fichier de configuration YAML (string, défaut: auto-détection)
- `--skip-tools` : Ignorer la détection des outils (bool, défaut: false)
- `--skip-skills` : Ignorer l'installation des skills (bool, défaut: false)
- `--skip-mcp` : Ignorer la configuration MCP (bool, défaut: false)
- `--halt-on-error` : Arrêter le setup dès la première erreur (bool, défaut: false) — par défaut, toutes les étapes sont exécutées et un rapport récapitulatif est affiché

**Rapport d'exécution :**
La commande affiche un rapport structuré après exécution :

```
✓ Configuration agentic terminée

Outils détectés :
  ✓ OpenCode trouvé : C:\Users\user\AppData\Local\Programs\opencode\opencode.exe
  ✗ Cursor non trouvé → https://cursor.com

Skills :
  ✓ 7 skills installés

Serveurs MCP :
  ✓ 2 serveurs configurés dans .opencode/config.json
```

**Messages pour outils manquants :**
Si OpenCode ou Cursor n'est pas installé, afficher un message clair avec les instructions d'installation et l'URL de téléchargement. Le message doit être affiché pendant l'exécution, pas seulement dans le rapport final.

**Gestion d'erreurs :**
- Par défaut (sans `--halt-on-error`) : toutes les étapes sont exécutées, les erreurs sont collectées et affichées dans le rapport final. Code de sortie 0 si au moins une étape a réussi, code 1 si toutes ont échoué.
- Avec `--halt-on-error` : la première erreur fatale arrête le setup immédiatement, code de sortie non-zero.

#### Tests

**Tests unitaires pour la commande :**
- Tester que `ade init` exécute les 3 étapes du setup agentic (tools, skills, mcp)
- Tester que `ade init --skip-*` ignore les étapes correspondantes
- Tester que `--halt-on-error` arrête le setup en cas d'erreur
- Tester que les flags sont correctement transmis

**Tests d'intégration :**
- Exécuter `ade init` dans un répertoire temporaire et vérifier les fichiers créés/outils simulés

**Test E2E :**
- Compiler le binaire et exécuter `ade init` avec simulateur de PATH

#### Documentation

Créer les fichiers de documentation suivants :
1. `docs/agentic/setup.md` — Guide de configuration des outils agentic
2. `docs/agentic/skills.md` — Liste et description des skills installés
3. `docs/configuration/yaml.md` — Configuration des chemins d'outils et MCP dans le YAML

Mettre à jour :
- `docs/cli/commands.md` — Ajouter la section `ade init` (setup agentic)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/init.go` | Modifier | Ajouter RunE pour le setup agentic |
| `internal/command/init_test.go` | Modifier | Tests de la commande `ade init` |
| `test/e2e/e2e_test.go` | Modifier | Test E2E pour `ade init` |
| `docs/agentic/setup.md` | Créer | Guide de configuration agentic |
| `docs/agentic/skills.md` | Créer | Liste des skills |
| `docs/configuration/yaml.md` | Créer | Config YAML des outils et MCP |
| `docs/cli/commands.md` | Modifier | Ajouter section `ade init` |

### Signatures

```go
// internal/command/init.go (modifications)

var (
	// Nouveaux flags pour le setup agentic
	initForce       bool
	initOutput      string
	initConfig      string
	initSkipTools   bool
	initSkipSkills  bool
	initSkipMCP     bool
	initHaltOnError bool
)

// Nouvelle fonction de setup agentic appelée par RunE
func runAgenticSetup(cmd *cobra.Command, args []string) error

// Fonctions auxiliaires pour chaque étape
func detectToolsAndReport(out io.Writer, configPath string)
func installSkillsAndReport(ctx context.Context, out io.Writer, outputDir string, force bool)
func configureMCPAndReport(ctx context.Context, out io.Writer, opts MCPOptions)
func printMissingTools(out io.Writer, result *DetectionResult)
```

### Contraintes techniques
- **Modification minimale** : Ne pas casser le comportement existant de `ade init specs` et `ade init ci`. Le nouveau `RunE` sur `initCmd` ne s'exécute que quand aucune sous-commande n'est invoquée (comportement Cobra standard : la commande parente exécute `RunE` seulement si aucune sous-commande n'est trouvée).
- **Flags** : Les nouveaux flags sont définis dans `init()` du fichier `init.go`
- **Rapport** : Utiliser `cmd.OutOrStdout()` pour les sorties, `fmt.Fprintf(os.Stderr)` pour les warnings
- **Ordre d'exécution** : Strict : détection → skills → MCP
- **Skip flags** : Chaque `--skip-*` flag évite l'appel à l'étape correspondante
- **Halt-on-error** : Si `--halt-on-error` est actif, chaque fonction `*AndReport()` doit vérifier une erreur partagée et arrêter la chaîne. Sinon, collecter les erreurs et continuer.
- **Code de sortie** : Sans `--halt-on-error`, retourner `nil` si au moins une étape a réussi. Retourner une erreur seulement si toutes les étapes ont échoué.
- **Tests** : Les tests de la commande utilisent `cmd.SetArgs()` et `cmd.SetOut()` pour capturer la sortie
- **Fichier de config YAML** : Toujours chercher automatiquement via `config.FindConfigPath()` si `--config` n'est pas spécifié

#### Contenu attendu de la documentation

**`docs/agentic/setup.md` :**
```markdown
# Configuration des outils agentic

## Description
La commande `ade init` configure automatiquement les outils agentic (OpenCode, Cursor)
et les serveurs MCP pour votre projet.

## Utilisation

### Configuration complète
```powershell
ade init
```

### Configuration avec options
```powershell
# Spécifier le répertoire du projet
ade init --output C:\Projects\mon-app

# Écraser les fichiers existants
ade init --force

# Ignorer certaines étapes
ade init --skip-tools --skip-mcp

# Arrêter à la première erreur
ade init --halt-on-error
```

## Étapes exécutées

1. **Détection des outils** — Vérifie la présence d'OpenCode et Cursor
2. **Installation des skills** — Copie les skills OpenCode dans `.opencode/skills/`
3. **Configuration MCP** — Configure les serveurs MCP dans `.opencode/config.json`

## Configuration YAML
Voir `docs/configuration/yaml.md` pour la configuration des chemins d'outils et des serveurs MCP.
```

**`docs/agentic/skills.md` :**
```markdown
# Skills OpenCode

## Description
Les skills sont des instructions spécialisées qui étendent les capacités d'OpenCode.
Ils sont installés dans le répertoire `.opencode/skills/` du projet.

## Skills installés

| Skill | Description |
|-------|-------------|
| `specification-en` | Création de spécifications (anglais) |
| `specification-fr` | Création de spécifications (français) |
| `story-en` | Découpage en user stories (anglais) |
| `story-fr` | Découpage en user stories (français) |
| `tasks-en` | Découpage en tâches exécutables (anglais) |
| `tasks-fr` | Découpage en tâches exécutables (français) |
| `feedback-fr` | Génération de stories depuis des feedbacks |

## Installation
Les skills sont automatiquement installés par `ade init` ou `ade init specs --force`.
```

**`docs/configuration/yaml.md` :**
```markdown
# Configuration YAML

## Emplacement
Le fichier de configuration peut être placé à :
- `./ade-config.yaml` (racine du projet)
- `./.ade.yaml` (racine du projet)
- `%USERPROFILE%\.ade\config.yaml` (global utilisateur)

## Structure

```yaml
# Configuration des outils agentic
tools:
  opencode:
    path: "C:\\chemin\\personnalise\\opencode.exe"
  cursor:
    path: "C:\\chemin\\personnalise\\Cursor.exe"

# Configuration des serveurs MCP
mcp_servers:
  - name: "filesystem"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
```

## Sections

| Section | Description |
|---------|-------------|
| `tools` | Surcharge des chemins des outils agentic |
| `mcp_servers` | Définition des serveurs MCP à configurer |
```

### Tests à implémenter

#### Tests unitaires — `internal/command/init_test.go`

- Scénario 1 : `ade init` sans sous-commande exécute le setup agentic
  - Setup : mocker les 3 étapes pour retourner succès
  - Résultat attendu : les 3 étapes sont appelées

- Scénario 2 : `ade init --skip-tools` ignore la détection
  - Résultat attendu : la détection n'est pas appelée

- Scénario 3 : `ade init --skip-skills` ignore l'installation des skills
  - Résultat attendu : l'installation des skills n'est pas appelée

- Scénario 4 : `ade init --skip-mcp` ignore la config MCP
  - Résultat attendu : la configuration MCP n'est pas appelée

- Scénario 5 : `ade init --skip-tools --skip-skills --skip-mcp` n'exécute rien
  - Résultat attendu : aucune étape appelée, rapport vide

- Scénario 6 : `ade init --halt-on-error` avec échec de la détection d'outils
  - Setup : la détection retourne une erreur
  - Résultat attendu : le setup s'arrête, les étapes suivantes ne sont pas appelées

- Scénario 7 : `ade init` sans `--halt-on-error` avec échec de la détection d'outils
  - Setup : la détection retourne une erreur
  - Résultat attendu : les étapes suivantes (skills, MCP) sont appelées, rapport final affiché

- Scénario 8 : `ade init --help` affiche les nouveaux flags
  - Résultat attendu : la sortie contient "--force", "--output", "--config", "--skip-tools", "--skip-skills", "--skip-mcp", "--halt-on-error"

- Scénario 9 : `ade init specs` fonctionne toujours (non cassé)
  - Résultat attendu : `ade init specs` fonctionne comme avant

- Scénario 10 : `ade init ci` fonctionne toujours (non cassé)
  - Résultat attendu : `ade init ci` fonctionne comme avant

#### Tests E2E — `test/e2e/e2e_test.go` (ajouts, build tag `e2e`)

- Scénario 1 : `TestE2E_AgenticSetup` exécute `ade init` dans un répertoire temporaire
  - Setup : créer un répertoire temporaire, compiler le binaire
  - Exécution : `ade.exe init --force --output {tmpDir}` dans un environnement simulé
  - Résultat attendu : code de sortie 0, rapport affiché

- Scénario 2 : `TestE2E_AgenticSetup_SkipAll` exécute `ade init` avec tous les `--skip-*`
  - Résultat attendu : code de sortie 0, rapport minimal

### Documentation

Voir le contenu attendu ci-dessus pour les fichiers de documentation.

### Exemples d'utilisation
```powershell
# Configuration complète
ade init

# Configuration sans outils
ade init --skip-tools

# Configuration personnalisée
ade init --output C:\projets\mon-app --force

# Mode strict (arrêt à la première erreur)
ade init --halt-on-error

# Voir les options disponibles
ade init --help

# Utilisation avec fichier de config personnalisé
ade init --config C:\config\ade-config.yaml
```
