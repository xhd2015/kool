# Scenario

**Feature**: show prints no_submit for tabs that set it (version-1 optional field)

```
show staged -> tab b has no_submit=true; tab a (omit) does not claim no_submit
```

## Steps

1. Write version-1 fixture: tab a omits `no_submit`; tab b has `"no_submit": true`.
2. SetName=staged; Subcommand show (inherited).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const stagedNoSubmitJSON = `{
  "version": 1,
  "window_name": "local-staged",
  "tabs": [
    {"id": "a", "name": "a", "command": "echo a"},
    {"id": "b", "name": "b", "command": "grok --resume", "cwd": "/tmp", "no_submit": true}
  ]
}
`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeConfigFile(t, req.ConfigDir, "staged", stagedNoSubmitJSON)
	req.SetName = "staged"
	return nil
}
```
