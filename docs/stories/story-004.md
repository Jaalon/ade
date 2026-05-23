# Story #004 : Déploiement docker-compose de préproduction

## Description
En tant que développeur, je veux exécuter `ade init ci` afin de déployer un environnement de préproduction locale via docker-compose, incluant le conteneur de configuration et les services nécessaires à mon application.

## Critères d'acceptation
- [ ] `ade init ci` génère un `docker-compose.yml` adapté au projet
- [ ] La commande détecte si Docker ou Podman est disponible
- [ ] Le conteneur de configuration (avec web UI) est inclus dans le déploiement
- [ ] Les conteneurs sont créés à la volée selon les besoins définis dans la configuration
- [ ] Un fichier `.env` est généré avec les variables d'environnement par défaut
- [ ] La commande affiche l'état des conteneurs après déploiement

## Tests automatisés
- Test unitaire : Génération du `docker-compose.yml` depuis différentes configurations
- Test d'intégration : Lancement des conteneurs dans un environnement Docker de test
- Test E2E : Cycle complet (génération → déploiement → vérification → arrêt)

## Documentation
- `docs/commands/init-ci.md` — Documentation de la sous-commande
- `docs/deployment/preprod.md` — Guide de l'environnement de préproduction
- `docs/deployment/config-container.md` — Conteneur de configuration et web UI

## Valeur utilisateur
Automatise le déploiement de l'environnement de préproduction, avec le conteneur de configuration comme point central d'orchestration.

## Dépendances
- Story #001
- Story #002
