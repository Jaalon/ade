# ade init ci

## Description

Initialise et déploie l'environnement de préproduction locale via Docker Compose.

Cette commande détecte automatiquement Docker ou Podman, génère les fichiers
de configuration nécessaires et déploie les conteneurs.

## Utilisation

```
ade init ci [flags]
```

## Flags

| Flag | Court | Défaut | Description |
|------|-------|--------|-------------|
| `--output` | `-o` | `.` | Répertoire de sortie pour les fichiers générés |
| `--force` | `-f` | `false` | Écraser les fichiers existants sans confirmation |
| `--name` | | (nom du répertoire) | Nom du projet pour le déploiement |
| `--port` | | `8080` | Port du conteneur de configuration (web UI) |
| `--network` | | `ade-network` | Nom du réseau Docker |
| `--image` | | `ade/ade-config:latest` | Image de l'orchestrateur (défaut via `docker.DefaultConfigImage`) |

## Exemples

```bash
# Déploiement de base
ade init ci

# Avec options personnalisées
ade init ci --output ./preprod --port 9090 --name mon-app

# Avec une image personnalisée
ade init ci --image registry.example.com/ade-config:v2

# Forcer l'écrasement des fichiers existants
ade init ci --force
```

## Dépendances

- Docker Desktop (Windows) ou Podman
- Docker Compose (plugin inclus dans Docker Desktop)
- Pour Podman : `podman-compose`

## Dépannage

| Problème | Solution |
|----------|----------|
| "Docker ou Podman requis" | Installer Docker Desktop depuis https://www.docker.com/products/docker-desktop/ |
| "Démon inaccessible" | Démarrer Docker Desktop depuis le menu Démarrer |
| "Commande compose non trouvée" | Mettre à jour Docker Desktop ou installer podman-compose |
