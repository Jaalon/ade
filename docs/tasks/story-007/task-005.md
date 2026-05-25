# Tâche #005 - Story #007 : Intégration CLI (commandes plugins)

## Objectif
Ajouter au CLI `ade` les commandes permettant de lister, inspecter et interagir avec les plugins via l'API de l'orchestrateur, avec gestion de la dégradation si l'orchestrateur est indisponible.

## Contexte
- Story #007 : `docs/stories/story-007.md`
- Dépend de : Tâche #003 (registry orchestrateur), Story #001 (CLI Cobra)
- Nécessaire pour : Rien

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Ajouter les sous-commandes `ade plugin` au CLI existant. Ces commandes contactent l'orchestrateur via l'API REST pour interagir avec les plugins. Le CLI détecte automatiquement l'URL de l'orchestrateur (via configuration ou variables d'environnement).

**Cas nominaux :**
- `ade plugin list` → affiche la liste des plugins enregistrés avec leur statut (tableau formaté)
- `ade plugin info <name>` → affiche les détails d'un plugin (descriptif, capacités, endpoints, santé)
- `ade plugin install <image>` → télécharge l'image Docker, crée et démarre un conteneur plugin avec les labels appropriés, sur le réseau ade-network
- `ade plugin uninstall <name>` → arrête et supprime le conteneur du plugin
- `ade plugin` seul → affiche l'aide de la commande plugin (comportement Cobra)
- Le CLI détecte l'URL de l'orchestrateur : d'abord variable d'env `ADE_ORCHESTRATOR_URL`, puis config YAML, puis défaut `http://localhost:8080`
- Le CLI essaie d'abord la communication gRPC avec l'orchestrateur ; si gRPC échoue (timeout, non supporté), fallback automatique vers REST

**Cas limites :**
- Orchestrateur injoignable → message clair "Orchestrateur non disponible. Déployez avec 'ade init ci'."
- Aucun plugin enregistré → `ade plugin list` affiche "Aucun plugin enregistré."
- Plugin `info` sur un nom inconnu → message "Plugin 'xyz' introuvable."
- Timeout de connexion à l'orchestrateur → message après 5s

**Gestion d'erreurs :**
- Orchestrateur injoignable → ne pas planter, afficher un message informatif et retourner code 0 (dégradation douce)
- Erreur HTTP de l'orchestrateur → afficher le message d'erreur de l'API
- Format de réponse inattendu → afficher "Erreur de communication avec l'orchestrateur"

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/plugin.go` | Créer | Commande Cobra `ade plugin` |
| `internal/command/plugin_list.go` | Créer | Sous-commande `ade plugin list` |
| `internal/command/plugin_info.go` | Créer | Sous-commande `ade plugin info` |
| `internal/command/plugin_install.go` | Créer | Sous-commande `ade plugin install <image>` |
| `internal/command/plugin_uninstall.go` | Créer | Sous-commande `ade plugin uninstall <name>` |
| `internal/command/plugin_test.go` | Créer | Tests des commandes plugins |
| `internal/orchestrator/client.go` | Créer | Client HTTP pour l'API orchestrateur |
| `internal/orchestrator/client_grpc.go` | Créer | Client gRPC pour l'API orchestrateur (fallback REST) |
| `internal/orchestrator/client_test.go` | Créer | Tests du client |

### Signatures

```go
// internal/command/plugin.go
package command

var pluginCmd = &cobra.Command{
    Use:   "plugin",
    Short: "Gère les plugins Docker",
    Long:  `Liste, inspecte et interagit avec les plugins Docker...`,
}

func init() {
    rootCmd.AddCommand(pluginCmd)
}


// internal/command/plugin_list.go
var pluginListCmd = &cobra.Command{
    Use:   "list",
    Short: "Liste les plugins enregistrés",
    RunE: func(cmd *cobra.Command, args []string) error {
        // ...
    },
}

func init() {
    pluginCmd.AddCommand(pluginListCmd)
}


// internal/command/plugin_info.go
var pluginInfoCmd = &cobra.Command{
    Use:   "info <name>",
    Short: "Affiche les détails d'un plugin",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        // ...
    },
}

func init() {
    pluginCmd.AddCommand(pluginInfoCmd)
}


// internal/orchestrator/client.go
package orchestrator

// Client communique avec l'API de l'orchestrateur (REST avec fallback gRPC).
type Client struct {
    restURL    string
    grpcURL    string
    httpClient *http.Client
    grpcConn   *grpc.ClientConn
    timeout    time.Duration
}

