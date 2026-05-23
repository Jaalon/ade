# Tâche #001 - Story #002 : Système de templates embarqués

## Objectif
Créer le système de templates Go embarqués (`//go:embed`) contenant tous les fichiers de configuration à générer par `ade init specs`, ainsi qu'un chargeur et moteur de rendu de templates.

## Contexte
- Story #002 : `docs/stories/story-002.md`
- Plan d'implémentation : V1 avec templates embarqués (Phase 1). La V2 utilisera l'API des plugins Docker (Story #007).
- Dépend de : Story #001 (structure du projet, package `internal/`)
- Nécessaire pour : Tâches #002, #003, #004, #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer un répertoire `internal/templates/` contenant :
1. Un package Go `templates` avec un système de chargement et rendu de templates
2. Les fichiers de templates embarqués via `//go:embed` pour chaque type de configuration

**Types de templates à fournir :**
- `.gitignore` — Fichier gitignore standard pour un projet Go
- `.opencode/skills/` — Structure des skills (fichiers SKILL.md copiés depuis les skills existants du projet)
- `.opencode/config.json` — Configuration des serveurs MCP
- `.opencode/workflow.yaml` — Workflow de développement décrivant l'ordre des skills

**Cas nominaux :**
- Le package `templates` expose une fonction `Render(templateName string, data interface{}) (string, error)` qui rend un template avec les données fournies
- Le package `templates` expose une fonction `ListTemplates() ([]TemplateInfo, error)` qui retourne la liste des templates disponibles avec leur description et le chemin de destination prévu
- Chaque template a un nom logique unique (ex: `gitignore`, `idea-modules`, `skills-structure`, `mcp-config`, `workflow`)
- Les templates sont rendus avec `text/template` de la stdlib
- Les templates peuvent utiliser des données dynamiques (nom du projet, chemins, etc.)

**Cas limites :**
- Template inexistant → erreur descriptive
- Données manquantes pour le rendu → `text/template` produit une erreur standard
- Fichiers de template vides → retournés tels quels

**Gestion d'erreurs :**
- Template `X` introuvable → erreur `ErrTemplateNotFound`
- Erreur de parsing du template → erreur retournée avec le nom du template et la cause
- Erreur de rendu (exécution) → erreur retournée avec le nom du template et la cause

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/templates/template.go` | Créer | Package templates : Render(), ListTemplates(), types |
| `internal/templates/errors.go` | Créer | Définitions des erreurs sentinelles |
| `internal/templates/embed.go` | Créer | Déclaration `//go:embed` du FS |
| `internal/templates/embed/gitignore.tmpl` | Créer | Template `.gitignore` |
| `internal/templates/embed/opencode/config.json.tmpl` | Créer | Template `.opencode/config.json` (MCP) |
| `internal/templates/embed/opencode/skills/specification-en/SKILL.md` | Créer | Copie du skill existant (contenu statique) |
| `internal/templates/embed/opencode/skills/specification-fr/SKILL.md` | Créer | Copie du skill existant (contenu statique) |
| `internal/templates/embed/opencode/skills/story-en/SKILL.md` | Créer | Copie du skill existant (contenu statique) |
| `internal/templates/embed/opencode/skills/story-fr/SKILL.md` | Créer | Copie du skill existant (contenu statique) |
| `internal/templates/embed/opencode/skills/tasks-fr/SKILL.md` | Créer | Copie du skill existant (contenu statique) |
| `internal/templates/embed/opencode/skills/tasks-en/SKILL.md` | Créer | Copie du skill existant (contenu statique) |
| `internal/templates/embed/opencode/skills/feedback-fr/SKILL.md` | Créer | Copie du skill existant (contenu statique) |
| `internal/templates/embed/opencode/workflow.yaml.tmpl` | Créer | Template workflow de développement |

### Signatures

```go
// internal/templates/errors.go
package templates

var (
    ErrTemplateNotFound = errors.New("template not found")
    ErrTemplateParse    = errors.New("template parse error")
    ErrTemplateRender   = errors.New("template render error")
)

// internal/templates/template.go
package templates

// TemplateInfo décrit un template disponible.
type TemplateInfo struct {
    Name            string // Nom logique unique du template
    Description     string // Description en français
    TargetPath      string // Chemin de destination relatif à la racine du projet (ex: ".gitignore")
    EmbeddedPath    string // Chemin dans l'embed FS (ex: "embed/gitignore.tmpl")
}

// Render rend un template nommé avec les données fournies.
// data peut être nil si le template n'a pas de variables dynamiques.
func Render(name string, data interface{}) (string, error)

// ListTemplates retourne la liste de tous les templates disponibles.
func ListTemplates() []TemplateInfo

// RenderFS est une fonction auxiliaire qui rend un template directement depuis un fs.FS.
func RenderFS(fsys fs.FS, path string, data interface{}) (string, error)
```

### Structure des données

```go
// internal/templates/template.go

// TemplateData contient les variables disponibles pour tous les templates.
type TemplateData struct {
    ProjectName string // Nom du projet (défaut: nom du répertoire)
    GoVersion   string // Version de Go (défaut: "1.26")
    ModulePath  string // Path du module Go (ex: "github.com/user/project")
    Lang        string // Langue pour les skills : "fr" ou "en" (défaut: "fr")
}
```

### Contraintes techniques
- **Go `embed`** : Utiliser `//go:embed internal/templates/embed/*` avec un `embed.FS`. Les fichiers `.tmpl` ne doivent pas être exportés dans le binaire final de manière lisible (embed = compilé dans le binaire).
- **`text/template`** : Utiliser `text/template` de la stdlib (pas `html/template`). Les fonctions de template doivent inclure `sprig` si nécessaire pour des manipulations avancées, sinon se limiter aux fonctions built-in.
- **Package `templates`** : Aucune dépendance externe à part la stdlib (sauf `sprig` optionnel). Le package est purement utilitaire sans état global.
- **`sync.Once`** : Utiliser `sync.Once` pour le parsing paresseux des templates (parser au premier appel, pas au init).
- **Convention de nommage** : Les fichiers de template portent l'extension `.tmpl`. Le chemin dans l'embed reflète la destination (ex: `embed/gitignore.tmpl` pour `.gitignore`).
- **Pas de logique métier** dans ce package : il ne fait que charger, parser et rendre des templates.

### Détails des templates

#### `.gitignore` (embed/gitignore.tmpl)
```gitignore
# Go
*.exe
*.exe~
*.test
*.out
*.test.exe
*.iws
/bin/
/go/
/vendor/
/tmp/
/ade
/ade.exe

# IDE
.idea/
.vscode/
*.iml
*.ipr

# OS
.DS_Store
Thumbs.db
*.swp
*.swo

# Env
.env.local
.env.*.local
```

#### `.opencode/config.json` (embed/opencode/config.json.tmpl)
Template JSON avec les serveurs MCP. Le template reçoit `TemplateData`.

`embed/opencode/config.json.tmpl` :
```json
{
  "mcp_servers": [],
  "skills_path": ".opencode/skills"
}
```

#### `.opencode/workflow.yaml` (embed/opencode/workflow.yaml.tmpl)
```yaml
# Workflow de développement généré par ade init specs
version: "1.0"
name: "Développement standard"
description: "Workflow de développement itératif avec spécification, stories et tâches"
steps:
  - id: specification
    name: "Spécification"
    skill: "specification-{{.Lang}}"
    description: "Définir ou mettre à jour les spécifications du projet"
  - id: stories
    name: "User Stories"
    skill: "story-{{.Lang}}"
    description: "Découper les spécifications en user stories"
  - id: tasks
    name: "Tâches"
    skill: "tasks-{{.Lang}}"
    description: "Découper les stories en tâches exécutables"
  - id: implement
    name: "Implémentation"
    description: "Implémenter les tâches avec l'agent de code"
```

#### Skills SKILL.md
Chaque fichier SKILL.md dans `internal/templates/embed/opencode/skills/` doit être une **copie exacte** du fichier correspondant dans `.opencode/skills/` à la racine du projet.
Ces fichiers n'ont pas de variables dynamiques (ce sont des copies statiques). L'agent doit :

1. Lire chaque fichier dans `.opencode/skills/` (spécification-en, spécification-fr, story-en, story-fr, tasks-en, tasks-fr, feedback-fr)
2. Copier le contenu exact dans `internal/templates/embed/opencode/skills/<name>/SKILL.md`
3. Ces fichiers n'ont pas l'extension `.tmpl` car ils ne contiennent pas de variables de template

**Fichiers sources à copier (lus depuis la racine du projet) :**
- `.opencode/skills/specification-en/SKILL.md` → `internal/templates/embed/opencode/skills/specification-en/SKILL.md`
- `.opencode/skills/specification-fr/SKILL.md` → `internal/templates/embed/opencode/skills/specification-fr/SKILL.md`
- `.opencode/skills/story-en/SKILL.md` → `internal/templates/embed/opencode/skills/story-en/SKILL.md`
- `.opencode/skills/story-fr/SKILL.md` → `internal/templates/embed/opencode/skills/story-fr/SKILL.md`
- `.opencode/skills/tasks-fr/SKILL.md` → `internal/templates/embed/opencode/skills/tasks-fr/SKILL.md`
- `.opencode/skills/tasks-en/SKILL.md` → `internal/templates/embed/opencode/skills/tasks-en/SKILL.md`
- `.opencode/skills/feedback-fr/SKILL.md` → `internal/templates/embed/opencode/skills/feedback-fr/SKILL.md`

Ne pas modifier le contenu — recopier textuellement.

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/templates/template_test.go`
- Scénario 1 : `ListTemplates()` retourne au moins 6 templates (gitignore, mcp-config, workflow + 3+ skills)
  - Données : appel à `ListTemplates()`
  - Résultat attendu : slice non-nil avec au moins 6 éléments
- Scénario 2 : `Render("gitignore", nil)` retourne un `.gitignore` valide contenant "*.exe"
  - Résultat attendu : la sortie contient "*.exe" et "*.test"
- Scénario 3 : `Render("mcp-config", nil)` retourne un JSON valide contenant `mcp_servers`
  - Résultat attendu : la sortie est un JSON valide
- Scénario 4 : `Render("inexistant", nil)` retourne `ErrTemplateNotFound`
  - Résultat attendu : l'erreur match `ErrTemplateNotFound`
- Scénario 5 : `Render("workflow", TemplateData{Lang: "fr"})` contient "tasks-fr" dans la sortie
  - Résultat attendu : la sortie YAML contient "tasks-fr"
- Scénario 6 : Tous les templates se parsent sans erreur (test Table-driven)
  - Données : pour chaque template de `ListTemplates()`
  - Résultat attendu : `Render(t.Name, defaultData)` ne retourne pas d'erreur de parsing

### Documentation
- Aucune documentation spécifique pour cette tâche (la documentation du système de templates sera couverte par la Tâche #005)

### Exemples d'utilisation
```go
// Exemple d'utilisation du package templates
templates := templates.ListTemplates()
for _, t := range templates {
    log.Printf("Template: %s → %s", t.Name, t.TargetPath)
}

// Rendu d'un template avec données
data := templates.TemplateData{
    ProjectName: "mon-projet",
    GoVersion:   "1.26",
    ModulePath:  "github.com/user/mon-projet",
}
content, err := templates.Render("gitignore", data)
if err != nil {
    return err
}
fmt.Println(content)
```
