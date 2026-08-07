# Scenario

**Feature**: relative --load-devbox path is rejected at runtime

```
./primary --load-devbox relative/path -- sh -c 'true'
  -> non-zero; stderr Error: (absolute / relative / load-devbox)
```

## Steps

1. Primary packs MARKER env.
2. SealedLoadDevbox = relative path string as-is (not Abs-resolved).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ExtraEnv = []string{"MARKER=1"}
	req.SealedLoadDevbox = []string{"relative/load-target.bin"}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
