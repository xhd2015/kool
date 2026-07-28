# Scenario

**Feature**: version != 1 is rejected

```
{"version":2,...} -> show badver -> Error
```

## Steps

1. Write badver.json with version 2; show badver.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeConfigFile(t, req.ConfigDir, "badver", `{
  "version": 2,
  "tabs": [{"id": "a", "name": "a", "command": "echo a"}]
}
`)
	req.SetName = "badver"
	req.Subcommand = "show"
	return nil
}
```
