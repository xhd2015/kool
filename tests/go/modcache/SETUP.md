# Scenario

**Feature**: kool go modcache inspect / prune

```
user -> HandleWith([inspect|prune] [flags])
  -> inspect reports $GOMODCACHE; prune deletes legacy versions
```

## Preconditions

- Product package `github.com/xhd2015/kool/tools/go/modcache` exports `Handle` /
  `HandleWith` with injectable `Stdout`/`Stderr`.
- `tools/go/go_tools.go` routes `case "modcache"`.
- Fixtures use absolute `--modcache` / `--root` under `req.WorkingDir`.
- No `os.Chdir` / process env mutation.
- `--root` leaves need `git` on PATH.

## Steps

1. Root `Setup` creates an isolated `WorkingDir` under `d.DOCTEST_CASE`.
2. Leaves write a fake module cache and set flags.
3. `Run` calls `modcache.HandleWith`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WorkingDir == "" {
		// TempDir: fake GOMODCACHE uses <path>@<version> dirs; those must
		// not live under this Go module (go list rejects @ in import paths).
		req.WorkingDir = t.TempDir()
	}
	return nil
}
```
