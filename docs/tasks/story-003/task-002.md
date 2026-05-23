# Tâche #002 - Story #003 : Gestion des skills OpenCode

## Objectif
Implémenter la gestion des skills OpenCode : installation, référencement et vérification des skills disponibles dans `.opencode/skills/`, en réutilisant le système de templates embarqués existant (`internal/templates/`).

## Contexte
- Story #003 : `docs/stories/story-003.md`
- Dépend de : Tâche #001 (types partagés), Story #002 (système de templates existant)
- Nécessaire pour : Tâche #005 (intégration dans la commande `ade init`)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer la gestion des skills OpenCode dans le répertoire `.opencode/skills/` du projet.

**Vérification des skills existants :**
1. Scanner le répertoire `.opencode/skills/` à la racine du projet
2. Lister les skills déjà présents (chaque sous-répertoire contenant un fichier `SKILL.md`)
3. Comparer avec la liste des skills disponibles dans les templates embarqués

**Installation des skills manquants :**
1. Pour chaque skill présent dans les templates embarqués mais pas dans `.opencode/skills/` :
   - Copier le fichier `SKILL.md` depuis les templates vers `.opencode/skills/<name>/SKILL.md`
2. Utiliser le système `templates.Render()` ou `templates.RenderFS()` pour cela
3. Ne pas écraser les skills existants (sauf si `--force` est actif)

**Référencement des skills :**
1. Vérifier que les skills sont correctement référencés dans `.opencode/config.json` (section `skills_path`)
2. Si `.opencode/config.json` n'existe pas, il sera créé par `ade init specs` (Story #002)
3. Lister les skills disponibles avec leur description

**Rapport :**
- Retourner un rapport listant : skills créés, skills déjà présents, skills ignorés, erreurs

**Cas nominaux :**
- `agentic.EnsureSkills(ctx, outputDir, force)` installe les skills manquants
- Skills déjà présents → ignorés (sauf `--force`)
- Rapport retourné avec le détail de chaque skill

**Cas limites :**
- `.opencode/skills/` n'existait pas → création automatique du répertoire
- Skills partiellement installés (certains dossiers existent, d'autres non) → installation partielle
- Template d'un skill corrompu → erreur dans le rapport pour ce skill, continue avec les autres

**Gestion d'erreurs :**
- Erreur d'écriture d'un skill → rapport avec erreur, continue avec les suivants
- Template introuvable pour un skill → rapport avec erreur
- Aucune erreur fatale

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/agentic/skills.go` | Créer | Gestion des skills : EnsureSkills(), ListMissing() |
| `internal/agentic/skills_test.go` | Créer | Tests unitaires |

### Signatures

```go
// internal/agentic/skills.go
package agentic

import "automated_dev_environment/internal/templates"

// SkillInfo décrit un skill disponible ou installé.
type SkillInfo struct {
	Name        string // Nom logique du skill (ex: "specification-en")
	Description string // Description du skill
	Installed   bool   // true si le fichier SKILL.md existe dans le projet
	TargetPath  string // Chemin de destination relatif (ex: ".opencode/skills/specification-en/SKILL.md")
}

// SkillsReport contient le rapport de l'installation des skills.
type SkillsReport struct {
	Installed   []string    // Skills installés avec succès
	AlreadyExist []string   // Skills déjà présents
	Skipped     []string    // Skills ignorés (déjà existants sans force)
	Errors      []SkillError // Erreurs rencontrées
	Total       int         // Nombre total de skills gérés
}

// SkillError décrit une erreur pour un skill spécifique.
type SkillError struct {
	Name string
	Err  error
}

// EnsureSkills installe les skills manquants dans le répertoire de sortie.
// force = true écrase les skills existants.
// Retourne un rapport complet (toujours non-nil).
func EnsureSkills(ctx context.Context, outputDir string, force bool) (*SkillsReport, error)

// ListProjectSkills liste les skills présents dans le répertoire projet.
// outputDir est la racine du projet (ex: ".").
func ListProjectSkills(outputDir string) ([]SkillInfo, error)

// ListAvailableSkills retourne les skills disponibles dans les templates.
func ListAvailableSkills() ([]SkillInfo, error)

// MissingSkills compare les skills disponibles et ceux installés,
// et retourne la liste des skills manquants à installer.
func MissingSkills(outputDir string) ([]SkillInfo, error)
```

### Contraintes techniques
- **Réutilisation** : Utiliser `templates.ListTemplates()` pour obtenir la liste des skill templates (ceux dont le nom commence par `skill-`)
- **Filtrage** : Filtrer les templates avec `strings.HasPrefix(tmpl.Name, "skill-")` pour identifier les skills
- **`os.Stat`** : Vérifier l'existence de chaque fichier `SKILL.md` avec `os.Stat(filepath.Join(outputDir, tmpl.TargetPath))`
- **Écriture** : Utiliser `os.WriteFile` pour créer les fichiers SKILL.md
- **Rapport structuré** : Similaire au pattern du package `generator` (avec `Report`)
- **Contexte** : `EnsureSkills()` accepte un `context.Context` pour permettre l'annulation
- **Chemins** : Utiliser `filepath.Join()` pour tous les chemins de fichier

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/agentic/skills_test.go`

- Scénario 1 : `ListAvailableSkills()` retourne au moins 6 skills
  - Résultat attendu : slice non-nil, au moins 6 entrées avec `Name` commençant par "skill-"

- Scénario 2 : `ListProjectSkills()` dans un répertoire vide
  - Setup : répertoire temporaire vide
  - Résultat attendu : slice vide, pas d'erreur

- Scénario 3 : `MissingSkills()` dans un répertoire vide
  - Setup : répertoire temporaire vide
  - Résultat attendu : retourne tous les skills disponibles

- Scénario 4 : `MissingSkills()` avec tous les skills déjà installés
  - Setup : créer `.opencode/skills/` avec tous les fichiers SKILL.md
  - Résultat attendu : slice vide (aucun manquant)

- Scénario 5 : `EnsureSkills()` dans un répertoire vide, sans force
  - Setup : répertoire temporaire vide
  - Résultat attendu : `Installed` non vide, tous les skills créés

- Scénario 6 : `EnsureSkills()` avec skills déjà existants, sans force
  - Setup : créer un skill existant, appeler EnsureSkills
  - Résultat attendu : `AlreadyExist` contient le skill existant

- Scénario 7 : `EnsureSkills()` avec skills déjà existants, avec force=true
  - Setup : créer un skill existant, appeler EnsureSkills avec force=true
  - Résultat attendu : `Installed` contient le skill (réinstallé)

### Documentation
- Aucune documentation spécifique (documentation globale dans Tâche #005)

### Exemples d'utilisation
```go
// Lister les skills disponibles
skills, _ := agentic.ListAvailableSkills()
for _, s := range skills {
    fmt.Printf("- %s: %s\n", s.Name, s.Description)
}

// Installer les skills manquants
report, err := agentic.EnsureSkills(context.Background(), ".", false)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Installés: %d, Déjà présents: %d, Erreurs: %d\n",
    len(report.Installed), len(report.AlreadyExist), len(report.Errors))
```