func NewClient(restURL string, grpcURL string) *Client
func (c *Client) ListPlugins(ctx context.Context) ([]registry.PluginInstance, error)
func (c *Client) GetPlugin(ctx context.Context, name string) (*registry.PluginInstance, error)
func (c *Client) Health(ctx context.Context) error
func (c *Client) dialGRPC() error

// Formatage console
func FormatPluginList(plugins []registry.PluginInstance) string
func FormatPluginInfo(plugin *registry.PluginInstance) string


// internal/command/plugin_install.go
// InstallPlugin télécharge et démarre un conteneur plugin Docker.
func installPlugin(image string) error
func uninstallPlugin(name string) error
```

### Contraintes techniques
- **Communication duale** : Le client essaie d'abord gRPC (port 9090 de l'orchestrateur). Si la connexion gRPC échoue dans les 2s, fallback REST sur `http://orchestrateur:8080/api/v1/plugins/`. L'orchestrateur proxyfie vers le registry sur le port 8082.
- **Client HTTP** : Utiliser `net/http` standard, timeout de 5s
- **Client gRPC** : Utiliser `google.golang.org/grpc` avec `grpc.WithInsecure()` et `grpc.WithBlock()` (timeout 2s)
- **Formatage** : Utiliser `text/tabwriter` pour les tableaux alignés dans le terminal
- **Dégradation** : Toute erreur réseau est catchée et affichée comme message友好, sans `os.Exit(1)`
- **Détection orchestrateur** : Variable d'env `ADE_ORCHESTRATOR_URL` > config YAML > défaut `http://localhost:8080` ; pour gRPC : déduire `grpc://localhost:9090` du même hôte
- **Installation plugin** : Utiliser le package `internal/docker/` pour pull l'image et run le conteneur avec les labels ade.plugin.*, sur le réseau ade-network
- **Pattern Cobra** : Suivre le même pattern que `init.go` et `validate.go` (enregistrement dans `init()`)

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `internal/orchestrator/client_test.go`
- Scénario 1 : `NewClient` avec URL → client configuré
- Scénario 2 : `ListPlugins` avec serveur mock → retourne la liste parsée
- Scénario 3 : `ListPlugins` avec serveur injoignable → erreur de connexion
- Scénario 4 : `GetPlugin` avec plugin existant → retourne l'instance
- Scénario 5 : `GetPlugin` avec 404 → erreur "plugin introuvable"
- Scénario 6 : `Health` → pas d'erreur si 200

- **Fichier** : `internal/command/plugin_test.go`
- Scénario 7 : `ade plugin list` sans plugins → "Aucun plugin enregistré."
- Scénario 8 : `ade plugin info nom` → détails formatés
- Scénario 9 : `ade plugin info inconnu` → message d'erreur
- Scénario 10 : `ade plugin` seul → affiche l'aide
- Scénario 11 : `ade plugin install nginx:latest` → vérifier que le Docker client est appelé avec les bons labels
- Scénario 12 : `ade plugin uninstall templates` → vérifier l'arrêt et la suppression du conteneur

#### Tests d'intégration
- Scénario 13 : Client gRPC avec serveur gRPC mock → `ListPlugins` retourne la liste
- Scénario 14 : Serveur gRPC injoignable → fallback REST vers un serveur HTTP mock → `ListPlugins` retourne la liste
- Scénario 15 : Les deux protocoles injoignables → erreur "Orchestrateur non disponible."
- Scénario 16 : Démarrer un serveur mock orchestrateur → `ade plugin list` affiche les plugins

### Documentation
- Mettre à jour `docs/cli/commands.md` : section `ade plugin`
- Créer `docs/commands/plugin.md` : Documentation complète des sous-commandes plugin

### Exemples d'utilisation

```bash
# URL par défaut (http://localhost:8080, grpc://localhost:9090)
ade plugin list

# URL personnalisée
ADE_ORCHESTRATOR_URL=http://192.168.1.50:8080 ade plugin list

# Détails d'un plugin
ade plugin info templates

# Installer un plugin
ade plugin install my-plugin:latest

# Désinstaller un plugin
ade plugin uninstall my-plugin
```

```
$ ade plugin list
NOM          VERSION   STATUT    CAPACITÉS                ADRESSE
templates    1.0.0     HEALTHY   template_provider        http://localhost:8081
github       0.1.0     HEALTHY   issue_importer           http://localhost:8082
slack        0.5.0     DEGRADED  notification             http://localhost:8083

$ ade plugin info templates
Nom         : templates
Version     : 1.0.0
Statut      : HEALTHY
Description : Fournit des templates de projets
API Version : v1

Capacités:
  template_provider v1.0.0 - Fournit des templates de projets

Endpoints:
  REST : http://localhost:8081
  gRPC : localhost:50051
  templates : /api/v1/templates
```
