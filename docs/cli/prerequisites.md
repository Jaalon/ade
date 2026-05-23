# Prérequis système

## Docker ou Podman

L'outil `ade` nécessite un moteur de conteneurs compatible Docker pour les fonctionnalités avancées (déploiement de préproduction, plugins, etc.).

### Options supportées
- **Docker Desktop** (recommandé sur Windows)
- **Podman** (alternative open-source, via `podman machine`)

### Vérification
```powershell
docker --version
```

### Installation
- **Docker Desktop** : https://docs.docker.com/desktop/setup/install/windows-install/
- **Podman** : https://podman.io/docs/installation

> 💡 L'outil fonctionne sans Docker pour les opérations de base (génération de configs, etc.).
> Les fonctionnalités conteneurisées nécessitent un moteur en cours d'exécution.
