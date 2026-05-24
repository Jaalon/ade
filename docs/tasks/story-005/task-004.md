# Tâche #004 - Story #005 : Documentation du pipeline CI local

## Objectif
Créer la documentation complète pour l'architecture du pipeline CI, les prérequis métier pour le choix futur du moteur CI, et le pilotage du pipeline via plugins Docker.

## Contexte
- Story #005 : `docs/stories/story-005.md`
- Dépend de : Tâche #001, Tâche #002, Tâche #003 (toutes les implémentations doivent être terminées)
- Nécessaire pour : Aucune (documentation finale)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer trois fichiers de documentation dans `docs/ci/` :

1. **`docs/ci/architecture.md`** : Architecture et interface du pipeline CI
2. **`docs/ci/requirements.md`** : Prérequis métier pour le choix futur du moteur CI
3. **`docs/ci/plugins.md`** : Pilotage du pipeline via plugins Docker (vision future)

**Cas nominaux :**
- Les trois fichiers sont en Markdown, en français, bien structurés
- `architecture.md` documente l'interface Pipeline, les types, l'ordre des stages, les exécuteurs disponibles, les commandes CLI
- `requirements.md` liste les prérequis identifiés : build Java, publication rapports de tests, promotion de builds, déploiement de conteneurs
- `plugins.md` décrit comment configurer et étendre le pipeline via des steps Docker (vision future)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/ci/architecture.md` | Créer | Documentation de l'architecture du pipeline CI |
| `docs/ci/requirements.md` | Créer | Prérequis métier pour le choix du moteur CI |
| `docs/ci/plugins.md` | Créer | Pilotage du pipeline via plugins Docker (vision future) |

### Contenu attendu

#### `docs/ci/architecture.md`

```markdown
# Architecture du pipeline CI local

## Vue d'ensemble
Le pipeline CI local est une architecture abstraite, agnostique du moteur
d'exécution. Il définit une interface `Pipeline` qui peut être implémentée
par différents exécuteurs (dry-run, local, ou futur moteur CI).

## Interface Pipeline

```go
type Pipeline interface {
    Run(ctx context.Context, config PipelineConfig) (*PipelineResult, error)
    Validate(config PipelineConfig) error
}
```

Toute implémentation doit respecter l'ordre des stages et retourner
des résultats standardisés.

## Stages du pipeline

Le pipeline se compose de 6 stages exécutés séquentiellement :

| Ordre | Stage | Description |
|-------|-------|-------------|
| 1 | `build` | Construction du projet (compilation Go, Java, etc.) |
| 2 | `unit-test` | Exécution des tests unitaires |
| 3 | `integration-test` | Exécution des tests d'intégration |
| 4 | `test-deploy` | Déploiement dans un environnement de test |
| 5 | `e2e` | Tests end-to-end |
| 6 | `preprod` | Déploiement en préproduction |

Si un stage échoue, les stages suivants sont ignorés (`skipped`).

## Exécuteurs disponibles

| Exécuteur | Description | Usage |
|-----------|-------------|-------|
| `DryRunExecutor` | Simule l'exécution sans effet réel | Défaut, tests, validation |
| `LocalExecutor` | Exécute les commandes sur la machine hôte | `ade pipeline run --local` |

