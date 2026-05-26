# Interface Web

L'interface web de l'orchestrateur est une SPA (Single Page Application) React 19 + TypeScript, buildée avec Vite 6.

## Pages

### Tableau de bord
- Statistiques globales (projets, plugins actifs, workflows, rapports)
- Activité récente (derniers événements : enregistrement plugins, créations projets, workflows)

### Projets
- Liste des projets avec recherche
- Création (modal), consultation, suppression
- Détail d'un projet avec ses workflows associés

### Plugins
- Liste des plugins avec statut (HEALTHY/DEGRADED/UNHEALTHY/UNKNOWN)
- Détail d'un plugin (version, capacités, adresses HTTP/gRPC)

### Workflows
- Liste des workflows exécutés
- Détail d'un workflow avec étapes, statuts et sorties

### Rapports
- Liste des rapports de validation
- Détail d'un rapport (modules, checks passés/échoués)

## Développement

```bash
# Installer les dépendances
cd frontend
npm install

# Mode développement (hot reload sur :5173)
npm run dev

# Build production
npm run build

# Build dev (sortie dans frontend/dist/)
npm run build:dev
```

## Architecture

- **Routing** : react-router-dom v7
- **API Client** : fetch natif (pas axios)
- **Styling** : CSS pur avec variables CSS (thème sombre)
- **Build** : Vite 6 → `internal/orchestrator/frontend/dist/` (embarqué dans le binaire Go via `embed.FS`)

## Production

Le frontend est embarqué dans le binaire Go via `//go:embed frontend/dist/*` dans `internal/orchestrator/embed.go`.

Le serveur Go sert les fichiers statiques :

1. Les routes `/api/*` et `/health` sont routées vers l'API REST
2. Les autres routes servent les fichiers statiques (JS, CSS)
3. Les routes inconnues servent `index.html` (SPA fallback)

## Mode développement Go

```bash
ADE_DEV_MODE=true go run ./cmd/ade-config
```

En mode dev, les fichiers sont servis depuis le système de fichiers au lieu de l'embed.
