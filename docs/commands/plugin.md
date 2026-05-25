# Commande `ade plugin`

Gère les plugins Docker de l'environnement de développement.

## Usage

```
ade plugin [command] [flags]
```

## Sous-commandes

### list

Liste les plugins enregistrés auprès de l'orchestrateur.

```
ade plugin list
```

Exemple de sortie :

```
NOM          VERSION   STATUT    CAPACITÉS                ADRESSE HTTP
templates    1.0.0     HEALTHY   template_provider        localhost:8081
```

### info

Affiche les détails d'un plugin spécifique.

```
ade plugin info <name>
```

Exemple :

```
$ ade plugin info templates
Nom         : templates
Version     : 1.0.0
Statut      : HEALTHY
Description : Fournit des templates de projets
API Version : v1
HTTP        : localhost:8081
gRPC        : localhost:50051

Capacités:
  template_provider 1.0.0 - Provides project templates

Endpoints:
  http : :8081
  grpc : :50051
```

### install

Télécharge une image Docker et démarre un conteneur plugin.

```
ade plugin install <image>
```

Exemple :

```
ade plugin install my-plugin:latest
```

### uninstall

Arrête et supprime un conteneur plugin.

```
ade plugin uninstall <name>
```

Exemple :

```
ade plugin uninstall my-plugin
```

## Configuration

Le CLI détecte l'URL de l'orchestrateur dans cet ordre :
1. Variable d'environnement `ADE_ORCHESTRATOR_URL`
2. Défaut : `http://localhost:8080`

## Gestion des erreurs

- Orchestrateur injoignable : message informatif sans code d'erreur
- Plugin introuvable : message "Plugin 'xyz' introuvable."
- Aucun plugin : "Aucun plugin enregistré."
