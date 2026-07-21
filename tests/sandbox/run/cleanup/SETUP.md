# Scenario

**Feature**: sealed runner removes the session materialize directory on exit

```
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'true'
  -> exit 0; PARENT has no remaining session children
```

## Steps

1. Cleanup leaves force a known empty parent via `SandboxRootParent`.
2. After successful guest exit, assert parent is empty again.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SealedDoubleDash = true
	return nil
}
```
