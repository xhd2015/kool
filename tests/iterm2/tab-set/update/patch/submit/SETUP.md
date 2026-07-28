# Scenario

**Feature**: `--submit` clears `no_submit` on an existing tab

```
staged.json tab b has no_submit=true
  -> update staged --tab-id b --submit
  -> tab b no_submit omit/false; tab a unchanged
```

## Steps

1. Write fixture with tab b `no_submit: true`.
2. SetName=staged; TabID=b; UpdateSubmit=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const updatePatchSubmitFixtureJSON = `{
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

	writeConfigFile(t, req.ConfigDir, "staged", updatePatchSubmitFixtureJSON)
	req.SetName = "staged"
	req.TabID = "b"
	req.UpdateSubmit = true
	return nil
}
```
