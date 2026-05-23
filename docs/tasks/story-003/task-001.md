# Tâche #001 - Story #003 : Détection des outils agentic et configuration YAML

## Objectif
Créer le système de détection des outils agentic (OpenCode, Cursor) sur Windows, avec support des chemins par défaut, recherche dans le PATH, surcharge via configuration YAML, et messages d'installation pour les outils manquants.

## Contexte
- Story #003 : `docs/stories/story-003.md`
- Spécification : `docs/specification/specification.md`
- Dépend de : Story #001, Story #002 (structure du projet, templates embarqués)
- Nécessaire pour : Tâche #004 (intégration dans la commande `ade init`)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

#### Détection des outils
Créer un package `internal/agentic/` contenant la logique de détection des outils agentic.

**Détection d'OpenCode :**
1. Vérifier le PATH avec `exec.LookPath("opencode")`
2. Vérifier les chemins d'installation Windows par défaut (dans l'ordre) :
   - `%LOCALAPPDATA%\Programs\opencode\opencode.exe`
   - `%USERPROFILE%\.opencode\opencode.exe`
   - `%PROGRAMFILES%\opencode\opencode.exe`
   - `%PROGRAMFILES(x86)%\opencode\opencode.exe`

**Détection de Cursor :**
1. Vérifier le PATH avec `exec.LookPath("cursor")`
2. Vérifier les chemins d'installation Windows par défaut (dans l'ordre) :
   - `%LOCALAPPDATA%\Programs\cursor\cursor.exe`
   - `%USERPROFILE%\AppData\Local\cursor\cursor.exe`
   - `%PROGRAMFILES%\cursor\cursor.exe`

**Configuration YAML :**
- La lecture du fichier de configuration YAML est centralisée dans `internal/config/config.go` (package partagé `config`) 
- Le package `config` expose une fonction `LoadConfig(path string) (*AgenticConfig, error)` qui lit et parse le fichier YAML
- Le package `config` gère l'auto-détection du fichier de config (cherche dans : `./ade-config.yaml`, `./.ade.yaml`, `%USERPROFILE%\.ade\config.yaml`)
- Si un fichier `ade-config.yaml` ou `.ade.yaml` existe, lire les surcharges de chemins d'outils via le package `config`
- La section `tools` du YAML peut contenir :
  ```yaml
  tools:
    opencode:
      path: "C:\\custom\\path\\opencode.exe"
    cursor:
      path: "C:\\custom\\path\\cursor.exe"
  ```

**Messages pour outils manquants :**
- Si OpenCode n'est pas trouvé → message avec URL d'installation : "OpenCode n'est pas installé. Téléchargez-le depuis https://opencode.ai"
- Si Cursor n'est pas trouvé → message : "Cursor n'est pas installé. Téléchargez-le depuis https://cursor.com"
- Les messages doivent être clairs, en français, et inclure à la fois l'URL et la suggestion d'ajout au PATH

**Cas nominaux :**
- `agentic.DetectTools(ctx, configPath)` retourne une slice de `ToolInfo` avec tous les outils trouvés
- OpenCode trouvé dans le PATH → `Found: true, Path: "opencode"`
- OpenCode trouvé via un chemin par défaut → `Found: true, Path: "C:\Users\...\opencode.exe"`
- Chemin surchargé dans le YAML → le chemin YAML est prioritaire devant PATH et chemins par défaut
- Fichier YAML inexistant → pas d'erreur, utilise PATH + chemins par défaut

**Cas limites :**
- Aucun outil installé → slice avec tous les `Found: false`, messages d'installation disponibles
- Un seul outil installé (ex: OpenCode mais pas Cursor) → OpenCode trouvé, Cursor non trouvé
- Chemin YAML valide mais binaire inexistant (fichier supprimé) → `Found: false`, message d'erreur sur le chemin configuré
- PATH vide ou inexistant → recherche uniquement dans les chemins par défaut Windows
- Utilisateur sans `%LOCALAPPDATA%` ou variable d'environnement manquante → ignorer ce chemin

