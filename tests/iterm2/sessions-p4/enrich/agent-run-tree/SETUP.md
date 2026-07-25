# Scenario

**Feature**: agent-run → serve → grok tree resolve shows full Unicode connectors

```
# three-level tree (matches procresolve agent-run-tree fixture shape)
AgentResolveByTTY["ttys002"] = agentRunTreeResolve()
  Tree: 200 agent-run → 201 serve → 202 grok
sessions snapshot --no-color
  -> session id
  -> ├── and └── and │ connectors
  -> pids 200/201/202 and cmd substrings agent-run / grok
```

## Steps

1. Replace default resolve with agent-run multi-level tree fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoEnrich = false
	req.NoTree = false
	req.AgentResolveByTTY = map[string]iterm2.AgentResolveFixture{
		"ttys002": agentRunTreeResolve(),
	}
	return nil
}
```
