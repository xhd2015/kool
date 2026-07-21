# Scenario

**Feature**: guest exit 42 is returned as the sealed binary exit code

```
kool sandbox build -o sandbox.bin --env MARKER=1
KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'exit 42'
  -> sealed RunExitCode == 42
```

## Steps

1. Minimal pack; guest exits 42.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = []string{"MARKER=1"}
	req.SealedArgs = []string{"sh", "-c", "exit 42"}
	return nil
}
```
