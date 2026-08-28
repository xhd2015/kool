# kool cloudflare

`kool cloudflare` is a thin CLI that exposes a local HTTP origin on a public
Cloudflare hostname via `github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare`
`StartSession` / `Attach` (same tunnel lifecycle as `lifelog agent serve`, without
agent registration). Primary subcommand: `serve`.

After `StartSession`, `serve` blocks on a public readiness probe (local-bot-style
wait progress) unless `--no-wait-ready` is set. Ready means the public host
responds with any HTTP status other than Cloudflare connector-down (530 + body
`1033`). Timeout prints a yellow `warning:` on stderr and keeps serving.

## Version

0.0.3

## DSN (Domain Specific Notion)

### Participants

- **User** — invokes `kool cloudflare [serve] [flags]`.
- **kool CLI router** — `main.go` dispatches `case "cloudflare"` to
  `tools/cloudflare.Handle` (implementer must wire; these tests call the package
  in-process).
- **cloudflare handler** (`tools/cloudflare`) — root help / subcommand routing /
  `serve` flag parse / validation / tunnel lifecycle / wait-ready / color.
- **StartSession hook** — production default wraps
  `cloudflare.StartSession(SessionOptions{Domain, LocalURL, TunnelName, …})`;
  tests inject a fake so no real `cloudflared` or network runs.
- **WaitReady hook** — production polls the public probe URL; tests inject so no
  network runs. Timeout → warning on stderr, serve continues.
- **WaitSignal hook** — production blocks on SIGINT/SIGTERM; tests inject an
  immediate return so serve exits cleanly after StartSession.
- **Session** — stoppable surface: `Stop() error`, `PublicBaseURL() string`.

### Behaviors

- **Root help** — `kool cloudflare -h|--help` (and handler-equivalent args)
  prints usage mentioning `serve`, `--domain`, `--url`, `--tunnel`, wait/color
  flags; exit 0; stdout ends with `\n`.
- **Serve help** — `kool cloudflare serve -h|--help` documents serve flags
  including `--ready-timeout`, `--no-wait-ready`, `--ready-path`, `--color`,
  `--no-color`; exit 0; stdout ends with `\n`.
- **No subcommand** — bare `kool cloudflare` → non-zero; stderr suggests
  subcommands / help.
- **Unknown subcommand** — e.g. `nosuch` → non-zero; stderr indicates unknown /
  unrecognized command.
- **Serve validation** — missing `--domain` or missing `--url` → non-zero before
  StartSession; message mentions the missing flag/name.
- **Color conflict** — `--color` and `--no-color` together → non-zero;
  `--color and --no-color cannot be specified together`.
- **Serve happy path** — with inject: call StartSession with Domain, LocalURL,
  and TunnelName (derived or explicit); print public URL; WaitReady (unless
  `--no-wait-ready`); WaitSignal; Session.Stop; exit 0.
- **Default tunnel name** — when `--tunnel` omitted/empty: normalize domain
  (trim, strip `http(s)://`, trailing `/`, lower), take leftmost label, slugify
  to `[a-z0-9-]`, prefix **`kool-lb-`**. Example:
  `a.example.com` → `kool-lb-a`; `life-backup.xhd2015.xyz` → `kool-lb-life-backup`.
  Empty slug after slugify → `kool-lb-` + short hash (documented acceptable).
- **Explicit tunnel** — `--tunnel my-tun` passes TunnelName `my-tun` unchanged.
- **Wait-ready ready** — inject success; stdout contains checking line; WaitReady
  called with probe URL `https://<domain>/` (or `--ready-path`).
- **Wait-ready timeout** — inject `ErrReadyTimeout`; stderr contains
  `warning: public ready timeout`; still WaitSignal + Stop; exit 0.
- **`--no-wait-ready`** — Skip WaitReady; print Ctrl+C immediately after tunnel
  lines.
- **StartSession error** — non-zero exit; error surfaced on stderr.
- **No real network** — doctest `Run` always injects StartSession + WaitSignal +
  WaitReady.

### Expected implementer API (contract for GREEN)

Package: `github.com/xhd2015/kool/tools/cloudflare`

```go
// Handle is production entry: HandleWith(args, HandleOpts{}).
func Handle(args []string) error

// HandleWith is the injectable entry used by doctests and optionally by tests of main.
func HandleWith(args []string, opts HandleOpts) error

type HandleOpts struct {
    // StartSession nil → real cloudflare.StartSession adapter.
    StartSession func(SessionStartOptions) (Session, error)
    // WaitSignal nil → block until SIGINT/SIGTERM.
    WaitSignal func() error
    // WaitReady nil → WaitPublicReady against the public probe URL.
    WaitReady func(WaitReadyOptions) (time.Duration, error)
    // Stdout/Stderr nil → os.Stdout / os.Stderr (help + user messages).
    Stdout io.Writer
    Stderr io.Writer
}

type SessionStartOptions struct {
    Domain     string
    LocalURL   string
    TunnelName string
}

type Session interface {
    Stop() error
    PublicBaseURL() string
}
```

