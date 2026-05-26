# Conteneur de configuration (ade-config)

## Rôle

Le conteneur `ade-config` est le point central d'orchestration de
l'environnement de développement. Il expose :

- Une **API REST** (port 8080) pour l'interface web et le CLI
- Une **API gRPC** (port 9090) pour les plugins
- Une **interface web** (React) pour la configuration visuelle
- Un **healthcheck** Docker via `GET /health`
- La **découverte et surveillance** des plugins Docker

## Déploiement

Le conteneur est automatiquement inclus dans le `docker-compose.yml` généré
par `ade init ci`. Il est défini comme suit :

### Démarrage automatique

Le CLI (`ade`) détecte et démarre automatiquement l'orchestrateur via
`EnsureOrchestratorRunning()` :

1. Vérifie que Docker/Podman est disponible
2. Si le conteneur `ade-config` est déjà en cours d'exécution → OK
3. Si l'image `ade/ade-config:latest` est disponible localement → démarre
   automatiquement le conteneur avec les ports 8080 (REST) et 9090 (gRPC)
4. Si l'image n'est pas disponible → message invitant à exécuter `ade init ci`
   ou `scripts/build-orchestrator.ps1`

Ce comportement est activé par les commandes qui interagissent avec
l'orchestrateur (`ade plugin list`, `ade version -o`). L'image utilisée
peut être surchargée via la variable d'environnement `ADE_CONFIG_IMAGE`.

### Template docker-compose

```yaml
services:
  ade-config:
    image: ade/ade-config:latest
    container_name: ade-config
    ports:
      - "${ADE_CONFIG_PORT:-8080}:8080"
      - "9090:9090"
    env_file:
      - .env
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

## Build de l'image

```powershell
# Builder l'image localement
.\scripts\build-orchestrator.ps1
```

L'image est taguée `ade/ade-config:latest` et publiée sur le registre d'images.
La liste des images disponibles est gérée par le service registry séparé.

## Personnalisation

- **Port REST** : Modifier `ADE_CONFIG_PORT` dans `.env` ou utiliser le flag `--port`
- **Image** : Modifier `ADE_CONFIG_IMAGE` dans `.env` pour utiliser une image différente
- **Réseau** : Modifier `ADE_COMPOSE_NETWORK` dans `.env`

## Variables d'environnement

| Variable | Défaut | Description |
|----------|--------|-------------|
| `ADE_CONFIG_REST_PORT` | `8080` | Port HTTP REST |
| `ADE_CONFIG_GRPC_PORT` | `9090` | Port gRPC |
| `ADE_CONFIG_DATA_DIR` | `/data` | Répertoire de données |
| `ADE_CONFIG_DISCOVERY_INTERVAL` | `30s` | Intervalle de découverte des plugins |
| `ADE_CONFIG_HEALTH_INTERVAL` | `15s` | Intervalle de healthcheck des plugins |
| `ADE_CONFIG_MAX_HEALTH_FAILS` | `3` | Échecs max avant désenregistrement |
| `ADE_LOG_LEVEL` | `info` | Niveau de log |
