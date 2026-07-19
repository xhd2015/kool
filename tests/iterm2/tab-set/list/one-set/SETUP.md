# Scenario

**Feature**: list shows a configured set name

```
bots.json in config dir -> list -> stdout contains bots and tab count hint
```

## Steps

1. Write `bots.json` fixture.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeBotsConfig(t, req.ConfigDir)
	return nil
}
```
