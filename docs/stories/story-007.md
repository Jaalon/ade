# Story #007 : Architecture de plugins Docker (REST + gRPC)

## Description
En tant que développeur, je veux étendre les fonctionnalités du CLI et de l'orchestrateur via des conteneurs Docker "plugins" indépendants qui exposent une API (REST et gRPC), afin d'ajouter des connecteurs pour services tiers sans modifier les composants centraux.

## Critères d'acceptation
- [ ] Les plugins sont des conteneurs Docker indépendants, découverts par l'orchestrateur
- [ ] Les plugins exposent à la fois une API REST/HTTP et une API gRPC
- [ ] L'orchestrateur découvre les plugins via un mécanisme d'enregistrement
- [ ] Le CLI détecte et se connecte aux plugins via l'orchestrateur
- [ ] Un premier plugin d'exemple est fourni (ex: fourniture de templates)
- [ ] Les plugins sont configurables via l'interface web de l'orchestrateur
- [ ] La documentation explique comment créer un nouveau plugin (REST et gRPC)
- [ ] Le CLI et l'orchestrateur restent fonctionnels même si aucun plugin n'est disponible

## Tests automatisés
- Test unitaire : Interface et contrat API du plugin (REST + gRPC)
- Test d'intégration : Communication orchestrateur ↔ plugin via REST et gRPC dans un réseau Docker de test
- Test E2E : Installation, enregistrement et utilisation d'un plugin complet avec les deux protocoles

## Documentation
- `docs/plugins/architecture.md` — Architecture des plugins Docker (REST + gRPC)
- `docs/plugins/discovery.md` — Mécanisme de découverte et d'enregistrement
- `docs/plugins/development.md` — Guide de développement d'un plugin
- `docs/plugins/examples.md` — Exemples de plugins (templates, GitHub, etc.)

## Valeur utilisateur
Permet d'étendre les capacités de l'outil via un écosystème de plugins Docker indépendants, découverts dynamiquement par l'orchestrateur, avec le choix du protocole adapté à chaque cas d'usage.

## Dépendances
- Story #001
- Story #004
- Story #008
