# Tâche #005 - Story #001 : Build, tests et documentation

## Objectif
Mettre en place les scripts de build Windows, la suite de tests complète (unitaire, intégration, E2E) et la documentation du CLI.

## Contexte
- Story #001 : `docs/stories/story-001.md`
- Dépend de : Tâches #001, #002, #003, #004
- Nécessaire pour : Rien (dernière tâche de la story)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer les scripts de build, la configuration des tests et la documentation.

**Cas nominaux :**
- Un script PowerShell `scripts/build.ps1` compile le binaire Windows
- Un script PowerShell `scripts/test.ps1` exécute tous les tests
- Les tests unitaires passent : `go test ./internal/...` (ou avec `-v`)
- Le fichier `docs/cli/commands.md` documente toutes les commandes
- Le fichier `README.md` est mis à jour avec les instructions de base

**Cas limites :**
- Le script de build doit supporter la cross-compilation pour Windows (même si on est déjà sur Windows, préparer `GOOS=windows GOARCH=amd64`)
- Le binaire de sortie doit s'appeler `ade.exe`

**Gestion d'erreurs :**
- Si `go test` échoue, le script retourne un code d'erreur non-zero
- Si `go build` échoue, message d'erreur explicite

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `scripts/build.ps1` | Créer | Script de build Windows |
| `scripts/test.ps1` | Créer | Script d'exécution des tests |
| `test/e2e/e2e_test.go` | Créer | Test E2E (build + execution) avec build tag `e2e` |
| `docs/cli/commands.md` | Créer | Documentation des commandes |
| `README.md` | Créer | Documentation du projet |

### Signatures

```powershell
# scripts/build.ps1
# Build le projet et produit ade.exe
# Usage: .\scripts\build.ps1

# scripts/test.ps1
# Exécute les tests unitaires, d'intégration et E2E
# Usage: .\scripts\test.ps1
```

### Contraintes techniques
- **PowerShell** : Les scripts sont en PowerShell (.ps1), pas en Bash
- **Build** : `go build -o ade.exe ./cmd/ade` avec les flags `-ldflags="-s -w"` pour réduire la taille
- **Test E2E** : Doit être dans le répertoire `test/e2e/` avec le build tag `e2e` (`//go:build e2e`) pour ne pas s'exécuter avec les tests unitaires par défaut
- **README.md** : Format markdown standard, sections : Présentation, Prérequis, Installation/ Build, Utilisation, Commandes, Développement

### Tests à implémenter

#### Tests d'intégration
- **Fichier** : `internal/command/root_test.go` (ajouter aux tests existants)
- Scénario 1 : Exécuter la fonction `Execute()` et vérifier qu'elle ne bloque pas indéfiniment (timeout context)
- Scénario 2 : `ade init` affiche "specs" et "ci" dans l'aide

#### Tests E2E
- **Fichier** : `test/e2e/e2e_test.go` (package `e2e`, build tag `//go:build e2e`)
- Scénario 1 : Builder le projet, exécuter `ade.exe --help`, vérifier la sortie
  - Données : `go build -o ade_test.exe ./cmd/ade` puis `.\ade_test.exe --help`
  - Résultat attendu : stdout contient "Automated Dev Environment"

### Documentation

#### Documentation à créer
- `docs/cli/commands.md` :
  ```markdown
  # Commandes CLI

  ## ade
  Commande racine de l'outil Automated Dev Environment.

  ### Usage
  ade [command] [flags]

  ### Commandes disponibles
  - `init` — Initialise les composants du projet
    - `specs` — Génère les fichiers de spécification
    - `ci` — Initialise l'intégration continue
  - `version` — Affiche la version

  ### Flags globaux
  - `--help` — Affiche l'aide
  - `--version` — Affiche la version
  ```

- `README.md` :
  ```markdown
  # Automated Dev Environment (ade)

  Outil CLI pour initialiser un environnement de développement agentic robuste.

  ## Prérequis
  - Go 1.26+
  - Docker ou Podman (pour les fonctionnalités conteneurisées)

  ## Installation
  ```powershell
  git clone <repo>
  cd automated_dev_environment
  .\scripts\build.ps1
  ```

  ## Utilisation
  ```powershell
  .\ade.exe --help
  .\ade.exe init specs
  .\ade.exe init ci
  ```

  ## Développement
  ```powershell
  .\scripts\test.ps1    # Tests unitaires et intégration
  go test -tags=e2e ./test/e2e/...  # Tests E2E
  ```
  ```

### Exemples d'utilisation
```powershell
# Build
.\scripts\build.ps1

# Tests
.\scripts\test.ps1

# Test E2E uniquement
go test -tags=e2e -v ./test/e2e/
```
