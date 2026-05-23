# Tâche #004 - Story #001 : Détection du conteneur de configuration (stub)

## Objectif
Implémenter un service de détection et de démarrage du conteneur de configuration Docker, utilisé par la commande racine et les sous-commandes.

## Contexte
- Story #001 : `docs/stories/story-001.md`
- Story #008 : `docs/stories/story-008.md` (conteneur de configuration futur)
- Plan d'implémentation : Phase 1 — ce n'est qu'un stub qui vérifie la présence de Docker
- Dépend de : Tâche #002
- Nécessaire pour : Tâche #005

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Créer un service de détection Docker (stub) qui vérifie si Docker/Podman est disponible et peut détecter un conteneur.

**Cas nominaux :**
- `docker.Check()` retourne `(true, nil)` si Docker est disponible dans le PATH
- `docker.EnsureConfigContainer(ctx)` vérifie si un conteneur nommé "ade-config" tourne
- Si le conteneur n'existe pas, affiche un message : "Conteneur de configuration non trouvé. Les fonctionnalités avancées nécessitent le déploiement via 'ade init ci'."
- La fonction retourne sans erreur (stub — ne tente pas vraiment de démarrer un conteneur)

**Cas limites :**
- Docker non installé → message d'information, pas de blocage
- Docker installé mais démon non lancé → message d'information
- La commande `ade` fonctionne parfaitement sans Docker pour les opérations de base

**Gestion d'erreurs :**
- Docker indisponible → loguer un warning, continuer normalement
- Pas d'erreur fatale — le CLI reste utilisable

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/docker/docker.go` | Créer | Service de détection Docker |
| `internal/docker/docker_test.go` | Créer | Tests du service |
| `internal/command/root.go` | Modifier | Appeler la détection Docker au démarrage |

### Signatures

```go
// internal/docker/docker.go
package docker

// Check vérifie si Docker ou Podman est disponible dans le PATH.
// Retourne le nom du binaire trouvé ("docker" ou "podman") ou une erreur.
func Check() (binaryName string, err error)

// IsContainerRunning vérifie si un conteneur Docker avec le nom donné est en cours d'exécution.
// Retourne true si le conteneur tourne.
func IsContainerRunning(ctx context.Context, containerName string) (bool, error)

// EnsureConfigContainer vérifie l'état du conteneur de configuration.
// Si le conteneur n'est pas trouvé ou ne tourne pas, affiche un message d'information.
func EnsureConfigContainer(ctx context.Context) error
```

### Contraintes techniques
- **SDK Docker** : Utiliser `github.com/docker/docker/client` pour toute interaction avec Docker (pas `os/exec`). Ce SDK est bien plus robuste que le parsing de CLI.
- **`context.Context`** : Toutes les fonctions acceptent un `context.Context`
- **Détection** : `client.NewClientWithOpts(client.FromEnv)` pour initialiser le client Docker. Vérifier la connexion avec `client.Ping()`.
- **Stub** : `EnsureConfigContainer` ne fait que vérifier et afficher un message
- **Architecture** : Définir une interface `Client` dans `internal/docker/docker.go` pour permettre le mock dans les tests
- **Logging** : Utiliser `fmt.Fprintf(os.Stderr, ...)` pour les messages d'information (pas de bruit sur stdout)

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/docker/docker_test.go`
- Scénario 1 : `Check()` retourne "docker" si docker.exe est dans le PATH
- Scénario 2 : `Check()` retourne "podman" si podman.exe est dans le PATH (et pas docker)
- Scénario 3 : `Check()` retourne une erreur si ni docker ni podman n'est trouvé
- Scénario 4 : `EnsureConfigContainer()` ne retourne pas d'erreur (stub)
- Scénario 5 : `IsContainerRunning()` avec contexte expiré retourne une erreur de timeout

### Documentation
- `docs/cli/prerequisites.md` : Prérequis système (Docker/Podman)

### Exemples d'API
```bash
# Vérification manuelle équivalente
docker ps --filter name=ade-config --format "{{.Names}}"
```
