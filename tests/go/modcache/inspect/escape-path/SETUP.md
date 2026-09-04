# Scenario

**Feature**: escaped !azure directory reports as github.com/Azure/...

```
github.com/!azure/ok@v1.0.0 -> inspect --json -> path github.com/Azure/ok
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.JSON = true
	writeExtracted(t, req.ModCache, "github.com/Azure/ok", "v1.0.0", "package ok\n")
	return nil
}
```
