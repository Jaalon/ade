# Prérequis métier — Pipeline CI local

## Objectif
Ce document identifie les prérequis métier pour le choix futur d'un moteur
CI réel (Jenkins, GitLab CI, GitHub Actions local, etc.). Ces critères
permettront de sélectionner le moteur le plus adapté.

## Prérequis identifiés

### 1. Support de builds Java
- **Nécessité** : Les projets Quarkus/Spring nécessitent Maven ou Gradle
- **Critères** :
  - Exécution de `mvn clean verify` ou `gradle build`
  - Cache des dépendances (`.m2/repository`)
  - Support multi-modules Maven/Gradle
  - Gestion des versions JDK (8, 11, 17, 21+)

### 2. Publication de rapports de tests
- **Nécessité** : Visualisation et historique des résultats de tests
- **Critères** :
  - Support du format JUnit XML (standard)
  - Publication de rapports de couverture (JaCoCo, Go cover)
  - Agrégation multi-module
  - Historique des tendances (succès/échec dans le temps)

### 3. Promotion de builds
- **Nécessité** : Promotion manuelle ou automatique d'un build vers l'étape suivante
- **Critères** :
  - Promotion manuelle : un opérateur valide le build avant déploiement
  - Promotion automatique : basée sur des critères (tests passés, couverture ≥ 80%)
  - Rollback : possibilité de revenir à un build précédent validé
  - Traçabilité : qui a promu quoi et quand

### 4. Déploiement de conteneurs
- **Nécessité** : Déploiement automatisé via conteneurs Docker
- **Critères** :
  - Build et push d'images Docker
  - Déploiement via docker-compose ou Kubernetes
  - Gestion des tags d'images (latest, stable, version)
  - Tests de santé (healthcheck) après déploiement

## Notes pour le choix du moteur

| Critère | Poids | Remarque |
|---------|-------|----------|
| Open source | Élevé | Pas de licence coûteuse |
| Exécution locale | Essentiel | Pas de dépendance cloud obligatoire |
| Windows support | Élevé | Cible principale V1 |
| Pipeline-as-code | Moyen | Déclaratif YAML souhaitable |
| Plugins/écosystème | Moyen | Doit pouvoir s'intégrer avec Docker |

## Prochaines étapes
Une fois l'architecture validée et les plugins Docker opérationnels
(Story #007), un moteur CI réel pourra être intégré via un adaptateur
implémentant l'interface `Pipeline`.
