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

## Commandes

| Commande | Description |
|----------|-------------|
| `ade init specs` | Génère les fichiers de spécification |
| `ade init ci` | Initialise l'intégration continue |
| `ade version` | Affiche la version |

Voir [docs/cli/commands.md](docs/cli/commands.md) pour la documentation détaillée.

## Développement

```powershell
.\scripts\test.ps1    # Tests unitaires et intégration
go test -tags=e2e ./test/e2e/...  # Tests E2E
```