`main.go`: `case "cloudflare": return cloudflare.Handle(args)`.

User-facing stdout for a successful serve (shape may vary slightly; soft-assert):

```text
Public URL: https://life-backup.xhd2015.xyz
Tunnel: kool-lb-life-backup
Checking public readiness (timeout 90s)…
Public ready: https://life-backup.xhd2015.xyz/ (200) after 6s
Press Ctrl+C to stop
```

## Decision Tree

```
cloudflare/
├── help/                              [usage; exit 0; no tunnel]
│   ├── root/                          kool cloudflare --help
│   └── serve/                         kool cloudflare serve --help
├── validation/                        [errors before StartSession]
│   ├── no-subcommand/                 bare root args
│   ├── unknown-subcommand/            nosuch
│   ├── missing-domain/                serve without --domain
│   ├── missing-url/                   serve without --url
│   └── color-conflict/                --color + --no-color
└── serve/                             [inject StartSession + WaitReady + WaitSignal]
    ├── happy-derived-tunnel/          domain a.example.com → kool-lb-a; wait ready
    ├── happy-explicit-tunnel/         --tunnel my-tun
    ├── start-error/                   StartSession error → non-zero
    ├── wait-timeout/                  WaitReady timeout → warning; keep serving
    └── no-wait-ready/                 --no-wait-ready skips WaitReady
```

## Test Index

| Leaf | Description |
|------|-------------|
| `help/root/` | Root `--help` exit 0; mentions serve + flags; trailing `\n` |
| `help/serve/` | Serve `--help` exit 0; documents domain/url/tunnel + wait/color flags |
| `validation/no-subcommand/` | No subcommand → non-zero; stderr hints help/commands |
| `validation/unknown-subcommand/` | Unknown subcommand → non-zero |
| `validation/missing-domain/` | Serve without `--domain` → non-zero; mentions domain |
| `validation/missing-url/` | Serve without `--url` → non-zero; mentions url |
| `validation/color-conflict/` | `--color` + `--no-color` → non-zero; conflict message |
| `serve/happy-derived-tunnel/` | Inject: Domain/LocalURL/TunnelName `kool-lb-a`; WaitReady; Stop; exit 0 |
| `serve/happy-explicit-tunnel/` | Inject receives TunnelName `my-tun` |
| `serve/start-error/` | Inject StartSession error → non-zero; stderr surfaces error |
| `serve/wait-timeout/` | WaitReady timeout → `warning:` on stderr; WaitSignal; exit 0 |
| `serve/no-wait-ready/` | `--no-wait-ready` → WaitReady not called; Ctrl+C printed |

## How to Run

```sh
doctest vet ./tests/cloudflare
doctest test ./tests/cloudflare
```

No real `cloudflared` and no network.

