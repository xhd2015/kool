# Scenario

**Feature**: --save --dry-run plans create without writing file

```
run newsset --tab "echo only" --save --dry-run
  -> exit 0; mentions create / would create; no newsset.json
```

## Steps

1. Empty ConfigDir.
2. Save + DryRun + one tab; no Force needed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "newsset"
	req.Save = true
	req.DryRun = true
	req.Tabs = []string{"echo only"}
	return nil
}
```
