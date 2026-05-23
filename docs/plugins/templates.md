# Architecture des templates

## Vue d'ensemble
Le système de templates de `ade` permet de générer des fichiers de configuration pour les projets. Actuellement, les templates sont embarqués dans le binaire Go (V1). L'architecture est conçue pour permettre une évolution vers des templates fournis dynamiquement par des plugins Docker (V2).

## V1 : Templates embarqués (actuel)

### Principe
Les templates sont compilés dans le binaire `ade.exe` via le package Go `embed`.

### Emplacement dans le code source
```
internal/templates/
├── embed.go              # Déclaration //go:embed
├── template.go           # API publique (Render, ListTemplates)
├── errors.go             # Erreurs sentinelles
└── embed/
    ├── gitignore.tmpl
    └── opencode/
        ├── config.json.tmpl
        ├── workflow.yaml.tmpl
        └── skills/
            ├── specification-en/SKILL.md
            ├── specification-fr/SKILL.md
            ├── story-en/SKILL.md
            ├── story-fr/SKILL.md
            ├── tasks-en/SKILL.md
            ├── tasks-fr/SKILL.md
            └── feedback-fr/SKILL.md
```

### Ajouter un nouveau template
1. Créer le fichier `.tmpl` dans `internal/templates/embed/`
2. Ajouter une entrée dans la liste des templates dans `template.go`
3. Le fichier est automatiquement embarqué via `embed.go`

### Variables de template
Les templates peuvent utiliser les variables suivantes via `TemplateData` :
- `{{.ProjectName}}` — Nom du projet
- `{{.GoVersion}}` — Version de Go
- `{{.ModulePath}}` — Module Go
- `{{.Lang}}` — Langue (fr/en)

## V2 : Templates via plugins (futur)
Après l'implémentation de la Story #007 (plugins Docker REST + gRPC), les templates pourront être fournis par des plugins.
Le package `templates` restera l'interface unique (`Render`, `ListTemplates`), mais une implémentation supplémentaire
interrogera l'API des plugins pour obtenir les templates.

### Architecture V2
```
ade (CLI)
  └─ internal/templates/
       ├─ embed/           ← Templates embarqués (inchangé)
       ├─ template.go      ← Interface commune
       └─ plugin/          ← Nouveau : templates via plugins Docker
            └─ fetcher.go  ← Interrogation API REST/gRPC des plugins
```

## Voir aussi
- `docs/commands/init-specs.md` — Utilisation de la commande
- `docs/plugins/architecture.md` — Architecture des plugins Docker
- `docs/plugins/development.md` — Guide de développement d'un plugin
