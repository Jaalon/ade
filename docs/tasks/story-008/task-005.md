# Tâche #005 - Story #008 : Frontend web (React)

## Objectif

Créer l'interface web (SPA React) pour l'orchestrateur avec les pages de gestion des projets, plugins, workflows et visualisation des rapports. L'application est buildée et embarquée dans le binaire Go via `embed.FS`.

## Contexte

- Story #008 : [docs/stories/story-008.md](../../stories/story-008.md)
- Dépend de : Tâche #004 (serveur de fichiers statiques)
- Nécessaire pour : Tâche #006

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

Créer une application React 19 (SPA) dans `frontend/` qui se build dans `frontend/dist/` pour être embarquée dans le binaire Go.

**Pages à implémenter :**

1. **Dashboard (page d'accueil)** 
   - Statistiques globales : nb projets, plugins actifs, workflows, rapports
   - Activité récente (liste des derniers événements)
   - Statut des connexions (orchestrateur, plugins)

2. **Projets (gestion CRUD)**
   - Liste des projets avec recherche/filtre
   - Création d'un projet (modal ou page dédiée)
   - Modification d'un projet
   - Suppression (avec confirmation)
   - Détail d'un projet (liens vers ses workflows)

3. **Plugins (découverte et gestion)**
   - Liste des plugins avec statut (HEALTHY/DEGRADED/UNHEALTHY)
   - Détail d'un plugin (capacités, endpoints, santé)
   - Désenregistrement d'un plugin

4. **Workflows (visualisation et historique)**
   - Liste des workflows exécutés
   - Détail d'un workflow avec le déroulé des steps
   - Statut visuel (pending, running, completed, failed)

5. **Rapports de validation**
   - Liste des rapports disponibles
   - Visualisation d'un rapport (modules, checks, passed/failed)

**Cas nominaux :**
- Navigation entre toutes les pages sans rechargement
- CRUD complet des projets depuis l'interface
- Consultation des plugins avec statut en direct
- Visualisation de l'historique des workflows

**Cas limites :**
- API non disponible → message "Orchestrateur non joignable" avec option de reconnexion
- Liste vide → message approprié ("Aucun projet", "Aucun plugin", etc.)
- Données en chargement → spinner/skeleton loading
- Erreur API → notification toast avec message d'erreur

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/package.json` | Créer | Dépendances React et scripts de build |
| `frontend/vite.config.ts` | Créer | Configuration Vite pour le build |
| `frontend/tsconfig.json` | Créer | Configuration TypeScript |
| `frontend/index.html` | Créer | Point d'entrée HTML |
| `frontend/src/main.tsx` | Créer | Point d'entrée React |
| `frontend/src/App.tsx` | Créer | Composant racine avec routing |
| `frontend/src/api/client.ts` | Créer | Client API REST de l'orchestrateur |
| `frontend/src/api/types.ts` | Créer | Types TypeScript pour les données API |
| `frontend/src/pages/Dashboard.tsx` | Créer | Page d'accueil / tableau de bord |
| `frontend/src/pages/Projects.tsx` | Créer | Gestion des projets |
| `frontend/src/pages/ProjectDetail.tsx` | Créer | Détail d'un projet |
| `frontend/src/pages/Plugins.tsx` | Créer | Liste des plugins |
| `frontend/src/pages/PluginDetail.tsx` | Créer | Détail d'un plugin |
| `frontend/src/pages/Workflows.tsx` | Créer | Liste des workflows |
| `frontend/src/pages/WorkflowDetail.tsx` | Créer | Détail d'un workflow |
| `frontend/src/pages/Reports.tsx` | Créer | Liste des rapports |
| `frontend/src/pages/ReportDetail.tsx` | Créer | Visualisation d'un rapport |
| `frontend/src/components/Layout.tsx` | Créer | Layout avec navigation |
| `frontend/src/components/StatusBadge.tsx` | Créer | Badge de statut (HEALTHY, FAILED, etc.) |
| `frontend/src/components/LoadingSpinner.tsx` | Créer | Indicateur de chargement |
| `frontend/src/styles/index.css` | Créer | Styles globaux (CSS moderne, minimal) |
| `Dockerfile` | Modifier | Ajouter le stage de build frontend (npm build) |
| `internal/orchestrator/embed.go` | Modifier | Mettre à jour le chemin embed si nécessaire |

### Signatures

```typescript
// frontend/src/api/types.ts
interface Project {
  name: string
  description?: string
  created_at: string
  updated_at: string
  labels?: Record<string, string>
}

interface PluginInfo {
  name: string
  version: string
  status: 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY' | 'UNKNOWN'
  http_address: string
  grpc_address: string
  capabilities: Capability[]
}

interface Capability {
  name: string
  version: string
  description: string
}

interface Workflow {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  started_at?: string
  completed_at?: string
  steps: WorkflowStep[]
  project_name?: string
}

interface WorkflowStep {
  name: string
  status: string
  started_at?: string
  completed_at?: string
  output?: string
}

interface ValidationReport {
  status: string
  duration: string
  num_modules: number
  num_checks: number
  num_passed: number
  num_failed: number
  modules: ValidationModule[]
  format_version: string
}

interface ValidationModule {
  name: string
  status: string
  checks: number
  passed: number
  failed: number
  details?: string
}

interface DashboardStats {
  projects: number
  plugins: number
  workflows: number
  reports: number
  active_plugins: number
}
```

### Contraintes techniques

- **Framework** : React 19 + TypeScript 5
- **Build** : Vite 6 (rapide, ESM natif)
- **Routing** : `react-router-dom` v7
- **Styles** : CSS pur ou Tailwind CSS (selon go.mod — vérifier si tailwind est déjà présent, sinon utiliser CSS modules simples)
- **Pas de bibliothèque UI lourde** (éviter MUI, Ant Design — préférer du CSS léger)
- **API Client** : `fetch` natif, pas de axios
- **Responsive** : L'interface doit être utilisable sur écran 1280px+
- **Build output** : `frontend/dist/` — tous les fichiers statiques dans ce dossier
- **Mode développement** : `npm run dev` pour le hot reload sur le port 5173

### Tests

Les tests du frontend sont allégés pour cette itération (les tests Go couvrent la logique métier). Un test basique de rendu peut être ajouté mais ce n'est pas bloquant.

### Documentation

#### Documentation à mettre à jour
- `docs/orchestrator/web-ui.md` — Compléter le guide avec captures d'écran et description de chaque page

### Mise à jour du Dockerfile

Le `Dockerfile` créé dans la tâche #001 contient un stage frontend placeholder. Cette tâche doit **remplacer** le placeholder par le vrai build :

```dockerfile
# Remplacer le stage frontend placeholder par :
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Dans le stage final, ajouter après COPY --from=builder :
COPY --from=frontend-builder /frontend/dist/ /frontend/dist/
```