```go
import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"

	cf "github.com/xhd2015/kool/tools/cloudflare"
	"github.com/xhd2015/kool/pkgs/errs"
)

// Request drives one in-process tools/cloudflare.HandleWith invocation.
type Request struct {
	// HelpAtRoot: cloudflare --help (ignores Subcommand/flags except WorkingDir).
	HelpAtRoot bool
	// HelpServe: cloudflare serve --help.
	HelpServe bool

	// Subcommand is the first positional after "cloudflare" (e.g. "serve", "nosuch").
	// Empty with HelpAtRoot false and HelpServe false = bare root (no subcommand).
	Subcommand string

	// Serve flags (used when Subcommand=="serve" and not HelpServe).
	Domain    string
	URL       string
	Tunnel    string
	DomainSet bool
	URLSet    bool
	TunnelSet bool

	ReadyTimeout    string
	ReadyTimeoutSet bool
	ReadyPath       string
	ReadyPathSet    bool
	NoWaitReady     bool
	Color           bool
	NoColor         bool

	// AllowStart: when false, injected StartSession fails if called (help/validation).
	// When true, inject records opts and returns fake session unless StartSessionErr set.
	AllowStart bool
	// StartSessionErr: if non-empty and AllowStart, inject returns this error.
	StartSessionErr string

	// WaitReadyMode: "ready" (default when AllowStart), "timeout", or "error".
	// Ignored when NoWaitReady.
	WaitReadyMode string
	// WaitReadyErr is used when WaitReadyMode=="error".
	WaitReadyErr string

	// WorkingDir is reserved for isolation (optional); not required for pure HandleWith.
	WorkingDir string
}

// Response is CLI capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int

	// Inject observations (serve path).
	StartCalled      bool
	StartDomain      string
	StartLocalURL    string
	StartTunnelName  string
	StopCalled       bool
	WaitSignalCalled bool
	WaitReadyCalled  bool
	WaitReadyURL     string
	WaitReadyTimeout time.Duration
}

// fakeSession implements the Session surface expected by tools/cloudflare.
type fakeSession struct {
	publicURL string
	onStop    func() error
}

func (s *fakeSession) PublicBaseURL() string {
	if s == nil {
		return ""
	}
	return s.publicURL
}

func (s *fakeSession) Stop() error {
	if s != nil && s.onStop != nil {
		return s.onStop()
	}
	return nil
}

func buildArgs(req *Request) []string {
	if req.HelpAtRoot {
		return []string{"--help"}
	}
	if req.HelpServe {
		return []string{"serve", "--help"}
	}
	var args []string
	if req.Subcommand != "" {
		args = append(args, req.Subcommand)
	}
	if req.Subcommand == "serve" {
		if req.DomainSet {
			args = append(args, "--domain", req.Domain)
		}
		if req.URLSet {
			args = append(args, "--url", req.URL)
		}
		if req.TunnelSet {
			args = append(args, "--tunnel", req.Tunnel)
		}
		if req.ReadyTimeoutSet {
			args = append(args, "--ready-timeout", req.ReadyTimeout)
		}
		if req.ReadyPathSet {
			args = append(args, "--ready-path", req.ReadyPath)
		}
		if req.NoWaitReady {
			args = append(args, "--no-wait-ready")
		}
		if req.Color {
			args = append(args, "--color")
		}
		if req.NoColor {
			args = append(args, "--no-color")
		}
	}
	return args
}

func mapExit(err error) (int, string) {
	if err == nil {
		return 0, ""
	}
	if se, ok := errs.IsSilenceExitCode(err); ok {
		return se.SilenceExitCode(), ""
	}
	var exitAware interface{ SilenceExitCode() int }
	if errors.As(err, &exitAware) {
		return exitAware.SilenceExitCode(), ""
	}
	return 1, err.Error()
}

// Run invokes tools/cloudflare.HandleWith with injectable StartSession/WaitSignal/WaitReady.
// Never calls real cloudflared or network.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	t.Helper()

	var stdout, stderr bytes.Buffer
	resp := &Response{}

	opts := cf.HandleOpts{
		Stdout: &stdout,
		Stderr: &stderr,
		StartSession: func(o cf.SessionStartOptions) (cf.Session, error) {
			resp.StartCalled = true
			resp.StartDomain = o.Domain
			resp.StartLocalURL = o.LocalURL
			resp.StartTunnelName = o.TunnelName
			if !req.AllowStart {
				return nil, errors.New("unexpected StartSession in this scenario")
			}
			if req.StartSessionErr != "" {
				return nil, errors.New(req.StartSessionErr)
			}
			public := o.Domain
			if public != "" && !strings.HasPrefix(public, "http://") && !strings.HasPrefix(public, "https://") {
				public = "https://" + public
			}
			return &fakeSession{
				publicURL: public,
				onStop: func() error {
					resp.StopCalled = true
					return nil
				},
			}, nil
		},
		WaitSignal: func() error {
			resp.WaitSignalCalled = true
			return nil
		},
		WaitReady: func(o cf.WaitReadyOptions) (time.Duration, error) {
			resp.WaitReadyCalled = true
			resp.WaitReadyURL = o.URL
			resp.WaitReadyTimeout = o.Timeout
			mode := req.WaitReadyMode
			if mode == "" {
				mode = "ready"
			}
			switch mode {
			case "timeout":
				return o.Timeout, cf.ErrReadyTimeout
			case "error":
				msg := req.WaitReadyErr
				if msg == "" {
					msg = "injected wait ready error"
				}
				return 0, errors.New(msg)
			default: // ready
				if o.Log != nil {
					_, _ = io.WriteString(o.Log, cf.FormatPublicReadyLine(o.Style, o.URL, 200, 0))
				}
				return 0, nil
			}
		},
	}

	runErr := cf.HandleWith(buildArgs(req), opts)
	code, errMsg := mapExit(runErr)
	if errMsg != "" {
		if stderr.Len() == 0 || !strings.Contains(stderr.String(), errMsg) {
			fmt.Fprintln(&stderr, errMsg)
		}
	}
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.ExitCode = code
	return resp, nil
}

var _ io.Writer = (*bytes.Buffer)(nil)
```
