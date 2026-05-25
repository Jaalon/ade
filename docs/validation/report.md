# Formats de rapport de validation

## Rapport JSON

Le rapport JSON est généré par `JSONReporter` et contient la structure
complète des résultats de validation.

### Exemple

```json
{
  "status": "failed",
  "duration": "4.23s",
  "started_at": "2026-05-24T10:30:00Z",
  "completed_at": "2026-05-24T10:30:04Z",
  "num_checks": 4,
  "num_passed": 3,
  "num_failed": 1,
  "modules": [
    {
      "module_name": "golang",
      "status": "failed",
      "duration": "3.84s",
      "checks": [
        {
          "name": "go-version",
          "status": "passed",
          "message": "Go 1.26 trouvé",
          "duration": "0.34s"
        },
        {
          "name": "go-build",
          "status": "passed",
          "message": "Build réussi",
          "duration": "1.23s"
        },
        {
          "name": "go-test",
          "status": "failed",
          "message": "2 tests échoués sur 15",
          "duration": "2.10s",
          "details": "--- FAIL: TestSomething (0.01s)\n    something_test.go:42: expected true, got false"
        },
        {
          "name": "go-vet",
          "status": "passed",
          "message": "Vet réussi",
          "duration": "0.17s"
        }
      ]
    }
  ]
}
```

## Rapport JUnit XML

Le rapport JUnit XML est généré par `JUnitReporter`. Il est compatible
avec les outils CI standards (Jenkins, GitLab CI, GitHub Actions).

### Exemple

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="ade-validate" tests="4" failures="1" errors="0" time="4.23">
  <testsuite name="golang" tests="4" failures="1" errors="0" skipped="0" time="3.84">
    <testcase name="go-version" classname="golang" time="0.34" />
    <testcase name="go-build" classname="golang" time="1.23" />
    <testcase name="go-test" classname="golang" time="2.10">
      <failure message="2 tests échoués sur 15" type="failure">--- FAIL: TestSomething (0.01s)
    something_test.go:42: expected true, got false</failure>
    </testcase>
    <testcase name="go-vet" classname="golang" time="0.17" />
  </testsuite>
</testsuites>
```

## Visualisation web (Story #008)

> ⚠ La visualisation web des rapports de validation sera disponible
> dans Story #008 via l'interface web de l'orchestrateur.

### Principe (vision future)

1. `ade validate run` génère les rapports sur le disque
2. L'orchestrateur (conteneur `ade-config`) expose une API REST
3. Les rapports sont envoyés à l'orchestrateur via `POST /api/v1/reports`
4. L'interface web affiche les rapports avec historique et tendances

### Format pour l'orchestrateur

En V1, les rapports sont sauvegardés dans un format JSON enrichi,
prêt à être consommé par l'orchestrateur :

```json
{
  "format": "ade-validation-report-v1",
  "generated_at": "2026-05-24T10:30:04Z",
  "project_name": "mon-projet",
  "report": { ... }
}
```

### État actuel

| Fonctionnalité | Statut |
|----------------|--------|
| Rapport JSON | ✅ Disponible |
| Rapport JUnit XML | ✅ Disponible |
| Sauvegarde disque | ✅ Disponible |
| Envoi orchestrateur | 📅 Story #008 |
| Visualisation web | 📅 Story #008 |
| Historique tendances | 📅 Story #008 |
