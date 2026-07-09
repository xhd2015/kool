# Scenario

**Feature**: --help lists both invocation forms and loop flags

```
kool for-every --help
  -> exit 0; stdout mentions for-every, for-every-<duration>, --max-runs,
     --max-failure, --allow-failure
```

## Steps

1. Help already true from parent; no further request fields required.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Help=true inherited; ensure we do not accidentally set a command.
	req.Command = ""
	req.Duration = ""
	return nil
}
```
