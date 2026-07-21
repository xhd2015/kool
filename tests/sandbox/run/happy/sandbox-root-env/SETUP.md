# Scenario

**Feature**: guest sees SANDBOX_ROOT equal to its absolute cwd (materialize root)

```
kool sandbox build -o sandbox.bin --env MARKER=1
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'printf %s "$SANDBOX_ROOT"; echo; pwd'
  -> exit 0; SANDBOX_ROOT line equals pwd line; both under PARENT
```

## Steps

1. Minimal pack.
2. Guest prints `$SANDBOX_ROOT` then `pwd` on separate lines.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = []string{"MARKER=1"}
	// Two lines: SANDBOX_ROOT then pwd (absolute).
	req.SealedArgs = []string{"sh", "-c", `printf '%s\n' "$SANDBOX_ROOT"; pwd`}
	return nil
}
```
