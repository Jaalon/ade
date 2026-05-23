# Tâche #002 - Story #001 : Commande racine Cobra avec --help

## Objectif
Implémenter la commande racine `ade` avec Cobra, incluant l'affichage de l'aide (`--help`), le support de `--version`, et l'architecture des sous-commandes.

## Contexte
- Story #001 : `docs/stories/story-001.md`
- Dépend de : Tâche #001
- Nécessaire pour : Tâches #003, #004, #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer la commande racine `ade` avec l'aide complète et la structure pour les sous-commandes.

**Cas nominaux :**
- `ade --help` affiche la description de l'outil et la liste des commandes disponibles
- `ade --version` affiche la version (hardcodée `0.1.0` pour l'instant)
- `ade version` est un alias de `--version`
- L'aide inclut : usage, commandes disponibles, flags globaux
- Les commandes disponibles listées : `init`, `version`
- Le binaire a un nom convivial : `ade` (ou `ade.exe` sur Windows)

**Cas limites :**
- `ade` sans argument affiche l'aide (comportement Cobra par défaut)
- `ade commande-inconnue` affiche une erreur "unknown command" et l'aide

**Gestion d'erreurs :**
- Commande inconnue → message d'erreur clair + suggestion avec `--help`

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/root.go` | Créer | Commande racine Cobra |
| `internal/command/version.go` | Créer | Sous-commande `version` |
| `internal/command/root_test.go` | Modifier | Tests de la racine |
| `cmd/ade/main.go` | Modifier | Appel à `cmd.Execute()` |

### Signatures

```go
// internal/command/root.go
package command

var rootCmd = &cobra.Command{
    Use:   "ade",
    Short: "Automated Dev Environment - CLI tool",
    Long:  `ade est un outil en ligne de commande pour initialiser...`,
}

func Execute() error

// internal/command/version.go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Affiche la version",
}
```

### Contraintes techniques
- **Cobra** : Utiliser `cobra.Command` pour la définition des commandes
- **Version** : Constante `Version = "0.1.0"` dans un fichier séparé `internal/command/version.go`
- **`init()`** : Enregistrer les sous-commandes dans `init()` de `root.go` (pattern standard Cobra)
- **Package** : Tout est dans `internal/command/`

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/command/root_test.go`
- Scénario 1 : `Execute()` ne retourne pas d'erreur
- Scénario 2 : `ade --help` contient "Automated Dev Environment"
- Scénario 3 : `ade --version` contient "0.1.0"
- Scénario 4 : `ade inconnu` retourne une erreur contenant "unknown command"

#### Tests d'intégration
- **Fichier** : `internal/command/root_test.go`
- Scénario : Compiler et exécuter `ade.exe --help`, vérifier la sortie

### Documentation
- Voir Tâche #005 pour la documentation globale

### Exemples d'API
```bash
ade --help
ade --version
ade version
ade init --help
```