**Gestion d'erreurs :**
- Erreur de parsing du YAML → loguer l'erreur et ignorer la config YAML, continuer avec les valeurs par défaut
- Erreur de lecture du fichier YAML (permissions) → message d'avertissement, continuer normalement
- Aucune erreur fatale : `DetectTools()` retourne toujours une slice, même si tous les outils sont manquants

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/config/config.go` | Créer | Lecture centralisée de la config YAML (tools + mcp_servers) |
| `internal/config/config_test.go` | Créer | Tests de la lecture YAML |
| `internal/agentic/types.go` | Créer | Types ToolType, ToolInfo (utilise `config.AgenticConfig`) |
| `internal/agentic/detector.go` | Créer | Détection des outils (OpenCode, Cursor) |
| `internal/agentic/detector_test.go` | Créer | Tests unitaires de détection |

### Signatures

```go
// internal/config/config.go
package config

// ToolType représente le type d'outil agentic.
type ToolType string

const (
	ToolOpenCode ToolType = "opencode"
	ToolCursor   ToolType = "cursor"
)

// AgenticConfig représente la configuration YAML complète.
type AgenticConfig struct {
	Tools      map[ToolType]ToolConfig `yaml:"tools"`
	MCPServers []MCPServer            `yaml:"mcp_servers"`
}

// ToolConfig représente la configuration YAML pour un outil.
type ToolConfig struct {
	Path string `yaml:"path"`
}

// MCPServer décrit un serveur MCP dans la config YAML.
type MCPServer struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

// LoadConfig charge la configuration depuis un fichier YAML.
// Si path est vide, cherche automatiquement (./ade-config.yaml, ./.ade.yaml, etc.).
// Retourne une config vide (sans erreur) si aucun fichier n'est trouvé.
func LoadConfig(path string) (*AgenticConfig, error)

// FindConfigPath cherche automatiquement un fichier de configuration.
// Cherche dans : ./ade-config.yaml, ./.ade.yaml, %USERPROFILE%\.ade\config.yaml
func FindConfigPath() string

// internal/agentic/types.go
package agentic

// ToolInfo décrit le résultat de détection pour un outil.
type ToolInfo struct {
	Type    config.ToolType // Type d'outil
	Name    string          // Nom affichable (ex: "OpenCode", "Cursor")
	Path    string          // Chemin complet du binaire (vide si non trouvé)
	Found   bool            // true si l'outil a été trouvé
	Version string          // Version détectée (peut être vide)
}

// DetectionResult regroupe tous les résultats de détection.
type DetectionResult struct {
	Tools      []ToolInfo           // Liste de tous les outils détectés
	Config     *config.AgenticConfig // Configuration YAML lue (peut être vide)
	ConfigPath string               // Chemin du fichier de config utilisé (vide si aucun)
}

// InstallInfo retourne les instructions d'installation pour un outil non trouvé.
type InstallInfo struct {
	Tool    config.ToolType
	URL     string
	Message string
}

// internal/agentic/detector.go

// DetectTools détecte les outils agentic installés sur le système.
// Utilise internal/config pour charger la config YAML.
func DetectTools(ctx context.Context, configPath string) (*DetectionResult, error)

// FindTool cherche un outil spécifique en utilisant PATH + chemins par défaut
// + surcharge YAML (si fournie dans config).
func FindTool(tool config.ToolType, cfg *config.AgenticConfig) ToolInfo

// defaultToolPaths retourne les chemins d'installation par défaut pour un outil.
func defaultToolPaths(tool config.ToolType) []string

