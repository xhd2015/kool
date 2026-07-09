# Scenario

**Feature**: spaced invocation `kool for-every <duration> <command>…`

```
kool for-every [OPTIONS] <duration> <command> [args...]
  -> parse duration positional, then run loop
```

## Steps

1. Force spaced form (Glued=false).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Glued = false
	return nil
}
```
