# Scenario

**Feature**: --max-runs 3 stops after exactly three successful iterations

```
kool for-every --max-runs 3 10ms echo run-ok
  -> three stdout lines; exit 0
```

## Steps

1. Max-runs 3; printing success command.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Duration = "10ms"
	req.MaxRuns = intPtr(3)
	req.Command = "echo"
	req.Args = []string{"run-ok"}
	return nil
}
```