// InstallInstructions retourne les instructions d'installation pour un outil.
func InstallInstructions(tool config.ToolType) InstallInfo
```

### Contraintes techniques
- **`os/exec`** : Utiliser `exec.LookPath()` pour la recherche dans le PATH (variable `execLookPath` pour permettre le mock dans les tests)
- **`gopkg.in/yaml.v3`** : Déjà dans les dépendances transitives, mais à ajouter comme dépendance directe si nécessaire. Sinon, utiliser `encoding/json` ou un parser manuel. Préférer `gopkg.in/yaml.v3` comme dépendance explicite.
- **Variables d'environnement** : Utiliser `os.Getenv()` pour `LOCALAPPDATA`, `USERPROFILE`, `PROGRAMFILES`, `PROGRAMFILES(X86)`. Ces variables existent sur Windows.
- **Chemins Windows** : Utiliser `filepath.Join()` et `filepath.Clean()` pour normaliser les chemins. Les chemins par défaut doivent être testés avec `os.Stat()`.
- **Mock** : Rendre `os.Stat` et `exec.LookPath` mockables via des variables de paquet (pattern `var execLookPath = exec.LookPath`)
- **Context** : Les fonctions acceptent `context.Context` pour permettre l'annulation et le timeout
- **Pas de logique métier** dans ce package : pure détection et configuration, pas d'installation réelle
- **Messages en français** : Tous les messages utilisateur sont en français

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/agentic/detector_test.go`

- Scénario 1 : `FindTool` avec OpenCode dans le PATH
  - Setup : Mocker `execLookPath` pour retourner `"opencode"` pour "opencode"
  - Résultat attendu : `Found: true`, `Path: "opencode"`

- Scénario 2 : `FindTool` avec OpenCode via chemin par défaut
  - Setup : Mocker `execLookPath` pour retourner une erreur. Mocker `os.Stat` pour réussir sur `%LOCALAPPDATA%\Programs\opencode\opencode.exe`
  - Résultat attendu : `Found: true`, `Path` contient `opencode.exe`

- Scénario 3 : `FindTool` avec surcharge YAML
  - Setup : `AgenticConfig{Tools: {ToolOpenCode: {Path: "C:\\custom\\opencode.exe"}}}`, mocker `os.Stat` pour réussir
  - Résultat attendu : `Found: true`, `Path: "C:\\custom\\opencode.exe"`

- Scénario 4 : `FindTool` avec outil non trouvé
  - Setup : Tous les `execLookPath` et `os.Stat` retournent des erreurs
  - Résultat attendu : `Found: false`, `Path: ""`

- Scénario 5 : `InstallInstructions` pour OpenCode
  - Résultat attendu : `URL == "https://opencode.ai"`, Message non vide en français

- Scénario 6 : `InstallInstructions` pour Cursor
  - Résultat attendu : `URL == "https://cursor.com"`, Message non vide en français

- Scénario 7 : `readConfig` avec fichier YAML valide
  - Données : Fichier temporaire avec `tools: {opencode: {path: "test.exe"}}`
  - Résultat attendu : `config.Tools[ToolOpenCode].Path == "test.exe"`

- Scénario 8 : `readConfig` avec fichier inexistant
  - Résultat attendu : config vide, pas d'erreur

- Scénario 9 : `readConfig` avec YAML invalide
  - Données : Fichier temporaire avec contenu invalide
  - Résultat attendu : config vide, pas d'erreur (graceful degradation)

- Scénario 10 : `DetectTools` complet avec les deux outils trouvés
  - Résultat attendu : `len(result.Tools) == 2`, au moins un `Found: true`

- Scénario 11 : `DetectTools` avec config YAML contenant des chemins valides
  - Résultat attendu : les chemins YAML sont utilisés

### Documentation
- Aucune documentation spécifique pour cette tâche (documentation globale dans Tâche #005)

### Exemples d'utilisation
```go
// Détection automatique
result, err := agentic.DetectTools(context.Background(), "")
if err != nil {
    log.Printf("Warning: %v", err)
}
for _, tool := range result.Tools {
    if tool.Found {
        fmt.Printf("✓ %s trouvé: %s\n", tool.Name, tool.Path)
    } else {
        info := agentic.InstallInstructions(tool.Type)
        fmt.Printf("✗ %s: %s\n", tool.Name, info.Message)
    }
}

// Détection avec config personnalisée
result, _ = agentic.DetectTools(context.Background(), "./ade-config.yaml")
```
