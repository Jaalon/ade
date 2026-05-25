# Exemples de plugins

## Plugin Templates

Le plugin **templates** fournit des templates de projets via une API REST. Il utilise le SDK de plugins pour gérer l'enregistrement, le health check et les capacités.

### Source

| Fichier | Description |
|---------|-------------|
| `plugins/templates/main.go` | Point d'entrée |
| `plugins/templates/server.go` | Handlers REST |
| `plugins/templates/templates.go` | Liste des templates et rendu |
| `plugins/templates/embed/` | Templates embarqués (`.tmpl`) |
| `plugins/templates/Dockerfile` | Image Docker multi-stage |

### Endpoints REST

| Méthode | Path | Description |
|---------|------|-------------|
| `GET` | `/api/v1/templates` | Liste les templates disponibles |
| `POST` | `/api/v1/templates/render` | Rend un template avec des variables |

### Exemples d'utilisation

```bash
# Lister les templates
curl http://localhost:8081/api/v1/templates

# Rendre un template
curl -X POST http://localhost:8081/api/v1/templates/render \
  -H "Content-Type: application/json" \
  -d '{"template_name": "gitignore", "variables": {"ProjectName": "mon-projet"}}'

# Mode strict (erreur si template inconnu)
curl -X POST http://localhost:8081/api/v1/templates/render \
  -H "Content-Type: application/json" \
  -d '{"template_name": "inconnu", "variables": {}, "strict": true}'
```

### Réponse de liste

```json
{
  "templates": [
    {"name": "gitignore", "description": "Fichier .gitignore pour un projet Go", "language": "go"},
    {"name": "docker-compose", "description": "Environnement de préproduction", "language": "yaml"}
  ]
}
```

### Réponse de rendu

```json
{
  "files": {
    "gitignore": "*.exe\n*.test\n*.out\n..."
  }
}
```

### Build Docker

```bash
docker build -t ade-plugin-templates:latest plugins/templates
```

### Déploiement avec docker-compose

```yaml
services:
  templates-plugin:
    image: ade-plugin-templates:latest
    environment:
      PLUGIN_NAME: templates
      PLUGIN_VERSION: "1.0.0"
      PLUGIN_DESCRIPTION: "Fournit des templates de projets"
      PLUGIN_ORCHESTRATOR_URL: "http://registry:8082"
      PLUGIN_HTTP_PORT: "8081"
      PLUGIN_GRPC_PORT: "50051"
    networks:
      - ade-network
    labels:
      ade.plugin.name: "templates"
      ade.plugin.version: "1.0.0"
    restart: unless-stopped
```

## Idées de plugins

- **GitHub** : Importer des issues, gérer les PRs via l'API GitHub
- **Slack** : Notifications de build, alertes de validation
- **Database** : Migrations, seeding de base de données
- **CI** : Adapter Jenkins, GitHub Actions, GitLab CI

## Patterns avancés

### Configuration via l'interface web

Les plugins peuvent exposer un endpoint `GET /api/v1/config` retournant un JSON Schema que l'orchestrateur utilise pour générer un formulaire de configuration dans l'interface web.

### Agrégation de données

Un plugin peut appeler d'autres plugins via le réseau Docker interne pour agréger des données. Par exemple, un plugin "dashboard" peut agréger les données des plugins GitHub et Slack.
