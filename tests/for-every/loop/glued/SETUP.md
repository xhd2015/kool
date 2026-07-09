# Scenario

**Feature**: glued invocation `kool for-every-<duration> <command>…`

```
kool for-every-<duration> [OPTIONS] <command> [args...]
  -> duration from command suffix; same loop as spaced form
```

## Steps

1. Force glued form.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Glued = true
	return nil
}
```
