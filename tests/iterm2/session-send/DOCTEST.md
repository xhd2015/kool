# kool iterm2 session send — flag form (`--tab` / `--tab-index` / `--session-id`)

Two distinguishable send forms:

```text
kool iterm2 session <session-id> send [flags] <text>     # positional (unchanged)
kool iterm2 session send (--session-id|…|--tab|…) <text> # flag form (new)
```

Flag form targets exactly one of `--session-id`, `--tab SEL`, or `--tab-index N`.
`--tab` accepts 1-based N or `next|left|right` (`right` ≡ `next`); no wrap.
Tab resolve uses the same window discovery as `kool iterm2 window status`, then
`SendText` (same write path as positional send).

## Version

0.0.1

## DSN (Domain Specific Notion)

### Participants

- **Caller** — `kool iterm2 session … send …`
- **Handler** — `tools/iterm2` routes `session send` vs `session <id> send`
- **Tab select** — `dot-pkgs/shell/iterm2/tabselect` `ParseTabFlag` / `SelectWindowTab`
- **Injected boundary** — `TestRun` ListSessions / CurrentStatus / SendText

### Behaviors

- Flag form requires exactly one session source.
- Positional form rejects `--tab` / `--tab-index` / `--session-id`.
- Success stdout: `sent to session <id>\n`

## Decision Tree

```text
session-send/
├── help/
├── flag-form/
│   ├── tab-next/
│   ├── tab-index/
│   ├── session-id/
│   └── validation/
│       ├── missing-source/
│       ├── conflict-tab-session/
│       ├── conflict-tab-tab-index/
│       ├── missing-text/
│       └── tab-next-at-edge/
└── positional/
    ├── unchanged/
    └── rejects-flag-source/
```

## How to Run

```sh
doctest vet ./tests/iterm2/session-send
doctest test ./tests/iterm2/session-send
```

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/doctest/session"
	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

const (
	fixtureTab1ID = "11111111-2222-3333-4444-555555555555"
	fixtureTab2ID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

type SendCall struct {
	SessionID string
	Text      string
	Opts      lib.SendTextOptions
}

type Request struct {
	Args             []string
	CurrentSessionID string // ITERM-like UUID for current pane (tab resolve)
	ITerm            []lib.SessionRef
	// Help-only when Args empty and Help true → session --help
	Help bool
}

type Response struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	SendCalls []SendCall
}

func defaultITerm() []lib.SessionRef {
	return []lib.SessionRef{
		{WindowID: "1", WindowName: "w", TabIndex: 1, SessionID: fixtureTab1ID, TTY: "/dev/ttys010", Name: "a"},
		{WindowID: "1", WindowName: "w", TabIndex: 2, SessionID: fixtureTab2ID, TTY: "/dev/ttys011", Name: "b"},
	}
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	if req.Help && len(req.Args) == 0 {
		req.Args = []string{"--help"}
	}
	refs := req.ITerm
	if refs == nil {
		refs = defaultITerm()
	}
	cur := req.CurrentSessionID
	if cur == "" {
		cur = fixtureTab1ID
	}
	resp := &Response{}
	env := iterm2cmd.TestRun{
		ListSessions: func() ([]lib.SessionRef, error) { return refs, nil },
		CurrentStatus: &lib.CurrentStatusConfig{
			SessionID:      func() string { return cur },
			ListSessions:   func() ([]lib.SessionRef, error) { return refs, nil },
			ControllingTTY: func() string { return "" },
			AncestorTTYs:   func() []string { return nil },
		},
		SendText: func(sessionID, text string, opts lib.SendTextOptions, cfg *lib.SendTextConfig) error {
			resp.SendCalls = append(resp.SendCalls, SendCall{SessionID: sessionID, Text: text, Opts: opts})
			return nil
		},
	}
	var stdout, stderr bytes.Buffer
	args := append([]string{"session"}, req.Args...)
	resp.ExitCode = iterm2cmd.RunForTestEnv(args, &stdout, &stderr, env)
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	return resp, nil
}
```
