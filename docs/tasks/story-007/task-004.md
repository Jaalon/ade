# Tâche #004 - Story #007 : Plugin d'exemple « templates »

## Objectif
Créer un premier plugin Docker fonctionnel qui fournit des templates de projets via REST et gRPC, en utilisant le SDK développé dans la Tâche #002, avec son Dockerfile et sa configuration d'intégration dans docker-compose.

## Contexte
- Story #007 : `docs/stories/story-007.md`
- Dépend de : Tâche #002 (SDK plugins), Tâche #003 (registry)
- Nécessaire pour : Tâche #005 (CLI)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer un plugin « templates » complet qui expose les templates de projets (actuellement embarqués dans le binaire Go) via une API REST et gRPC. Le plugin est un binaire Go standalone, empaqueté dans une image Docker légère.

**Cas nominaux :**
- Le plugin expose un endpoint REST `GET /api/v1/templates` qui liste les templates disponibles (nom, description, langage)
- Le plugin expose un endpoint REST `POST /api/v1/templates/render` qui rend un template avec des variables
- Le plugin expose les mêmes fonctionnalités via gRPC : `ListTemplates(Empty)` et `RenderTemplates(RenderRequest)`
- Le plugin embarque les mêmes templates que `internal/templates/embed/` dans son binaire via `//go:embed`
- Le plugin s'enregistre automatiquement auprès de l'orchestrateur via le SDK
- Un Dockerfile produit une image légère (scratch ou distroless)
- Une entrée dans `docker-compose.yml` permet de déployer le plugin avec l'orchestrateur

**Cas limites :**
- Template inexistant → 404 avec message "template not found"
- Variables manquantes pour le rendu → le template rendu conserve les `{{.Var}}` non résolues (comportement Go template par défaut) ou retourne une erreur selon le paramètre `strict`
- Aucun template disponible → liste vide (200 OK)
- Le plugin doit être configurable via les variables d'environnement du SDK (`PLUGIN_*`)

**Gestion d'erreurs :**
- Template inconnu → 404 + message explicite
- Requête `render` avec body invalide → 400 Bad Request
- Erreur de rendu Go template → 422 Unprocessable Entity avec détail

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `plugins/templates/main.go` | Créer | Point d'entrée du plugin templates |
| `plugins/templates/server.go` | Créer | Handlers REST + service gRPC |
| `plugins/templates/templates.go` | Créer | Liste des templates et rendu |
| `plugins/templates/embed/` | Créer | Templates embarqués (copie de `internal/templates/embed/`) |
| `plugins/templates/Dockerfile` | Créer | Dockerfile pour l'image du plugin |
| `plugins/templates/.dockerignore` | Créer | Ignorer fichiers inutiles dans le build |
| `plugins/templates/server_test.go` | Créer | Tests du plugin |
| `plugins/templates/go.mod` | Créer | Module Go du plugin |
| `internal/plugins/contract/` | Dépend de | Types du contrat |

### Signatures

```go
// plugins/templates/server.go
package main

// TemplateService implémente le service de templates (REST + gRPC).
type TemplatePlugin struct {
    templates map[string]*Template
    sdk       *sdk.PluginServer
}

// Template décrit un template disponible.
type Template struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Language    string `json:"language"`
    Path        string `json:"path"`
}

// RenderRequest est la requête de rendu d'un template.
type RenderRequest struct {
    TemplateName string            `json:"template_name"`
    Variables    map[string]string `json:"variables"`
    Strict       bool              `json:"strict"`
}

// RenderResponse est le résultat du rendu.
type RenderResponse struct {
    Files map[string]string `json:"files"`
}

func NewTemplatePlugin() *TemplatePlugin
func (p *TemplatePlugin) ListTemplates(w http.ResponseWriter, r *http.Request)
func (p *TemplatePlugin) RenderTemplate(w http.ResponseWriter, r *http.Request)
```

