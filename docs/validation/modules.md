# Création de modules de validation

## Principe

Un module de validation est un type Go qui implémente l'interface
`Validator`. Il peut être créé dans le package `internal/validation/`
ou dans un package externe.

## Structure d'un module

```go
package validation

import "context"

// Exemple : Validateur PostgreSQL
type PostgresValidator struct{}

func NewPostgresValidator() *PostgresValidator {
    return &PostgresValidator{}
}

func (v *PostgresValidator) Name() string {
    return "postgres"
}

func (v *PostgresValidator) Description() string {
    return "Validation de l'environnement PostgreSQL"
}

func (v *PostgresValidator) Detect(ctx context.Context) (bool, error) {
    // Vérifier si psql est disponible
    _, err := exec.LookPath("psql")
    return err == nil, nil
}

func (v *PostgresValidator) Validate(ctx context.Context, cfg ModuleConfig) (*ModuleResult, error) {
    var checks []CheckResult

    // Check 1 : psql disponible
    checks = append(checks, v.checkPsqlAvailable(ctx))

    // Check 2 : connexion à la base
    checks = append(checks, v.checkConnection(ctx))

    return &ModuleResult{
        ModuleName: v.Name(),
        Status:     aggregateStatus(checks),
        Checks:     checks,
    }, nil
}
```

## Enregistrement

```go
func init() {
    Register(NewPostgresValidator())
}
```

## Points clés

1. **Nom unique** : `Name()` doit retourner un identifiant unique
2. **Détection pertinente** : `Detect()` ne doit pas être agressive
   (préférer faux négatif que faux positif)
3. **Checks granulaires** : Chaque vérification atomique est un `CheckResult`
4. **Messages en français** : Les messages utilisateur sont en français
5. **Gestion d'erreurs** : Ne pas paniquer, retourner des erreurs proprement
6. **Timeouts** : Chaque check doit avoir un timeout raisonnable

## Tests

```go
func TestPostgresValidator_Detect(t *testing.T) {
    v := NewPostgresValidator()
    assert.Equal(t, "postgres", v.Name())
}

func TestPostgresValidator_Validate(t *testing.T) {
    v := NewPostgresValidator()
    result, err := v.Validate(context.Background(), ModuleConfig{Name: "postgres"})
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Checks)
}
```