> L'exécuteur Docker (`DockerStepExecutor`) sera disponible dans une version
> ultérieure (Story #007/008).

## Configuration

Le pipeline est configuré via un fichier YAML (`ade-pipeline.yaml`)

```yaml
stages:
  - type: build
    name: "Compilation"
    enabled: true
    steps:
      - name: "Compiler le projet"
        command: ["go", "build", "./..."]
```

## Commandes CLI

```bash
# Initialiser la configuration
ade pipeline init
ade pipeline init --template go
ade pipeline init --template java

# Exécuter le pipeline
ade pipeline run
ade pipeline run --local
ade pipeline run --verbose
ade pipeline run --dry-run
ade pipeline run --config ./mon-pipeline.yaml
```

## Extensibilité

Pour ajouter un nouvel exécuteur, implémenter l'interface `Executor` :

```go
type Executor interface {
    Execute(ctx context.Context, step StepConfig) (*StepResult, error)
}
```

Voir `docs/ci/plugins.md` pour l'extension via plugins Docker (vision future).
```

#### `docs/ci/requirements.md`

```markdown
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
```

#### `docs/ci/plugins.md`

```markdown
# Pilotage du pipeline via plugins Docker

> ⚠ **Note** : Cette fonctionnalité est prévue pour une version ultérieure
> (Story #007 — Architecture de plugins Docker, Story #008 — Orchestrateur).
> L'interface Executor est déjà définie (Task #001) mais `DockerStepExecutor`
> n'est pas encore implémenté.

## Principe (vision future)
Les plugins Docker permettront d'exécuter des étapes du pipeline dans des
conteneurs isolés, chacun apportant des outils et dépendances spécifiques.

## Configuration d'une étape Docker (vision future)

```yaml
stages:
  - type: integration-test
    name: "Tests d'intégration"
    enabled: true
    steps:
      - name: "Tests avec PostgreSQL"
        image: "golang:1.26-alpine"
        command: ["go", "test", "-tags=integration", "./..."]
        env:
          DB_HOST: "postgres"
          DB_PORT: "5432"
```

## Spécification des steps Docker

| Champ | Requis | Description |
|-------|--------|-------------|
| `name` | Oui | Nom de l'étape (affiché dans les logs) |
| `image` | Oui* | Image Docker à utiliser |
| `command` | Non | Commande à exécuter dans le conteneur |
| `env` | Non | Variables d'environnement |
| `workdir` | Non | Répertoire de travail dans le conteneur |

> *Soit `image` soit `command` doit être renseigné (ou les deux).

## Exemples de plugins Docker (vision future)

### Analyse de code (SonarQube)
```yaml
steps:
  - name: "Analyse SonarQube"
    image: "sonarsource/sonar-scanner-cli:latest"
    command: ["sonar-scanner"]
    env:
      SONAR_HOST_URL: "http://sonarqube:9000"
      SONAR_TOKEN: "${SONAR_TOKEN}"
```

### Build multi-plateforme
```yaml
steps:
  - name: "Build binaries"
    image: "golang:1.26-alpine"
    command:
      - "sh"
      - "-c"
      - "GOOS=linux GOARCH=amd64 go build -o ./bin/app-linux . && GOOS=windows GOARCH=amd64 go build -o ./bin/app.exe ."
```

### Tests de charge (k6)
```yaml
steps:
  - name: "Tests de charge"
    image: "grafana/k6:latest"
    command: ["run", "/tests/load-test.js"]
```

## Bonnes pratiques (vision future)

1. **Images légères** : Préférer `alpine` ou `slim` pour minimiser le temps de pull
2. **Variables sensibles** : Utiliser `env` avec des références à `.env` (ne pas hardcoder)
3. **Volumes** : Les répertoires du projet sont montés automatiquement dans le conteneur
4. **Réseau** : Les conteneurs peuvent communiquer via le réseau Docker du projet

## Extension avec l'orchestrateur (vision future)

Quand l'orchestrateur sera déployé (Story #008), les plugins Docker
pourront être :
- **Découverts automatiquement** : L'orchestrateur détecte les plugins disponibles
- **Configurés via l'interface web** : Activation, paramètres, version
- **Versionnés** : Chaque plugin a une version, rollback possible
```

### Contraintes techniques
- **Format** : Markdown, encodé en UTF-8
- **Langue** : Français
- **Style** : Clair, concis, avec des tableaux pour les références rapides
- **Liens** : Utiliser des liens relatifs entre fichiers de documentation
- **Exemples** : Les exemples YAML et Go doivent être valides et correspondre à l'implémentation réelle
- **Cohérence** : Vérifier que les exemples de commandes utilisent `ade pipeline` (pas `ade ci`)
- **Mentions "vision future"** : Bien indiquer ce qui est disponible maintenant vs. ce qui viendra plus tard
- **Aucun test automatisé** pour la documentation. Vérification manuelle recommandée.

### Documentation
Cette tâche **est** la documentation. Aucune documentation supplémentaire.

### Exemples d'utilisation
```bash
# Afficher la documentation du pipeline
ade pipeline --help

# Lire la documentation
type docs\ci\architecture.md

# Consulter les prérequis métier
type docs\ci\requirements.md
```
