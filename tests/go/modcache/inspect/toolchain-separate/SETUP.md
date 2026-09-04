# Scenario

**Feature**: golang.org/toolchain is not default legacy

```
two toolchain versions + one other module -> inspect --json
  -> toolchainVersions=2, none of toolchain marked legacy
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.JSON = true
	writeExtracted(t, req.ModCache, "golang.org/toolchain", "v0.0.1-go1.21.0.darwin-arm64", "t1")
	writeExtracted(t, req.ModCache, "golang.org/toolchain", "v0.0.1-go1.22.0.darwin-arm64", "t2")
	writeExtracted(t, req.ModCache, "example.com/foo", "v1.0.0", "foo")
	return nil
}
```
