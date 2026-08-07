# Scenario

**Feature**: sealed binary security bar (no plaintext secrets; unique seal per build)

```
# no leak
user -> build pack containing UNIQUE_SECRET
  -> strings(binary) must not contain UNIQUE_SECRET

# non-determinism of seal
user -> build same input twice
  -> two output binaries not byte-identical
```

## Steps

1. Security leaves use non-empty packs; no inspect unless leaf opts in.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.AfterBuildInspect = false
	return nil
}
```
