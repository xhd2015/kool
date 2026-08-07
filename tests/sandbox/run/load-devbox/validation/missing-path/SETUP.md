# Scenario

**Feature**: missing absolute --load-devbox path fails hard

```
./primary --load-devbox /abs/.../does-not-exist.bin -- sh -c 'true'
  -> non-zero; stderr Error:
```

## Steps

1. Primary packs minimal env/file so build succeeds.
2. SealedLoadDevbox points at abs path under WorkingDir that is never created.
3. No SecondaryPacks.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ExtraEnv = []string{"MARKER=1"}
	missing := filepath.Join(req.WorkingDir, "missing-load-targets", "no-such-sealed.bin")
	req.SealedLoadDevbox = []string{missing}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
