# Commandes CLI

## ade
Commande racine de l'outil Automated Dev Environment.

### Usage
```
ade [command] [flags]
```

### Commandes disponibles
- `init` — Initialise les composants du projet (setup agentic complet)
  - Flags : `--force`, `--output`, `--config`, `--skip-tools`, `--skip-skills`, `--skip-mcp`, `--halt-on-error`
  - Documentation : `docs/agentic/setup.md`
  - `specs` — Génère les fichiers de configuration du projet
    - Flags : `--force`, `--output`, `--name`, `--lang`, `--module`
    - Documentation : `docs/commands/init-specs.md`
  - `ci` — Initialise l'intégration continue
- `version` — Affiche la version

### Flags globaux
- `--help`, `-h` — Affiche l'aide
- `--version` — Affiche la version
