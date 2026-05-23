# Configuration des outils agentic

## Description
La commande `ade init` configure automatiquement les outils agentic (OpenCode, Cursor)
et les serveurs MCP pour votre projet.

## Utilisation

### Configuration complète
```powershell
ade init
```

### Configuration avec options
```powershell
# Spécifier le répertoire du projet
ade init --output C:\Projects\mon-app

# Écraser les fichiers existants
ade init --force

# Ignorer certaines étapes
ade init --skip-tools --skip-mcp

# Arrêter à la première erreur
ade init --halt-on-error
```

## Étapes exécutées

1. **Détection des outils** — Vérifie la présence d'OpenCode et Cursor
2. **Installation des skills** — Copie les skills OpenCode dans `.opencode/skills/`
3. **Configuration MCP** — Configure les serveurs MCP dans `.opencode/config.json`

## Configuration YAML
Voir `docs/configuration/yaml.md` pour la configuration des chemins d'outils et des serveurs MCP.
