# Tâche #003 - Story #001 : Sous-commande `init`

## Objectif
Implémenter la sous-commande `ade init` avec les sous-commandes `specs` et `ci`, toutes deux en version squelette.

## Contexte
- Story #001 : `docs/stories/story-001.md`
- Story #002 : `docs/stories/story-002.md` (pour comprendre le besoin `init specs`)
- Story #004 : `docs/stories/story-004.md` (pour comprendre le besoin `init ci`)
- Dépend de : Tâche #002
- Nécessaire pour : Tâche #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer la commande `ade init` et ses sous-commandes `ade init specs` et `ade init ci`.

**Cas nominaux :**
- `ade init` affiche l'aide des sous-commandes `init`
- `ade init --help` liste `specs` et `ci` avec description
- `ade init specs` affiche un message : "Initialisation des spécifications... (à implémenter)"
- `ade init ci` affiche un message : "Initialisation de l'intégration continue... (à implémenter)"
- Les commandes retournent un code de sortie 0

**Cas limites :**
- `ade init inconnu` affiche une erreur "unknown command" pour `init`
- Le comportement actuel est un placeholder : les vraies implémentations viendront dans les Stories #002 et #004

**Gestion d'erreurs :**
- Toute sous-commande inconnue sous `init` → message d'erreur + aide

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/init.go` | Créer | Commande `init` Cobra |
| `internal/command/init_specs.go` | Créer | Sous-commande `init specs` |
| `internal/command/init_ci.go` | Créer | Sous-commande `init ci` |
| `internal/command/init_test.go` | Créer | Tests de `init` |

### Signatures

```go
// internal/command/init.go
package command

func init() // enregistre initCmd et ses sous-commandes

// internal/command/init_specs.go
var initSpecsCmd = &cobra.Command{
    Use:   "specs",
    Short: "Initialise les fichiers de spécification",
    RunE: func(cmd *cobra.Command, args []string) error,
}

// internal/command/init_ci.go
var initCiCmd = &cobra.Command{
    Use:   "ci",
    Short: "Initialise l'intégration continue",
    RunE: func(cmd *cobra.Command, args []string) error,
}
```

### Contraintes techniques
- **Cobra `RunE`** : Utiliser `RunE` plutôt que `Run` pour permettre le retour d'erreurs
- **Messages de placeholder** : Utiliser `fmt.Println` pour les messages, `os.Stdout` pas `os.Stderr` (ce n'est pas une erreur)
- **`init()`** : L'enregistrement se fait dans `init()` de chaque fichier (pattern standard Cobra)

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/command/init_test.go`
- Scénario 1 : `ade init --help` contient "specs" et "ci"
- Scénario 2 : `ade init specs` retourne nil (pas d'erreur)
- Scénario 3 : `ade init ci` retourne nil (pas d'erreur)
- Scénario 4 : `ade init inconnu` retourne une erreur

### Documentation
- Voir Tâche #005 pour la documentation globale

### Exemples d'API
```bash
ade init --help
ade init specs
ade init ci
```
