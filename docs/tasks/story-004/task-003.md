# Tâche #003 - Story #004 : Documentation de la commande et du déploiement

## Objectif
Créer la documentation complète pour la commande `ade init ci`, l'environnement de préproduction locale, et le conteneur de configuration.

## Contexte
- Story #004 : `docs/stories/story-004.md`
- Dépend de : Tâche #002 (commande `ade init ci` implémentée)
- Nécessaire pour : Aucune (documentation finale)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer trois fichiers de documentation :

1. **`docs/commands/init-ci.md`** : Documentation utilisateur de la sous-commande `ade init ci`
2. **`docs/deployment/preprod.md`** : Guide de l'environnement de préproduction
3. **`docs/deployment/config-container.md`** : Documentation du conteneur de configuration

**Cas nominaux :**
- `docs/commands/init-ci.md` décrit toutes les options et flags de `ade init ci`
- `docs/deployment/preprod.md` explique le workflow complet de déploiement
- `docs/deployment/config-container.md` décrit le rôle du conteneur `ade-config`
- Les trois fichiers sont en Markdown, en français, lisibles et bien structurés

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `docs/commands/init-ci.md` | Créer | Documentation de la sous-commande `ade init ci` |
| `docs/deployment/preprod.md` | Créer | Guide de l'environnement de préproduction |
| `docs/deployment/config-container.md` | Créer | Documentation du conteneur de configuration |

### Contenu attendu

#### `docs/commands/init-ci.md`
```markdown
# ade init ci

## Description
Initialise et déploie l'environnement de préproduction locale via Docker Compose.

Cette commande détecte automatiquement Docker ou Podman, génère les fichiers
de configuration nécessaires et déploie les conteneurs.

## Utilisation
\`\`\`
ade init ci [flags]
\`\`\`

## Flags
| Flag | Court | Défaut | Description |
|------|-------|--------|-------------|
| `--output` | `-o` | `.` | Répertoire de sortie pour les fichiers générés |
| `--force` | `-f` | `false` | Écraser les fichiers existants sans confirmation |
| `--name` | | (nom du répertoire) | Nom du projet pour le déploiement |
| `--port` | | `8080` | Port du conteneur de configuration (web UI) |
| `--network` | | `ade-network` | Nom du réseau Docker |

## Exemples
\`\`\`bash
# Déploiement de base
ade init ci

# Avec options personnalisées
ade init ci --output ./preprod --port 9090 --name mon-app

# Forcer l'écrasement des fichiers existants
ade init ci --force
\`\`\`

## Dépendances
- Docker Desktop (Windows) ou Podman
- Docker Compose (plugin inclus dans Docker Desktop)
- Pour Podman : `podman-compose`

## Dépannage
| Problème | Solution |
|----------|----------|
| "Docker ou Podman requis" | Installer Docker Desktop depuis https://www.docker.com/products/docker-desktop/ |
| "Démon inaccessible" | Démarrer Docker Desktop depuis le menu Démarrer |
| "Commande compose non trouvée" | Mettre à jour Docker Desktop ou installer podman-compose |
```

#### `docs/deployment/preprod.md`
```markdown
# Environnement de préproduction locale

## Vue d'ensemble
L'environnement de préproduction locale permet de tester l'application dans
des conditions proches de la production, directement sur la machine du développeur.

## Architecture
L'environnement est déployé via Docker Compose et inclut :
- **ade-config** : Conteneur de configuration avec interface web (voir config-container.md)
- **Réseau dédié** : Les conteneurs communiquent via un réseau Docker isolé

## Workflow

### 1. Initialisation
\`\`\`bash
ade init ci
\`\`\`
Cette commande :
1. Détecte Docker ou Podman
2. Vérifie que le démon est accessible
3. Génère \`docker-compose.yml\` et \`.env\` dans le répertoire courant
4. Déploie les conteneurs avec \`docker compose up -d\`
5. Affiche le statut des conteneurs

### 2. Gestion du cycle de vie
\`\`\`bash
# Voir le statut
docker compose ps

# Voir les logs
docker compose logs -f

# Arrêter les conteneurs
docker compose down

# Redémarrer
docker compose restart
\`\`\`

### 3. Configuration
Les variables d'environnement sont définies dans \`.env\` (généré par \`ade init ci\`) :

| Variable | Défaut | Description |
|----------|--------|-------------|
| \`ADE_PROJECT_NAME\` | (nom du répertoire) | Nom du projet |
| \`ADE_CONFIG_PORT\` | \`8080\` | Port du conteneur de configuration |
| \`ADE_COMPOSE_NETWORK\` | \`ade-network\` | Nom du réseau Docker |
| \`ADE_LOG_LEVEL\` | \`info\` | Niveau de log (debug, info, warn, error) |
```

#### `docs/deployment/config-container.md`
```markdown
# Conteneur de configuration (ade-config)

## Rôle
Le conteneur \`ade-config\` est le point central d'orchestration de
l'environnement de développement. Il expose :
- Une **API REST/gRPC** pour interagir avec les plugins et services
- Une **interface web** pour la configuration visuelle

## Déploiement
Le conteneur est automatiquement inclus dans le \`docker-compose.yml\` généré
par \`ade init ci\`. Il est défini comme suit :

\`\`\`yaml
services:
  ade-config:
    image: nginx:alpine
    container_name: ade-config
    ports:
      - "\${ADE_CONFIG_PORT:-8080}:80"
    env_file:
      - .env
    restart: unless-stopped
\`\`\`

## Personnalisation
- **Port** : Modifier \`ADE_CONFIG_PORT\` dans \`.env\` ou utiliser le flag \`--port\`
- **Réseau** : Modifier \`ADE_COMPOSE_NETWORK\` dans \`.env\`

## Notes
- **V1** : Le conteneur utilise \`nginx:alpine\` comme image placeholder. Il est fonctionnel (sert une page nginx par défaut) mais sera remplacé par l'image réelle dans la Story #008.
- L'implémentation complète du conteneur de configuration avec API et web UI est prévue dans la Story #008.
- Le service est défini dans docker-compose.yml et sera déployé automatiquement par \`ade init ci\`.
```

### Contraintes techniques
- **Format** : Markdown, encodé en UTF-8
- **Langue** : Français
- **Style** : Clair, concis, avec des tableaux pour les flags et la configuration
- **Liens** : Utiliser des liens relatifs entre fichiers de documentation si nécessaire
- **Images** : Pas d'images (documentation purement textuelle)

### Tests à implémenter
Aucun test automatisé pour la documentation. Vérifier manuellement que :
- Les fichiers se lisent correctement dans VS Code / GitHub
- Les liens entre fichiers sont valides
- Les exemples de commandes sont corrects

### Documentation
Cette tâche **est** la documentation. Aucune documentation supplémentaire.

### Exemples d'utilisation
```bash
# Afficher la documentation en ligne
ade init ci --help

# Lire la documentation
type docs\commands\init-ci.md
```
