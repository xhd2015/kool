# Scenario

**Feature**: --save --force treats no_submit change as tab modified (or rewrites field)

```
existing bots.json tab a without no_submit
  + run bots --tab "[id=a,no_submit=true] echo a" --tab "[id=b] echo b" --save --force
  -> file tab a has no_submit true; diff/stdout shows modified (or file content proves change)
```

## Steps

1. Write bots fixture (a, b; no no_submit fields).
2. Ad-hoc same ids/commands but a gains no_submit=true; Save+Force.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.Save = true
	req.Force = true
	// Keep window_name so only no_submit on a is the semantic change of interest.
	req.WindowName = "local-bots"
	req.Tabs = []string{
		`[id=a,name=a,no_submit=true] echo a`,
		`[id=b,name=b,cwd=/tmp] echo b`,
	}
	return nil
}
```
