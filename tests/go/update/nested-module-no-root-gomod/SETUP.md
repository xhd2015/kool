# Scenario

**Feature**: nested Go module inside a git repo that has no root go.mod

Mirrors `dot-pkgs`: the git root has only a README; the module lives in `go-pkgs/`
with tags like `go-pkgs/v0.0.2`.

## Steps

1. Create git repo without root go.mod
2. Add `go-pkgs/` subdirectory with its own go.mod
3. Tag `go-pkgs/v0.0.2`, then add a post-tag commit
4. Create consumer module with require + replace pointing at `go-pkgs/`

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	initNestedModuleFixture(t, req)
	return nil
}
```