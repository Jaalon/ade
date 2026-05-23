# Tâche #004 - Story #002 : Tests automatisés

## Objectif
Implémenter la suite de tests complète pour la Story #002 : tests unitaires (pour les packages non encore couverts), tests d'intégration (exécution de `ade init specs` dans un répertoire temporaire) et test E2E (vérification du binaire compilé).

## Contexte
- Story #002 : `docs/stories/story-002.md`
- Dépend de : Tâches #001, #002, #003
- Nécessaire pour : Tâche #005 (documentation)
- Les tests unitaires pour `internal/templates` et `internal/generator` sont partiellement dans les Tâches #001 et #002 — cette tâche complète la couverture restante et ajoute les tests d'intégration/E2E.

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
Compléter la couverture de tests pour la Story #002 :

1. **Tests unitaires manquants** pour `internal/command/init_specs.go` (si non couverts dans Tâche #003)
2. **Tests d'intégration** : Exécution de `ade init specs` dans un répertoire temporaire avec plugin mock
3. **Test E2E** : Compilation du binaire et exécution complète

**Cas nominaux :**
- Tous les tests unitaires des packages `internal/templates`, `internal/generator` et `internal/command` passent
- Le test d'intégration crée un répertoire temporaire, exécute la commande, vérifie les fichiers créés
- Le test E2E compile le binaire, exécute `ade init specs --force`, vérifie la sortie
- La couverture de code pour `internal/templates/` ≥ 80%

**Cas limites :**
- Test avec un répertoire déjà partiellement initialisé (certains fichiers existent déjà)
- Test avec `--force` et sans `--force`
- Test avec un répertoire en lecture seule (vérification gestion d'erreurs)
- Test avec des templates filtrés

**Gestion d'erreurs :**
- Les tests doivent nettoyer les répertoires temporaires après exécution (via `t.Cleanup`)
- Les tests E2E doivent être derrière un build tag `e2e` pour ne pas s'exécuter par défaut

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `internal/command/init_test.go` | Modifier | Ajouter les scénarios d'intégration |
| `test/e2e/e2e_test.go` | Modifier | Ajouter le scénario E2E pour `init specs` |

### Signatures

```go
// internal/command/init_test.go (ajouts)

// TestInitSpecs_Integration exécute la commande dans un tmp dir et vérifie les fichiers
func TestInitSpecs_Integration(t *testing.T)

// TestInitSpecs_ForceFlag vérifie que --force écrase sans confirmation
func TestInitSpecs_ForceFlag(t *testing.T)

// test/e2e/e2e_test.go (ajouts)

// TestE2E_InitSpecs exécute le binaire compilé avec init specs
func TestE2E_InitSpecs(t *testing.T)
```

### Contraintes techniques
- **Build tags** : Le test E2E doit utiliser `//go:build e2e` (build tag) en première ligne
- **Tests parallèles** : Utiliser `t.Parallel()` pour les tests unitaires indépendants
- **Répertoires temporaires** : Utiliser `t.TempDir()` (auto-nettoyage) ou `os.MkdirTemp` + `t.Cleanup`
- **Mock du prompter** : Dans `internal/generator`, créer un `mockPrompter` pour les tests :

```go
// mockPrompter pour les tests
type mockPrompter struct {
    response bool
    called   bool
}

func (m *mockPrompter) Confirm(msg string) (bool, error) {
    m.called = true
    return m.response, nil
}
```

- **Capture de sortie** : Pour tester l'affichage des commandes Cobra, capturer `os.Stdout` avec un `bytes.Buffer` ou utiliser `cmd.SetOut(buffer)`
- **Tests d'intégration** : Ne doivent pas avoir de dépendances externes (pas de Docker, pas de réseau)
- **Coverage** : Vérifier avec `go test -coverprofile=coverage.out ./internal/templates/...`

### Tests à implémenter

#### Tests d'intégration — `internal/command/init_test.go`

- **Scénario 1** : `TestInitSpecs_Integration`
  - Créer un tmp dir, exécuter la commande `init specs` avec `--output {tmpDir} --force`
  - Vérifier que `.gitignore` existe et contient "*.exe"
  - Vérifier que `.opencode/config.json` existe et contient JSON valide
  - Vérifier que `.opencode/workflow.yaml` existe et contient "specification"
  - Vérifier que les fichiers SKILL.md sont copiés dans `.opencode/skills/`

- **Scénario 2** : `TestInitSpecs_ForceFlag`
  - Créer un tmp dir avec un `.gitignore` vide
  - Exécuter `init specs --output {tmpDir} --force`
  - Vérifier que `.gitignore` a été écrasé (taille > 0)

#### Test E2E — `test/e2e/e2e_test.go`

- **Scénario 1** : `TestE2E_InitSpecs`
  - Builder le projet avec `go build -o {tmpDir}/ade_test.exe ./cmd/ade`
  - Exécuter `{tmpDir}/ade_test.exe init specs --force` dans un tmp dir
  - Vérifier que le code de sortie est 0
  - Vérifier que stdout contient "Fichiers créés" (ou équivalent français)
  - Vérifier que les fichiers `.gitignore`, `.opencode/config.json`, `.opencode/workflow.yaml` existent

### Documentation
- Aucune documentation spécifique (les tests sont auto-documentés par leur nom et leurs commentaires Go)

### Exemples d'utilisation
```bash
# Exécuter les tests unitaires
go test ./internal/templates/... ./internal/generator/... ./internal/command/... -v

# Exécuter avec couverture
go test -coverprofile=coverage.out ./internal/templates/...
go tool cover -html=coverage.out

# Exécuter les tests E2E
go test -tags=e2e -v ./test/e2e/

# Exécuter tous les tests
go test ./... -v
# Tests E2E exclus (pas de -tags=e2e)
```
