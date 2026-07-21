# Scenario

**Feature**: --goos linux --goarch amd64 produces a binary

```
kool sandbox build -o sandbox.bin --goos linux --goarch amd64 --env X=1
  -> exit 0; binary exists; stdout mentions linux/amd64; optional `file` → ELF
```

## Steps

1. Flags-only pack with cross-compile targets.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = []string{"X=1"}
	req.Goos = "linux"
	req.Goarch = "amd64"
	return nil
}
```
