# Tâche #001 - Story #004 : Templates docker-compose.yml et .env

## Objectif
Ajouter les templates `docker-compose.yml` et `.env` au système de templates embarqués, enregistrer les templates dans le registre, et étendre `TemplateData` avec les champs nécessaires.

## Contexte
- Story #004 : `docs/stories/story-004.md`
- Plan d'implémentation : Phase 1, après Story #001. Parallélisable avec Story #002.
- Dépend de : Story #001 (système de templates, package `internal/templates/`)
- Nécessaire pour : Tâche #002

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Ajouter deux nouveaux templates au système existant :

1. **Template docker-compose.yml** : Génère un fichier `docker-compose.yml` définissant :
   - Un service `ade-config` (conteneur de configuration, sera implémenté dans Story #008)
   - Un réseau Docker dédié
   - L'utilisation des variables d'environnement depuis `.env`

2. **Template .env** : Génère un fichier `.env` avec les variables d'environnement par défaut pour l'environnement de préproduction.

**Cas nominaux :**
- `templates.Render("docker-compose", data)` retourne un `docker-compose.yml` valide contenant le service `ade-config`
- `templates.Render("env", data)` retourne un `.env` valide avec toutes les variables définies
- Les deux templates apparaissent dans `templates.ListTemplates()`
- Les templates utilisent `{{.Compose.ConfigPort}}` et `{{.Compose.Network}}` du `TemplateData`

**Cas limites :**
- `Compose.ConfigPort` vide → port par défaut `8080` doit être défini par l'appelant
- `Compose.Network` vide → réseau par défaut `ade-network` doit être défini par l'appelant
- `ProjectName` vide → nom par défaut `preprod` doit être défini par l'appelant

**Gestion d'erreurs :**
- Si les champs `Compose.ConfigPort`, `Compose.Network` ou `ProjectName` ne sont pas définis, le template utilise les valeurs passées (le template lui-même ne définit pas de defaults)
- Si le template ne peut pas être parsé (fichier `.tmpl` invalide), `buildRegistry()` doit le signaler comme une erreur de parsing

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/templates/embed/docker-compose.yml.tmpl` | Créer | Template docker-compose.yml pour la préproduction |
| `internal/templates/embed/env.tmpl` | Créer | Template .env pour les variables d'environnement |
| `internal/templates/template.go` | Modifier | Ajouter ConfigPort et ComposeNetwork à TemplateData, enregistrer les 2 nouveaux templates |
| `internal/templates/embed.go` | Modifier | Ajouter les chemins des nouveaux fichiers dans le `//go:embed` |
| `internal/templates/template_test.go` | Modifier | Ajouter les tests pour les nouveaux templates |

### Signatures

```go
// internal/templates/template.go — TemplateData étendu
type TemplateData struct {
    ProjectName string
    GoVersion   string
    ModulePath  string
    Lang        string
    Compose     ComposeConfig // Configuration docker-compose
}

type ComposeConfig struct {
    ConfigPort string // Port du conteneur de configuration (web UI), défaut "8080"
    Network    string // Nom du réseau Docker, défaut "ade-network"
}
```

### Contraintes techniques
- **Go `embed`** : Ajouter les nouveaux fichiers `.tmpl` dans le `//go:embed` de `embed.go`. Conserver la ligne existante (ne pas la recréer), ajouter les nouveaux chemins à la suite. La directive `//go:embed` doit contenir tous les fichiers sur une seule ligne.
- **`text/template`** : Les templates sont parsés avec `text/template` comme les autres. Pas de logique complexe dans les templates.
- **Convention de nommage** : Les templates s'appellent `docker-compose` et `env`. Leurs chemins embarqués sont `embed/docker-compose.yml.tmpl` et `embed/env.tmpl`.
- **Backward compatibility** : Le nouveau champ `Compose` a la valeur zéro `ComposeConfig{}` (tous ses champs à `""`). Les templates existants n'utilisent pas ce champ, donc ils ne sont pas impactés. La valeur zéro de `ComposeConfig` ne pose pas de problème dans les templates existants.
- **Descriptions en français** : Les descriptions des templates dans le registre doivent être en français (cf. convention existante dans `template.go`).

### Structure des templates

#### `embed/docker-compose.yml.tmpl`
```yaml
version: "3.8"
name: {{.ProjectName}}-preprod

services:
  ade-config:
    image: nginx:alpine
    container_name: ade-config
    ports:
      - "{{.Compose.ConfigPort}}:80"
    env_file:
      - .env
    networks:
      - {{.Compose.Network}}
    restart: unless-stopped

networks:
  {{.Compose.Network}}:
    driver: bridge
```

#### `embed/env.tmpl`
```env
# Environnement de pr\u00e9production - g\u00e9n\u00e9r\u00e9 par ade init ci
ADE_PROJECT_NAME={{.ProjectName}}
ADE_CONFIG_PORT={{.Compose.ConfigPort}}
ADE_COMPOSE_NETWORK={{.Compose.Network}}
ADE_LOG_LEVEL=info
```

### Tests à implémenter

#### Tests unitaires — `internal/templates/template_test.go`
- Scénario 1 : `Render("docker-compose", defaultData)` retourne un YAML contenant "nginx:alpine"
  - Données : `TemplateData{ProjectName: "mon-projet", Compose: ComposeConfig{ConfigPort: "9090", Network: "mon-network"}}`
  - Résultat attendu : la sortie contient "nginx:alpine", "9090:80", "mon-network"
- Scénario 2 : `Render("env", defaultData)` retourne un .env contenant "ADE_PROJECT_NAME"
  - Données : `TemplateData{ProjectName: "mon-projet", Compose: ComposeConfig{ConfigPort: "9090", Network: "mon-network"}}`
  - Résultat attendu : la sortie contient "ADE_PROJECT_NAME=mon-projet" et "ADE_CONFIG_PORT=9090"
- Scénario 3 : `ListTemplates()` inclut "docker-compose" et "env"
  - Résultat attendu : les noms "docker-compose" et "env" sont dans la liste
- Scénario 4 : `Render("docker-compose", minimalData)` avec valeurs par défaut
  - Données : `TemplateData{ProjectName: "test"}` (Compose non défini, valeur zéro)
  - Résultat attendu : la sortie contient "nginx:alpine", ":80" et "network: \n" (la chaîne vide produit ":" et "")

### Documentation
Aucune documentation spécifique pour cette tâche (la documentation sera couverte par la Tâche #003).

### Exemples d'utilisation
```go
data := templates.TemplateData{
    ProjectName: "mon-app",
    Compose: templates.ComposeConfig{
        ConfigPort: "9090",
        Network:    "ade-net",
    },
}
compose, _ := templates.Render("docker-compose", data)
env, _ := templates.Render("env", data)
fmt.Println(compose)
fmt.Println(env)
```
