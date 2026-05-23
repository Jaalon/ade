# Tâche #003 - Story #003 : Configuration des serveurs MCP

## Objectif
Implémenter la configuration des serveurs MCP (Model Context Protocol) dans `.opencode/config.json`, en intégrant les définitions provenant du package `config` partagé (`internal/config/config.go`), et en préservant la structure existante générée par `ade init specs`.

## Contexte
- Story #003 : `docs/stories/story-003.md`
- Dépend de : Tâche #001 (package `internal/config` partagé pour la section `mcp_servers`)
- Nécessaire pour : Tâche #004 (intégration dans la commande `ade init`)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer un service de configuration des serveurs MCP qui utilise le package `config` (`internal/config/config.go`) pour lire les définitions de serveurs MCP et met à jour `.opencode/config.json`.

**Sources de configuration MCP :**
1. Le package `internal/config/config.go` expose `config.AgenticConfig.MCPServers` (défini dans la section `mcp_servers` du fichier YAML)
2. Utiliser `config.LoadConfig(configPath)` pour charger la config (auto-détection si `configPath` est vide)
3. Si le fichier `.opencode/config.json` existe déjà (généré par `ade init specs`), le lire et fusionner les serveurs MCP

**Fusion :**
- Les serveurs MCP définis dans `config.AgenticConfig.MCPServers` sont ajoutés ou écrasent ceux existant dans `.opencode/config.json`
- La comparaison se fait par `name` (nom du serveur)
- Les serveurs existants dans `.opencode/config.json` mais absents de la config YAML sont conservés
- La structure de base `skills_path: ".opencode/skills"` est maintenue

**Rapport :**
- Retourner un rapport listant les serveurs ajoutés, mis à jour, inchangés et les erreurs

**Cas nominaux :**
- `agentic.ConfigureMCPServers(ctx, outputDir, configPath)` lit la config YAML via `config.LoadConfig()` et met à jour `.opencode/config.json`
- Serveurs MCP dans la config YAML → ajoutés/écrasés dans `.opencode/config.json`
- `.opencode/config.json` inexistant → créé avec les serveurs de la config YAML
- Aucun serveur MCP dans la config YAML → pas de modification, `.opencode/config.json` conservé tel quel

**Cas limites :**
- `.opencode/config.json` invalide (JSON mal formé) → avertir et recréer depuis la config YAML
- Variables d'environnement dans les args (`${VAR}`) → ne pas résoudre, laisser telles quelles (résolution par OpenCode/Cursor)
- Chemin de config YAML spécifié mais fichier inexistant → auto-détection via `config.FindConfigPath()`

