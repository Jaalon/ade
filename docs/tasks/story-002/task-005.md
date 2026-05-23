# Tâche #005 - Story #002 : Documentation

## Objectif
Créer la documentation utilisateur et technique pour la commande `ade init specs`, le système de templates et leur fourniture par les plugins (architecture V2 anticipée).

## Contexte
- Story #002 : `docs/stories/story-002.md`
- Dépend de : Tâches #001, #002, #003, #004
- Nécessaire pour : Rien (dernière tâche de la story)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer et mettre à jour les fichiers de documentation :

1. `docs/commands/init-specs.md` — Documentation complète de la sous-commande `ade init specs`
2. `docs/plugins/templates.md` — Documentation sur l'architecture des templates (V1 embarqués + V2 plugins)
3. Mettre à jour `docs/cli/commands.md` avec la section `init specs`

**Documentation de `init-specs.md` :**
- Description de la commande
- Syntaxe complète avec tous les flags
- Exemples d'utilisation courants
- Explication des fichiers générés
- Note sur la politique de non-écrasement

**Documentation de `templates.md` :**
- Architecture du système de templates (V1 embarqués avec Go `embed`)
- Structure des fichiers de templates dans le projet
- Variables de template disponibles (`TemplateData`)
- API de développement : comment ajouter un nouveau template
- Note sur l'évolution vers la V2 (templates via plugins Docker, Story #007)
- Référence à la documentation du développement de plugins

**Mise à jour de `commands.md` :**
- Ajouter la section "init specs" dans les commandes disponibles
- Lier vers `docs/commands/init-specs.md`

**Cas nominaux :**
- Les fichiers de documentation sont complets et cohérents avec l'implémentation
- Les exemples de commandes sont testables et fonctionnels
- La documentation explique clairement la différence V1 (embarqué) vs V2 (plugins)

**Cas limites :**
- Flags avec valeurs par défaut documentés
- Comportement avec `--force` clairement expliqué
- Note sur le comportement avec `--output` si le répertoire n'existe pas

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/commands/init-specs.md` | Créer | Documentation de la commande `ade init specs` |
| `docs/plugins/templates.md` | Créer | Architecture des templates (V1 + V2) |
| `docs/cli/commands.md` | Modifier | Ajouter la section `init specs` |

### Contenu attendu

#### `docs/commands/init-specs.md`

```markdown
# ade init specs

## Description
Génère les fichiers de configuration locaux du projet :
- `.gitignore` adapté à un projet Go
- Structure des skills : `.opencode/skills/`
- Configuration des serveurs MCP : `.opencode/config.json`
- Workflow de développement : `.opencode/workflow.yaml`

## Syntaxe
```powershell
ade init specs [flags]
```

## Flags

| Flag | Shorthand | Type | Défaut | Description |
|------|-----------|------|--------|-------------|
| `--force` | `-f` | bool | `false` | Écraser les fichiers existants sans confirmation |
| `--output` | `-o` | string | `.` | Répertoire de destination |
| `--name` | | string | nom du répertoire | Nom du projet |
| `--lang` | | string | `fr` | Langue des skills (`fr` ou `en`) |
| `--module` | | string | nom du répertoire | Module Go |

## Exemples

### Génération standard
```powershell
ade init specs
```

### Génération avec écrasement automatique
```powershell
ade init specs --force
```

### Génération dans un répertoire spécifique
```powershell
ade init specs --output C:\Projects\mon-app
```

### Génération en anglais
```powershell
ade init specs --lang en
```

### Génération complète avec paramètres
```powershell
ade init specs --name "mon-projet" --lang fr --module "github.com/user/mon-projet"
```

## Comportement

### Fichiers existants
Par défaut, la commande ne remplace pas les fichiers existants sans confirmation.
Utilisez `--force` pour écraser automatiquement.

### Répertoire de destination
Si le répertoire spécifié par `--output` n'existe pas, il est créé automatiquement.

## Fichiers générés

| Fichier | Description |
|---------|-------------|
| `.gitignore` | Fichier d'exclusion Git pour projet Go |
| `.opencode/config.json` | Configuration des serveurs MCP |
| `.opencode/skills/*/SKILL.md` | Fichiers de définition des skills |
| `.opencode/workflow.yaml` | Workflow de développement |

## Voir aussi
- `docs/plugins/templates.md` — Architecture des templates
- `docs/cli/commands.md` — Liste complète des commandes
```

#### `docs/plugins/templates.md`

```markdown
# Architecture des templates

## Vue d'ensemble
Le système de templates de `ade` permet de générer des fichiers de configuration pour les projets. Actuellement, les templates sont embarqués dans le binaire Go (V1). L'architecture est conçue pour permettre une évolution vers des templates fournis dynamiquement par des plugins Docker (V2).

## V1 : Templates embarqués (actuel)

### Principe
Les templates sont compilés dans le binaire `ade.exe` via le package Go `embed`.

### Emplacement dans le code source
```
internal/templates/
├── embed.go              # Déclaration //go:embed
├── template.go           # API publique (Render, ListTemplates)
├── errors.go             # Erreurs sentinelles
└── embed/
    ├── gitignore.tmpl
    └── opencode/
        ├── config.json.tmpl
        ├── workflow.yaml.tmpl
        └── skills/
            ├── specification-en/SKILL.md
            ├── specification-fr/SKILL.md
            ├── story-en/SKILL.md
            ├── story-fr/SKILL.md
            ├── tasks-en/SKILL.md
            ├── tasks-fr/SKILL.md
            └── feedback-fr/SKILL.md
```

### Ajouter un nouveau template
1. Créer le fichier `.tmpl` dans `internal/templates/embed/`
2. Ajouter une entrée dans la liste des templates dans `template.go`
3. Le fichier est automatiquement embarqué via `embed.go`

### Variables de template
Les templates peuvent utiliser les variables suivantes via `TemplateData` :
- `{{.ProjectName}}` — Nom du projet
- `{{.GoVersion}}` — Version de Go
- `{{.ModulePath}}` — Module Go
- `{{.Lang}}` — Langue (fr/en)

## V2 : Templates via plugins (futur)
Après l'implémentation de la Story #007 (plugins Docker REST + gRPC), les templates pourront être fournis par des plugins.
Le package `templates` restera l'interface unique (`Render`, `ListTemplates`), mais une implémentation supplémentaire
interrogera l'API des plugins pour obtenir les templates.

### Architecture V2
```
ade (CLI)
  └─ internal/templates/
       ├─ embed/           ← Templates embarqués (inchangé)
       ├─ template.go      ← Interface commune
       └─ plugin/          ← Nouveau : templates via plugins Docker
            └─ fetcher.go  ← Interrogation API REST/gRPC des plugins
```

## Voir aussi
- `docs/commands/init-specs.md` — Utilisation de la commande
- `docs/plugins/architecture.md` — Architecture des plugins Docker
- `docs/plugins/development.md` — Guide de développement d'un plugin
```

#### `docs/cli/commands.md` (modification section `init`)

Ajouter sous la commande `init` :
```markdown
  - `specs` — Génère les fichiers de configuration du projet
    - Flags : `--force`, `--output`, `--name`, `--lang`, `--module`
    - Documentation : `docs/commands/init-specs.md`
```

### Documentation
Documentation auto-portante (c'est l'objet de cette tâche).

### Exemples d'utilisation
```powershell
# Afficher la documentation
ade init specs --help

# Lire la documentation
Get-Content docs\commands\init-specs.md
```
