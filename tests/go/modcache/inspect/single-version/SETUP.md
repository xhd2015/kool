# Scenario

**Feature**: a module with one version is not legacy

```
example.com/foo@v1.2.0 only -> inspect -> not in TOP legacy
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.2.0", "package foo\n")
	writeZip(t, req.ModCache, "example.com/foo", "v1.2.0", "ZIP")
	return nil
}
```
