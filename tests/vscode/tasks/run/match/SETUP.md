# Scenario

**Feature**: run label matching — exact, unique substring, ambiguous, not found

```
run <query> --dry-run
  exact first; else unique CI substring; else error
```

## Steps

1. DryRun=true for match leaves (plan path only).
2. Leaves set Query and fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	return nil
}
```
