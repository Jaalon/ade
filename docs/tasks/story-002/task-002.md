# Tâche #002 - Story #002 : Service de génération de fichiers

## Objectif
Créer le service `generator` qui utilise le système de templates embarqués (Tâche #001) pour générer les fichiers de configuration sur le disque, avec une politique de non-écrasement des fichiers existants et confirmation utilisateur.

## Contexte
- Story #002 : `docs/stories/story-002.md`
- Dépend de : Tâche #001 (système de templates)
- Nécessaire pour : Tâche #003 (commande `ade init specs`)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer un package `internal/generator/` qui :
1. Récupère la liste des templates depuis `internal/templates`
2. Pour chaque template, rend le contenu avec les données du projet
3. Vérifie si le fichier de destination existe déjà sur le disque
4. Si le fichier existe et `--force` n'est pas passé, demande confirmation à l'utilisateur
5. Écrit le fichier si confirmé ou si `--force` est actif
6. Fournit un rapport des fichiers générés, ignorés, et des erreurs

**Cas nominaux :**
- `generator.Generate(ctx, opts)` génère tous les fichiers de configuration dans le répertoire de travail courant
- Les fichiers générés respectent la structure : `.gitignore`, `.idea/`, `.opencode/`
- Les répertoires intermédiaires sont créés automatiquement
- Un rapport est retourné listant chaque fichier avec son statut (created, skipped, error)

**Cas limites :**
- Fichier déjà existant → demande confirmation interactive (sauf si `Force` est true)
- Répertoire de destination inexistant → création automatique
- Template rendu vide → avertissement, fichier non créé
- Permissions insuffisantes → erreur retournée dans le rapport

**Gestion d'erreurs :**
- Erreur d'écriture fichier `X` → rapport avec erreur, continue avec les fichiers suivants
- Erreur de rendu template `Y` → rapport avec erreur, continue avec les fichiers suivants
- Aucune erreur fatale : `Generate()` retourne toujours un rapport complet

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/generator/generator.go` | Créer | Package generator : Generate(), FileResult, Report |
| `internal/generator/options.go` | Créer | Options de génération (Force, OutputDir, TemplateData, etc.) |
| `internal/generator/prompter.go` | Créer | Interface de confirmation utilisateur (pour permettre le mock dans les tests) |

### Signatures

```go
// internal/generator/generator.go
package generator

import (
    "context"
    "automated_dev_environment/internal/templates"
)

// FileStatus représente le statut d'un fichier après génération.
type FileStatus string

const (
    StatusCreated FileStatus = "created"
    StatusSkipped FileStatus = "skipped"
    StatusError   FileStatus = "error"
    StatusOverwritten FileStatus = "overwritten"
)

// FileResult décrit le résultat pour un fichier généré.
type FileResult struct {
    TemplateName string     // Nom du template
    TargetPath   string     // Chemin de destination (relatif)
    Status       FileStatus // Statut de la génération
    Error        error      // Erreur éventuelle (nil si succès)
}

// Report contient le rapport complet de la génération.
type Report struct {
    Files    []FileResult
    Success  int // Nombre de fichiers créés avec succès
    Skipped  int // Nombre de fichiers ignorés (existant + non confirmé)
    Errors   int // Nombre d'erreurs
}

// Generate exécute la génération des fichiers de configuration.
// Elle rend chaque template, vérifie les existants, et écrit les fichiers.
func Generate(ctx context.Context, opts Options) (*Report, error)

// internal/generator/options.go

// Options contient les paramètres de génération.
type Options struct {
    // OutputDir est le répertoire de destination (défaut: répertoire courant)
    OutputDir string

    // Force force l'écrasement des fichiers existants sans confirmation
    Force bool

    // TemplateData contient les variables pour le rendu des templates
    TemplateData templates.TemplateData

    // Prompter est l'interface pour la confirmation utilisateur.
    // Si nil, utilise l'implémentation standard (os.Stdin/os.Stdout).
    Prompter Prompter

    // TemplatesFilter permet de filtrer les templates à générer.
    // Si nil, tous les templates sont générés.
    TemplatesFilter []string
}

// internal/generator/prompter.go

// Prompter est l'interface pour la confirmation utilisateur.
// Permet le mock dans les tests.
type Prompter interface {
    // Confirm pose une question oui/non à l'utilisateur.
    // Retourne true si l'utilisateur répond oui.
    Confirm(message string) (bool, error)
}

// StdPrompter est l'implémentation standard qui lit sur os.Stdin.
type StdPrompter struct{}

func (p *StdPrompter) Confirm(message string) (bool, error)
```

### Contraintes techniques
- **`os.MkdirAll`** : Créer les répertoires parents avec `os.MkdirAll(dir, 0755)` avant d'écrire chaque fichier
- **`os.WriteFile`** : Écrire les fichiers avec `os.WriteFile(path, []byte(content), 0644)`
- **Évitement d'écrasement** : Utiliser `os.Stat()` pour vérifier l'existence avant d'écrire. Si le fichier existe et `opts.Force` est faux, utiliser `opts.Prompter.Confirm()` pour demander confirmation.
- **Atomicité** : Chaque fichier est écrit indépendamment. Une erreur sur un fichier n'empêche pas la génération des autres.
- **Rapport structuré** : Le `Report` doit être utilisable par la commande Cobra pour afficher un résumé clair.
- **Chemins** : Tous les TargetPath des templates doivent être combinés avec `filepath.Join(opts.OutputDir, ...)` pour supporter les chemins Windows.
- **Context** : `Generate()` accepte un `context.Context` pour permettre l'annulation. Utiliser `select` + `ctx.Done()` dans les boucles longues si nécessaire.

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/generator/generator_test.go`
- Scénario 1 : `Generate()` avec répertoire temporaire vide crée tous les fichiers
  - Données : `Options{OutputDir: tmpDir, Force: true}`
  - Résultat attendu : `Report.Success > 0`, tous les fichiers existent
- Scénario 2 : `Generate()` sans `Force` avec un fichier existant → demande confirmation
  - Données : `Options{OutputDir: tmpDir, Prompter: mockPrompt(false)}` (mock qui répond non)
  - Résultat attendu : le fichier existant est dans `Report.Skipped`
- Scénario 3 : `Generate()` avec `Force=true` écrase les fichiers existants
  - Données : `Options{OutputDir: tmpDir, Force: true}` avec fichier pré-existant
  - Résultat attendu : `Report.Success` inclut le fichier écrasé
- Scénario 4 : `Generate()` avec `TemplatesFilter` ne génère que les templates filtrés
  - Données : `Options{OutputDir: tmpDir, Force: true, TemplatesFilter: []string{"gitignore"}}`
  - Résultat attendu : seul `.gitignore` est créé
- Scénario 5 : `Generate()` avec `OutputDir` pointant vers un fichier existant retourne une erreur
  - Données : créer un fichier régulier, l'utiliser comme OutputDir
  - Résultat attendu : Generate() retourne une erreur (ou Report.Errors > 0)

### Documentation
- Aucune documentation spécifique pour cette tâche

### Exemples d'utilisation
```go
// Exemple d'utilisation basique
opts := generator.Options{
    Force: true,
    TemplateData: templates.TemplateData{
        ProjectName: "mon-app",
    },
}
report, err := generator.Generate(context.Background(), opts)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Créés: %d, Ignorés: %d, Erreurs: %d\n",
    report.Success, report.Skipped, report.Errors)

// Exemple avec confirmation interactive
opts2 := generator.Options{
    Force:     false,
    Prompter:  &generator.StdPrompter{},
}
report2, _ := generator.Generate(context.Background(), opts2)
for _, f := range report2.Files {
    fmt.Printf("- %s: %s\n", f.TargetPath, f.Status)
}
```
