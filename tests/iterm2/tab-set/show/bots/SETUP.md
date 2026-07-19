# Scenario

**Feature**: show bots fixture details

```
show bots -> local-bots, tab ids a/b, commands echo a / echo b
```

## Steps

1. Write bots.json; SetName=bots.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	return nil
}
```
