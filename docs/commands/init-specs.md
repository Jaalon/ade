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