### Contraintes techniques
- **Module unique** : Le plugin est dans `plugins/templates/` avec son propre `go.mod`, il importe le SDK via `replace` directive pointant vers `../../internal/plugins/sdk` (en attendant la publication du module SDK séparé)
- **Image légère** : Multi-stage Dockerfile : build avec `golang:1.26-alpine`, run avec `scratch` ou `gcr.io/distroless/base`
- **Templates** : Copier les fichiers `.tmpl` depuis `internal/templates/embed/` dans le plugin
- **Ports** : HTTP sur 8081, gRPC sur 50051 (par défaut SDK)
- **Réseau** : Le plugin s'enregistre auprès du registry sur `http://orchestrator:8082` (port du registry, pas le port principal de l'orchestrateur). `PLUGIN_ORCHESTRATOR_URL` = `http://orchestrator:8082`
- **Labels Docker** : Le Dockerfile doit inclure les labels de découverte : `ade.plugin.name=templates`, `ade.plugin.version=1.0.0`, `ade.plugin.http-port=8081`, `ade.plugin.grpc-port=50051`

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `plugins/templates/server_test.go`
- Scénario 1 : `ListTemplates` → retourne la liste des templates embarqués
- Scénario 2 : `RenderTemplate` avec template valide et variables → rendu correct
- Scénario 3 : `RenderTemplate` avec template inconnu → erreur 404
- Scénario 4 : `RenderTemplate` avec variables manquantes en mode strict → erreur
- Scénario 5 : `RenderTemplate` avec body JSON invalide → 400

#### Tests d'intégration
- **Fichier** : `plugins/templates/server_test.go`
- Scénario 6 : Démarrage du SDK PluginServer, appel REST à `/api/v1/templates` → 200 + liste JSON
- Scénario 7 : Démarrage du SDK PluginServer, appel gRPC `ListTemplates` → même résultat que REST

### Documentation
- Mettre à jour `docs/plugins/examples.md` : section "Plugin d'exemple : Templates" avec description complète
- Documenter la procédure de build : `docker build -t ade-plugin-templates plugins/templates`

### Exemples d'utilisation

```dockerfile
# plugins/templates/Dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o /plugin .

FROM gcr.io/distroless/base:latest
COPY --from=builder /plugin /plugin

LABEL ade.plugin.name=templates
LABEL ade.plugin.version=1.0.0
LABEL ade.plugin.http-port=8081
LABEL ade.plugin.grpc-port=50051

EXPOSE 8081 50051
ENTRYPOINT ["/plugin"]
```

```bash
# Build de l'image
docker build -t ade-plugin-templates:latest plugins/templates

# Exécution standalone (pour test)
docker run --rm \
  -e PLUGIN_NAME=templates \
  -e PLUGIN_VERSION=1.0.0 \
  -e PLUGIN_ORCHESTRATOR_URL=http://host.docker.internal:8082 \
  -p 8081:8081 \
  --label ade.plugin.name=templates \
  --label ade.plugin.version=1.0.0 \
  ade-plugin-templates:latest
```

```http
### Lister les templates
GET http://localhost:8081/api/v1/templates

### Rendre un template
POST http://localhost:8081/api/v1/templates/render
Content-Type: application/json

{
    "template_name": "gitignore",
    "variables": {
        "ProjectName": "mon-projet",
        "Lang": "go"
    },
    "strict": false
}
```

### docker-compose.yml (extrait)

```yaml
services:
  orchestrator:
    image: ade-orchestrator:latest
    ports:
      - "8080:8080"   # API publique + proxying vers registry
    networks:
      - ade-network

  registry:
    image: ade-orchestrator:latest
    command: ["ade-orchestrator", "registry"]
    ports:
      - "8082:8082"   # API registry (inscription plugins)
    networks:
      - ade-network
    depends_on:
      - docker-proxy

  docker-proxy:
    image: ade-docker-proxy:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    ports:
      - "8083:8083"
    networks:
      - ade-network

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
    depends_on:
      - registry
    labels:
      ade.plugin.name: "templates"
      ade.plugin.version: "1.0.0"
    restart: unless-stopped

networks:
  ade-network:
    driver: bridge
```
