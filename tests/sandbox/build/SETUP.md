# Scenario

**Feature**: sandbox build packs and seals a binary (or validates and fails)

```
# validation
user -> kool sandbox build [bad/missing inputs]
  -> stderr error, non-zero

# success
user -> kool sandbox build -o OUT …
  -> sealed binary at OUT; summary on stdout
```

## Steps

1. Subcommand=build for all descendants.
2. Default output name `sandbox.bin` under WorkingDir unless leaf overrides.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HelpAtRoot = false
	req.HelpBuild = false
	req.Subcommand = "build"
	if !req.OutputSet && req.Output == "" {
		req.Output = "sandbox.bin"
		req.OutputSet = true
	}
	return nil
}
```
