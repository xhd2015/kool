# Scenario

**Feature**: sealed binary with no command args fails with usage-style Error

```
kool sandbox build -o sandbox.bin --env MARKER=1
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin
  -> non-zero; stderr mentions command and/or usage (Error: style)
```

## Steps

1. Pack a minimal non-empty env so build succeeds.
2. Invoke sealed binary with empty argv (no guest command).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = []string{"MARKER=1"}
	req.SealedArgs = nil
	req.SealedDoubleDash = false
	return nil
}
```
