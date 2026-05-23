# Story #005 : Pipeline CI locale automatisée

## Description
En tant que développeur, je veux disposer d'une architecture de pipeline CI locale extensible et agnostique, pilotée par plugins, afin de build, tester et déployer mon application sans dépendance cloud.

## Critères d'acceptation
- [ ] Une interface abstraite de pipeline CI est définie, agnostique du moteur sous-jacent
- [ ] Les étapes attendues sont documentées : build → tests unitaires/intégration → déploiement test → e2e → préprod
- [ ] Le pipeline est configurable et pilotable via des plugins Docker
- [ ] Les prérequis métier sont identifiés : support build Java, publication rapports de tests, promotion de builds, déploiement de conteneurs
- [ ] Un adaptateur "dry-run" simule le pipeline sans moteur CI réel
- [ ] La configuration du pipeline se fait via le conteneur de configuration (orchestrateur)

## Tests automatisés
- Test unitaire : Interface du pipeline et simulation dry-run
- Test d'intégration : Boucle complète simulée avec l'adaptateur dry-run
- Test E2E : Non applicable (pas de moteur CI réel implémenté)

## Documentation
- `docs/ci/architecture.md` — Architecture et interface du pipeline CI
- `docs/ci/requirements.md` — Prérequis métier pour le choix futur du moteur
- `docs/ci/plugins.md` — Pilotage du pipeline via plugins Docker

## Valeur utilisateur
Prépare une architecture CI locale agnostique et pilotable par plugins, sans bloquer le développement sur le choix d'un moteur spécifique.

## Dépendances
- Story #001
- Story #004
- Story #008 (orchestrateur)
