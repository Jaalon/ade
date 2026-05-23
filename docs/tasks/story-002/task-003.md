# Tâche #003 - Story #002 : Commande `ade init specs`

## Objectif
Modifier la sous-commande `ade init specs` (placeholder créé dans Story #001) pour intégrer le service de génération de fichiers (Tâche #002) et le système de templates (Tâche #001), avec les flags `--force` et `--output`.

## Contexte
- Story #002 : `docs/stories/story-002.md`
- Dépend de : Tâche #001, Tâche #002
- Nécessaire pour : Tâche #004, Tâche #005
- Story #001 Tâche #003 : a créé `internal/command/init_specs.go` en placeholder

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Remplacer le placeholder de `ade init specs` par une implémentation complète qui :
1. Récupère la configuration utilisateur (nom du projet, etc.) depuis le fichier YAML de config s'il existe (Story #002 non dépendante de viper pour l'instant — utiliser des valeurs par défaut + flags)
2. Appelle `generator.Generate()` avec les options appropriées
3. Affiche un rapport clair des fichiers générés, ignorés ou en erreur
4. Supporte les flags `--force` et `--output`

**Flags :**
- `--force, -f` : Écrase les fichiers existants sans confirmation (bool, défaut: false)
- `--output, -o` : Répertoire de destination (string, défaut: "." répertoire courant)
- `--name` : Nom du projet (string, défaut: nom du répertoire courant)
- `--lang` : Langue pour les skills (string, défaut: "fr")
- `--module` : Module Go (string, défaut: nom du répertoire courant)

**Cas nominaux :**
- `ade init specs` génère tous les fichiers de configuration dans le répertoire courant
- `ade init specs --force` écrase les fichiers existants sans demander confirmation
- `ade init specs --output C:\mon-projet` génère dans un répertoire spécifique
- `ade init specs --lang en` génère les templates avec la langue anglaise
- À la fin, un récapitulatif est affiché :
  ```
  ✓ Fichiers créés : 12
  ∼ Fichiers ignorés (existant) : 2
  ✗ Erreurs : 0
  ```

**Cas limites :**
- `ade init specs` dans un répertoire déjà initialisé → confirmation pour chaque fichier (sauf `--force`)
- `ade init specs --output C:\inexistant` → le répertoire est créé automatiquement
- `ade init specs --output C:\fichier` (où le chemin pointe vers un fichier) → erreur explicite
- Aucun template disponible → message "Aucun template à générer"

**Gestion d'erreurs :**
- Erreur de génération d'un fichier → affichée dans le récapitulatif, la commande continue
- Erreur fatale (ex: output est un fichier) → erreur retournée par `RunE`
- Tous les templates invalides → rapport avec erreurs, message invitant à mettre à jour les templates

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/init_specs.go` | Modifier | Implémentation réelle de `ade init specs` |
| `internal/command/init_test.go` | Modifier | Ajouter les tests pour `init specs` |

### Signatures

```go
// internal/command/init_specs.go
package command

// init enregistre initSpecsCmd auprès de initCmd

var initSpecsCmd = &cobra.Command{
    Use:   "specs",
    Short: "Génère les fichiers de configuration du projet",
    Long:  `Génère les fichiers de configuration locaux (.gitignore, config IDE,
           skills, serveurs MCP, workflow de développement) depuis des templates intégrés.`,
    RunE: func(cmd *cobra.Command, args []string) error,
}

// Flags
var (
    initSpecsForce  bool
    initSpecsOutput string
    initSpecsName   string
    initSpecsLang   string
    initSpecsModule string
)

func init() {
    // Enregistrement de la commande et des flags
    initSpecsCmd.Flags().BoolVarP(&initSpecsForce, "force", "f", false,
        "Écraser les fichiers existants sans confirmation")
    initSpecsCmd.Flags().StringVarP(&initSpecsOutput, "output", "o", ".",
        "Répertoire de destination")
    initSpecsCmd.Flags().StringVar(&initSpecsName, "name", "",
        "Nom du projet (défaut: nom du répertoire)")
    initSpecsCmd.Flags().StringVar(&initSpecsLang, "lang", "fr",
        "Langue pour les skills (fr ou en)")
    initSpecsCmd.Flags().StringVar(&initSpecsModule, "module", "",
        "Module Go (défaut: nom du répertoire)")
}
```

### Contraintes techniques
- **Cobra `RunE`** : Déjà en place (Story #001). Utiliser `RunE` pour retourner les erreurs.
- **Flags** : Enregistrer tous les flags dans `init()` du fichier. Utiliser `BoolVarP` et `StringVarP` pour les flags avec shorthand (`-f`, `-o`).
- **Défauts intelligents** :
  - `--name` : Si non fourni, utiliser `filepath.Base(outputDir)`
  - `--module` : Si non fourni, utiliser la valeur de `--name`
  - `--lang` : "fr" par défaut
- **Affichage du rapport** : Utiliser `fmt.Fprintf(os.Stdout, ...)` pour le rapport. Formater avec des émojis unicode (✓, ∼, ✗) ou des symboles ASCII.
- **Sortie** : Éviter les logs superflus sur stdout (seulement le rapport). Les warnings sur stderr.
- **Respect du pattern `init()`** : Suivre le même pattern que les autres commandes Cobra du projet.

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/command/init_test.go` (ajouter aux tests existants)
- Scénario 1 : `ade init specs --help` affiche les flags disponibles
  - Résultat attendu : la sortie contient "--force", "--output", "--name", "--lang"
- Scénario 2 : `ade init specs` avec `--output` sur un répertoire temporaire vide retourne nil
  - Données : `os.MkdirTemp`, exécuter la commande avec `--output {tmpDir} --force`
  - Résultat attendu : `err == nil`, les fichiers sont créés dans tmpDir
- Scénario 3 : `ade init specs` avec `--output` sur un fichier existant retourne une erreur
  - Données : créer un fichier régulier, l'utiliser comme --output
  - Résultat attendu : erreur retournée (output n'est pas un répertoire)
- Scénario 4 : Vérifier que la commande affiche un rapport structuré
  - Capturer stdout, vérifier la présence de "créés" et du nombre de fichiers

### Documentation
- Les flags sont auto-documentés via Cobra (`--help`)
- Voir Tâche #005 pour la documentation complète de la commande

### Exemples d'utilisation
```bash
# Génération standard dans le répertoire courant
ade init specs

# Génération avec écrasement automatique
ade init specs --force

# Génération dans un répertoire spécifique
ade init specs --output C:\Projects\mon-app

# Génération avec paramètres personnalisés
ade init specs --name "mon-app" --lang en --module "github.com/user/mon-app"
```
