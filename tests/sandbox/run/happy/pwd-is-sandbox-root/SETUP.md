# Scenario

**Feature**: guest cwd is the session materialize root under KOOL_SANDBOX_ROOT

```
kool sandbox build -o sandbox.bin --env MARKER=1
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'pwd'
  -> exit 0; stdout is abs path under PARENT (session child, not PARENT itself)
```

## Steps

1. Minimal pack (env only) so build succeeds.
2. Run `sh -c 'pwd'`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = []string{"MARKER=1"}
	req.SealedArgs = []string{"sh", "-c", "pwd"}
	return nil
}
```
