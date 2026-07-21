# Scenario

**Feature**: sealed runner rejects invalid invocation before (or without) guest exec

```
./sandbox.bin   # no command
  -> non-zero; stderr Error: style (mentions command/usage)
```

## Steps

1. Validation leaves pack a non-empty blob so build succeeds; they vary sealed argv only.
2. Ensure AfterBuildRun remains enabled and double-dash is off by default (no guest flags).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Validation branch: run sealed binary after build; no -- unless a leaf sets it.
	req.AfterBuildRun = true
	req.SealedDoubleDash = false
	return nil
}
```
