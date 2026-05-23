# Tâche #001 - Story #001 : Structure du projet et dépendances

## Objectif
Initialiser la structure du projet Go avec les répertoires standards, les dépendances Cobra et testify, et un fichier `cmd/ade/main.go` minimal.

## Contexte
- Story #001 : `docs/stories/story-001.md`
- Dépend de : Aucune
- Nécessaire pour : Tâches #002 à #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Mettre en place la structure squelette du projet Go pour l'outil CLI `ade`.

**Cas nominaux :**
- Le module Go existe (`go.mod` avec `module automated_dev_environment`, déjà présent)
- Les répertoires suivants sont créés :
  - `cmd/ade/` — point d'entrée du binaire
  - `internal/command/` — logique des commandes Cobra
  - `internal/docker/` — interaction Docker/Podman
  - `internal/config/` — configuration du projet
- Les dépendances sont ajoutées au `go.mod` :
  - `github.com/spf13/cobra` (dernière version)
  - `github.com/docker/docker` (dernière version, SDK Docker)
  - `github.com/stretchr/testify` (dernière version, pour les tests)
- Note : `github.com/spf13/viper` sera ajouté dans la Story #002 (config YAML)
- `go mod tidy` peut prendre un certain temps à cause du SDK Docker (nombreuses dépendances transitives)
- Le fichier `cmd/ade/main.go` est créé avec un appel à `cmd.Execute()`
- La commande `go build ./cmd/ade` compile sans erreur

**Gestion d'erreurs :**
- `go mod tidy` doit réussir sans erreur après ajout des dépendances

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `go.mod` | Modifier | Ajouter les dépendances |
| `go.sum` | Créer | Généré par `go mod tidy` |
| `cmd/ade/main.go` | Créer | Point d'entrée du binaire |
| `.gitignore` | Créer | Ignorer binaires, `.idea/`, etc. |

### Signatures

```go
// cmd/ade/main.go
package main

func main()
```

### Contraintes techniques
- **Go 1.26** : Utiliser la version du module existant
- **Package main** : `cmd/ade/main.go` est le point d'entrée, package `main`
- **Structure** : Suivre le layout standard `cmd/` + `internal/` des projets Go
- **Pas de logique métier** dans le `main.go` — uniquement l'appel à `cmd.Execute()`

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/command/root_test.go`
- Scénario 1 : Vérifier que la commande racine existe
  - Données : Appel à `rootCmd`
  - Résultat attendu : `rootCmd` n'est pas nil

### Documentation
Aucune documentation spécifique pour cette tâche.

### Exemples d'utilisation
```bash
# Compilation
go build ./cmd/ade

# Vérification du binaire
./ade.exe --help
```
