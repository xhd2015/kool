# Scenario

**Feature**: inspect streams [n/total] stage markers on stderr while sizing

```
two extracted versions -> inspect
  -> stderr [1/3] extracted walking/sizing/ok, download, vcs
  -> stdout still has GOMODCACHE report
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.0.0", "old\n")
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.2.0", "new\n")
	return nil
}
```
