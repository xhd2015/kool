# Scenario

**Feature**: build from config directory (`-i`)

```
user -> kool sandbox build -o OUT -i DIR
  -> merge files/ + env.yaml (+ optional meta.yaml) → sealed OUT
```

## Steps

1. Leaves write fixture dirs under WorkingDir and set Input/InputSet.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.BuildTwice = false
	req.AfterBuildInspect = false
	return nil
}
```
