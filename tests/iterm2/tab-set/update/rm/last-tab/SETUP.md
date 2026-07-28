# Scenario

**Feature**: `--rm` of the last remaining tab is an error

```
solo.json has only tab x
  -> update solo --tab-id x --rm --force
  -> error (empty tabs forbidden); file unchanged
```

## Steps

1. Write single-tab fixture.
2. SetName=solo; TabID=x; Rm=true; Force=true (so non-TTY confirm is not the failure reason).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const soloJSON = `{
  "version": 1,
  "window_name": "solo-win",
  "tabs": [
    {"id": "x", "name": "x", "command": "echo solo"}
  ]
}
`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeConfigFile(t, req.ConfigDir, "solo", soloJSON)
	req.SetName = "solo"
	req.TabID = "x"
	req.Force = true
	return nil
}
```
