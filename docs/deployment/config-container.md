# Conteneur de configuration (ade-config)

## Rôle

Le conteneur `ade-config` est le point central d'orchestration de
l'environnement de développement. Il expose :

- Une **API REST/gRPC** pour interagir avec les plugins et services
- Une **interface web** pour la configuration visuelle

## Déploiement

Le conteneur est automatiquement inclus dans le `docker-compose.yml` généré
par `ade init ci`. Il est défini comme suit :

```yaml
services:
  ade-config:
    image: nginx:alpine
    container_name: ade-config
    ports:
      - "${ADE_CONFIG_PORT:-8080}:80"
    env_file:
      - .env
    restart: unless-stopped
```

## Personnalisation

- **Port** : Modifier `ADE_CONFIG_PORT` dans `.env` ou utiliser le flag `--port`
- **Réseau** : Modifier `ADE_COMPOSE_NETWORK` dans `.env`

## Notes

- **V1** : Le conteneur utilise `nginx:alpine` comme image placeholder. Il est fonctionnel (sert une page nginx par défaut) mais sera remplacé par l'image réelle dans la Story #008.
- L'implémentation complète du conteneur de configuration avec API et web UI est prévue dans la Story #008.
- Le service est défini dans docker-compose.yml et sera déployé automatiquement par `ade init ci`.
