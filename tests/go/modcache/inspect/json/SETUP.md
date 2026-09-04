# Scenario

**Feature**: inspect --json reports newest and legacy versions

```
foo@v1.0.0 + foo@v1.2.0 -> inspect --json
  -> newest v1.2.0; v1.0.0 legacy true; v1.2.0 legacy false
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.JSON = true
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.0.0", "old\n")
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.2.0", "new\n")
	return nil
}
```