**Gestion d'erreurs :**
- Erreur de lecture/écriture de `.opencode/config.json` → rapport avec erreur
- Erreur de parsing JSON de `.opencode/config.json` existant → recréation avec avertissement
- `config.LoadConfig()` retourne une config vide (pas d'erreur) si aucun fichier trouvé
- Aucune erreur fatale, rapport complet retourné

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/agentic/mcp.go` | Créer | Configuration MCP : ConfigureMCPServers() |
| `internal/agentic/mcp_test.go` | Créer | Tests unitaires |

### Signatures

```go
// internal/agentic/mcp.go
package agentic

import "automated_dev_environment/internal/config"

// MCPConfig représente la structure de .opencode/config.json.
type MCPConfig struct {
	MCPServers []config.MCPServer `json:"mcp_servers"`
	SkillsPath string             `json:"skills_path"`
}

// MCPOptions contient les paramètres de configuration MCP.
type MCPOptions struct {
	OutputDir  string // Répertoire racine du projet (défaut: ".")
	ConfigPath string // Chemin explicite du fichier de config YAML (vide = auto-détection)
}

// MCPResult décrit le résultat pour un serveur MCP.
type MCPResult struct {
	ServerName string // Nom du serveur MCP
	Action     string // "added", "updated", "unchanged", "error"
	Error      error  // Erreur éventuelle
}

// MCPReport contient le rapport de configuration MCP.
type MCPReport struct {
	Results   []MCPResult
	Added     int
	Updated   int
	Unchanged int
	Errors    int
}

// ConfigureMCPServers configure les serveurs MCP dans .opencode/config.json.
// Utilise config.LoadConfig() pour charger les définitions MCP depuis le YAML.
func ConfigureMCPServers(ctx context.Context, opts MCPOptions) (*MCPReport, error)

// loadOpenCodeConfig lit le fichier .opencode/config.json existant.
func loadOpenCodeConfig(projectDir string) (*MCPConfig, error)

// saveOpenCodeConfig écrit le fichier .opencode/config.json.
func saveOpenCodeConfig(projectDir string, config *MCPConfig) error

// mergeMCPServers fusionne les serveurs MCP : les nouveaux écrasent les existants.
func mergeMCPServers(existing, incoming []config.MCPServer) []config.MCPServer
```

### Contraintes techniques
- **JSON** : Utiliser `encoding/json` de la stdlib (pas de dépendance externe)
- **Config partagée** : Utiliser `config.LoadConfig()` et `config.MCPServer` du package `internal/config` (Tâche #001)
- **`os.ReadFile` / `os.WriteFile`** : Lire et écrire les fichiers JSON
- **Indentation JSON** : Utiliser `json.MarshalIndent(config, "", "  ")` pour une sortie lisible
- **Fusion par nom** : La clé de fusion est le `Name` du serveur MCP (comparaison insensible à la casse)
- **Préservation** : Les propriétés des serveurs existants non redéfinies dans la config YAML sont conservées. L'écrasement est complet par serveur (si le nom correspond, tout le serveur est remplacé)
- **Création du répertoire** : Si `.opencode/` n'existe pas, le créer avec `os.MkdirAll`
- **`filepath.Join`** : Utiliser pour tous les chemins

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/agentic/mcp_test.go`

- Scénario 1 : `mergeMCPServers()` avec liste vide de nouveaux serveurs
  - Données : existing = `[{Name: "fs", ...}]`, incoming = `[]`
  - Résultat attendu : `[{Name: "fs", ...}]` (inchangé)

- Scénario 2 : `mergeMCPServers()` avec nouveau serveur
  - Données : existing = `[{Name: "fs"}]`, incoming = `[{Name: "github"}]`
  - Résultat attendu : 2 serveurs, les deux présents

- Scénario 3 : `mergeMCPServers()` avec écrasement
  - Données : existing = `[{Name: "fs", Command: "old"}]`, incoming = `[{Name: "fs", Command: "new"}]`
  - Résultat attendu : `[{Name: "fs", Command: "new"}]` (écrasé)

- Scénario 4 : `ConfigureMCPServers()` dans un répertoire vide
  - Setup : répertoire temporaire vide, config YAML avec 2 serveurs MCP
  - Résultat attendu : `Report.Added == 2`, `.opencode/config.json` créé avec les 2 serveurs

- Scénario 5 : `ConfigureMCPServers()` avec `.opencode/config.json` existant et serveur supplémentaire
  - Setup : créer `.opencode/config.json` avec 1 serveur, config YAML avec 1 serveur différent
  - Résultat attendu : `Report.Added == 1`, `Report.Unchanged == 1`

- Scénario 6 : `ConfigureMCPServers()` avec `.opencode/config.json` existant et serveur écrasé
  - Setup : créer `.opencode/config.json` avec 1 serveur, config YAML avec le même serveur modifié
  - Résultat attendu : `Report.Updated == 1`

- Scénario 7 : `ConfigureMCPServers()` sans config YAML (aucune trouvée)
  - Setup : répertoire vide sans fichier YAML
  - Résultat attendu : `Report` vide, `.opencode/config.json` non modifié

- Scénario 8 : `loadOpenCodeConfig()` avec JSON invalide
  - Setup : créer `.opencode/config.json` avec `{invalid json}`
  - Résultat attendu : erreur retournée

### Documentation
- Aucune documentation spécifique (documentation globale dans Tâche #004)

### Exemples d'utilisation
```go
// Configuration MCP standard (utilise config.LoadConfig en interne)
report, err := agentic.ConfigureMCPServers(context.Background(), agentic.MCPOptions{
    OutputDir:  ".",
    ConfigPath: "", // auto-détection du fichier YAML
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Ajoutés: %d, Mis à jour: %d, Inchangés: %d\n",
    report.Added, report.Updated, report.Unchanged)
```
