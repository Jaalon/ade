# Environnement de préproduction locale

## Vue d'ensemble

L'environnement de préproduction locale permet de tester l'application dans
des conditions proches de la production, directement sur la machine du développeur.

## Architecture

L'environnement est déployé via Docker Compose et inclut :

- **ade-config** : Conteneur de configuration avec interface web (voir [config-container.md](config-container.md))
- **Réseau dédié** : Les conteneurs communiquent via un réseau Docker isolé

## Workflow

### 1. Initialisation

```bash
ade init ci
```

Cette commande :
1. Détecte Docker ou Podman
2. Vérifie que le démon est accessible
3. Génère `docker-compose.yml` et `.env` dans le répertoire courant
4. Déploie les conteneurs avec `docker compose up -d`
5. Affiche le statut des conteneurs

### 2. Gestion du cycle de vie

```bash
# Voir le statut
docker compose ps

# Voir les logs
docker compose logs -f

# Arrêter les conteneurs
docker compose down

# Redémarrer
docker compose restart
```

### 3. Configuration

Les variables d'environnement sont définies dans `.env` (généré par `ade init ci`) :

| Variable | Défaut | Description |
|----------|--------|-------------|
| `ADE_PROJECT_NAME` | (nom du répertoire) | Nom du projet |
| `ADE_CONFIG_PORT` | `8080` | Port du conteneur de configuration (REST) |
| `ADE_CONFIG_IMAGE` | `ade/ade-config:latest` | Image de l'orchestrateur |
| `ADE_COMPOSE_NETWORK` | `ade-network` | Nom du réseau Docker |
| `ADE_LOG_LEVEL` | `info` | Niveau de log (debug, info, warn, error) |

### Auto-start

Le CLI démarre automatiquement l'orchestrateur si l'image Docker est
disponible localement. Utilisez `ade version -o` pour vérifier le statut :
