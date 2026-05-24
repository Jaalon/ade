# Pilotage du pipeline via plugins Docker

> **Note** : Cette fonctionnalité est prévue pour une version ultérieure
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
