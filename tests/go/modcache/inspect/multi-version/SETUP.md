# Scenario

**Feature**: older of two versions is legacy; KEEP is newest

```
foo@v1.0.0 + foo@v1.2.0 -> inspect -> TOP KEEP v1.2.0, path example.com/foo
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
	writeZip(t, req.ModCache, "example.com/foo", "v1.0.0", "OLDZIP")
	writeZip(t, req.ModCache, "example.com/foo", "v1.2.0", "NEWZIP")
	return nil
}
```
