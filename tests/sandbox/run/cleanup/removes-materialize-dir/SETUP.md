# Scenario

**Feature**: after a successful run, session materialize dirs under KOOL_SANDBOX_ROOT are gone

```
kool sandbox build -o sandbox.bin --env MARKER=1
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'true'
  -> exit 0; PARENT empty (or no session children remain)
```

## Steps

1. Minimal pack; guest `true` succeeds.
2. Harness lists `SandboxRootParent` after process exit (`MaterializeEmpty`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = []string{"MARKER=1"}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
