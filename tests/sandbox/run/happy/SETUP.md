# Scenario

**Feature**: guest process sees materialized files/env with cwd = SANDBOX_ROOT

```
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c '…'
  -> exit 0; stdout reflects pack + materialize root
```

## Steps

1. Happy leaves pack files/env as needed and set `SealedArgs` for the guest command.
2. Prefer `sh -c '…'` for portable one-liners on macOS/Linux hosts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Optional: end runner flags with -- before guest argv.
	req.SealedDoubleDash = true
	return nil
}
```
