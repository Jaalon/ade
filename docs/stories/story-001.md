# Story #001 : Initialisation du projet Go CLI

## Description
En tant que développeur, je veux initialiser un squelette de projet Go structuré avec un point d'entrée CLI (Cobra) afin de commencer à développer l'outil `ade`.

## Critères d'acceptation
- [ ] Le projet Go est structuré avec `cmd/` et `internal/` selon les bonnes pratiques
- [ ] Le CLI utilise la bibliothèque Cobra pour gérer les sous-commandes
- [ ] Un point d'entrée CLI est défini avec support de sous-commandes (`init`)
- [ ] La commande `ade` détecte si le conteneur de configuration est disponible et le démarre si nécessaire
- [ ] La compilation produit un binaire Windows fonctionnel
- [ ] L'exécution de `ade --help` affiche la documentation du CLI

## Tests automatisés
- Test unitaire : Test de la structure des commandes et sous-commandes Cobra
- Test d'intégration : Test d'exécution du binaire compilé avec `--help`
- Test E2E : Validation du cycle complet (build → execution → sortie)

## Documentation
- `docs/cli/commands.md` — Liste des commandes disponibles
- `README.md` — Instructions de build et d'utilisation de base

## Valeur utilisateur
Permet de démarrer le développement de l'outil avec une architecture propre et testable, en suivant les conventions Go standard avec Cobra.

## Dépendances
- Aucune
