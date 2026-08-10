# kool iterm2 sessions auto-backup

Long-running in-process loop that periodically checkpoints critical
**grok** / **codex** / **mark** tabs (same filter as `sessions save`) so crash
recovery can `sessions restore --file <auto-path>`.

**Sibling of** `./tests/iterm2/sessions-save` (manual save/restore). Default
auto path is **`~/.config/iterm2/sessions-auto.json`** — distinct from
manual `sessions-save.json`.

**Out of scope (no leaves):** install / uninstall / status / launchd /
crontab / multi-file rotation / pidfile lock.

## Version

0.0.2

**Classic TDD (greenfield):** all leaves are intentionally **RED** until the
implementer lands `kool iterm2 sessions auto-backup`. Today the handler is
unknown-subcommand. Prefer `--once` on every write leaf (no real 10m sleep).

## DSN (Domain Specific Notion)

### Participants

- **Caller** — invokes `kool iterm2 sessions auto-backup [options]`.
- **kool CLI / handler** — `tools/iterm2` routes `sessions auto-backup`;
  flags `--interval`, `--file`, `--once`, `--dry-run`, `--color` /
  `--no-color`, `--ignore-macos-space`, `--spaces`, `-h/--help`.
- **Interval parser** — `pkgs/duration.Parse` (`10m`, `60s`, bare seconds);
  default **10m**. Invalid → `Error:` + non-zero before loop.
- **SnapshotCollector** — injectable via `InstallPhasedFixtureCollectorForTest`;
  same critical filter as save (grok/codex session_id, live mark).
- **Auto checkpoint writer** — always overwrites the auto file on successful
  non-empty save (no TTY prompt even if pending). Uses `SaveDocument`
  version 1 + atomic write. Prefer `source: kool-iterm2-sessions-auto`.
- **Zero-critical guard** — message + **no write / no clobber** of existing file.
- **Soft-fail path** — capture / iTerm fail → `warning:` on stderr, exit 0 for
  that cycle (with `--once` process exits 0); keep prior file.
- **Fixture installer (tests)** — phased critical / idle fixtures,
  `SeedDoc`, `FailSnapshotCapture`. No live iTerm / launchd.

### Behaviors

- First tick **immediate**, then sleep interval (live loop only).
- **`--dry-run`**: one plan cycle then **exit** (no interval loop; no write).
- L2 live write leaves use `--once` (no real 10m sleep). Dry-run leaf omits `--once`.
- Success cycle stdout ≈ `Saved N critical sessions …` + path line.
- Zero: `0 critical sessions (nothing to save; previous backup kept)` (wording flexible).
- Soft fail: `warning: …` + exit 0 with `--once`.
- Help: documents auto-backup, `--interval` default 10m, `--once`, default path;
  `--dry-run` as one-cycle plan exit.
- Parent `sessions -h` lists `auto-backup`.

## Decision Tree

```
sessions-auto-backup/
├── help/
│   ├── auto-backup-usage/   auto-backup -h: 10m, --once, sessions-auto.json, core flags
│   └── sessions-lists/      sessions -h lists auto-backup
├── once/                    always --once (no real sleep)
│   ├── write/               critical fixture → writes version + sessions; exit 0
│   ├── overwrite-pending/   seed pending auto file → always overwrite (non-TTY); exit 0
│   ├── zero/
│   │   ├── no-file/         idle → 0 critical; no file created
│   │   └── keep-existing/   idle + seed → previous file kept (no clobber)
│   ├── capture-fail/
│   │   ├── soft-exit/       FailSnapshotCapture → warning + exit 0; no file
│   │   └── no-clobber/      fail + seed → previous file kept
│   ├── dry-run/             --dry-run alone: plan; no write; exits (no loop)
│   └── custom-file/         --file PATH used in write + stdout path line
└── validation/
    └── bad-interval/        invalid --interval → Error + non-zero (before loop)
```

## Test Index

| Leaf | Description | Expect |
|------|-------------|--------|
| `help/auto-backup-usage/` | auto-backup -h: interval 10m, --once, sessions-auto.json, --file, --dry-run | RED |
| `help/sessions-lists/` | sessions -h mentions auto-backup | RED |
| `once/write/` | critical → Saved; version=1; grok+mark; exit 0 | RED |
| `once/overwrite-pending/` | pending seed → overwrite OK non-TTY; new sessions not old-sess | RED |
| `once/zero/no-file/` | idle → 0 critical; no file | RED |
| `once/zero/keep-existing/` | idle + seed → old content kept | RED |
| `once/capture-fail/soft-exit/` | capture fail → warning; exit 0; no file | RED |
| `once/capture-fail/no-clobber/` | capture fail + seed → old kept | RED |
| `once/dry-run/` | --dry-run without --once: Would save; no file; exits | GREEN |
| `once/custom-file/` | --file absolute path written + mentioned | RED |
| `validation/bad-interval/` | bad --interval → Error; non-zero | RED |

## How to Run

```sh
doctest vet ./tests/iterm2/sessions-auto-backup
doctest test ./tests/iterm2/sessions-auto-backup
doctest test ./tests/iterm2/sessions-auto-backup/help
doctest test ./tests/iterm2/sessions-auto-backup/once
doctest test ./tests/iterm2/sessions-auto-backup/validation
```

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	iterm2 "github.com/xhd2015/kool/tools/iterm2"
)

const (
	ModeHelp       = "help"
	ModeAutoBackup = "auto-backup"
)

const (
	fixtureGrokSessionID = "019fabcdef-1234-5678-9abc-def012345678"
	fixtureMarkMessage   = "still waiting for CI"
)

// Request drives L2 in-process auto-backup via iterm2.RunForTest.
// Live write leaves set Once=true (no multi-cycle sleep). DryRun alone also
// exits after one cycle without Once (product: --dry-run implies one-shot).
type Request struct {
	Mode string

	HelpArgs []string

	// Once maps to --once (required for L2 live write leaves).
	Once bool
	// DryRun maps to --dry-run (plan only; no write; always one cycle then exit).
	DryRun bool
	// FilePath is absolute or relative to WorkingDir; empty omits --file
	// (product default ~/.config/iterm2/sessions-auto.json — avoid in write
	// leaves; assert default via help text).
	FilePath string
	// Interval maps to --interval DUR (empty = omit flag → product default 10m).
	Interval string
	ExtraArgs []string

	Color   bool
	NoColor bool

	IgnoreMacOSSpace bool
	Spaces           string

	// Install critical fixture (one window: grok + mark + idle).
	UseCriticalFixture bool
	// Idle-only fixture (zero critical).
	UseIdleOnlyFixture bool
	// FailSnapshotCapture: iTerm not running / capture fails soft.
	FailSnapshotCapture bool

	// Pre-seed checkpoint at FilePath before run.
	SeedDoc     *iterm2.SaveDocument
	SeedRawJSON string

	WorkingDir string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	FileJSON string
	Doc      *iterm2.SaveDocument
	// ResolvedPath is the absolute path used for --file / seed / readback.
	ResolvedPath string
}

func criticalWindows() []iterm2.SnapshotWindow {
	return []iterm2.SnapshotWindow{
		{
			Index: 1,
			Name:  "Win-Crit",
			Tabs: []iterm2.SnapshotTab{
				{Index: 1, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "AAAAAAAA-0000-0000-0000-000000000001", TTY: "/dev/ttys001", Profile: "Default"},
				}},
				{Index: 2, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "BBBBBBBB-0000-0000-0000-000000000002", TTY: "/dev/ttys002", Profile: "Default"},
				}},
				{Index: 3, Sessions: []iterm2.SnapshotSession{
					{Index: 1, ID: "CCCCCCCC-0000-0000-0000-000000000003", TTY: "/dev/ttys003", Profile: "Default"},
				}},
			},
		},
	}
}

func installCritical(t *testing.T) {
	t.Helper()
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows:      criticalWindows(),
		BusyTTYs:     []string{"ttys001", "ttys002"},
		IdleTTYs:     []string{"ttys003"},
		BusyLeafByTTY: map[string]string{
			"ttys002": "mark still waiting for CI",
		},
		CwdByTTY: map[string]string{
			"ttys001": "/proj/grok",
			"ttys002": "/proj/mark",
		},
		AgentResolveByTTY: map[string]iterm2.AgentResolveFixture{
			"ttys001": {Kind: "grok", SessionID: fixtureGrokSessionID, Title: "fixture-title"},
		},
		Now:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		Hostname: "testhost",
	})
}

func installIdleOnly(t *testing.T) {
	t.Helper()
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: true,
		Windows: []iterm2.SnapshotWindow{{
			Index: 1,
			Tabs: []iterm2.SnapshotTab{{Index: 1, Sessions: []iterm2.SnapshotSession{
				{Index: 1, ID: "AAAAAAAA-0000-0000-0000-000000000001", TTY: "/dev/ttys001", Profile: "Default"},
			}}},
		}},
		IdleTTYs: []string{"ttys001"},
		Hostname: "testhost",
	})
}

func installFailCapture(t *testing.T) {
	t.Helper()
	iterm2.InstallPhasedFixtureCollectorForTest(t, iterm2.PhasedFixtureOpts{
		ITermRunning: false,
		Windows:      nil,
		Hostname:     "testhost",
	})
}

// pendingSeedDoc is an unconsumed (restored_at null) checkpoint with old-sess.
func pendingSeedDoc() *iterm2.SaveDocument {
	return &iterm2.SaveDocument{
		Version: 1,
		SavedAt: "2026-07-01T00:00:00+0800",
		Host:    "old",
		Source:  "kool-iterm2-sessions-save",
		Summary: iterm2.SaveSummary{Windows: 1, Tabs: 1, Sessions: 1, ByKind: map[string]int{"grok": 1}},
		Windows: []iterm2.SaveWindow{{
			SourceIndex: 1,
			Tabs: []iterm2.SaveTab{{
				Cwd: "/old", Kind: "grok", SessionID: "old-sess", ResumeCmd: "grok --resume old-sess",
			}},
		}},
	}
}

func resolveFile(req *Request) string {
	p := req.FilePath
	if p == "" {
		return ""
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(req.WorkingDir, p)
	}
	return p
}

func seedCheckpoint(path string, req *Request) error {
	if path == "" {
		return nil
	}
	if req.SeedRawJSON != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(req.SeedRawJSON), 0o644)
	}
	if req.SeedDoc != nil {
		return iterm2.WriteSaveDocument(path, req.SeedDoc)
	}
	return nil
}

func appendColorFlags(args []string, req *Request) []string {
	if req.Color {
		args = append(args, "--color")
	}
	if req.NoColor {
		args = append(args, "--no-color")
	}
	return args
}

func appendSpaceFlags(args []string, req *Request) []string {
	if req.IgnoreMacOSSpace {
		args = append(args, "--ignore-macos-space")
	}
	if req.Spaces != "" {
		args = append(args, "--spaces", req.Spaces)
	}
	return args
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	path := resolveFile(req)

	if err := seedCheckpoint(path, req); err != nil {
		return nil, err
	}

	switch req.Mode {
	case ModeHelp:
		var stdout, stderr bytes.Buffer
		args := req.HelpArgs
		if len(args) == 0 {
			args = []string{"sessions", "auto-backup", "-h"}
		}
		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		return &Response{
			Stdout:       stdout.String(),
			Stderr:       stderr.String(),
			ExitCode:     code,
			ResolvedPath: path,
		}, nil

	case ModeAutoBackup:
		var stdout, stderr bytes.Buffer

		if req.FailSnapshotCapture {
			installFailCapture(t)
		} else if req.UseIdleOnlyFixture {
			installIdleOnly(t)
		} else if req.UseCriticalFixture {
			installCritical(t)
		}

		args := []string{"sessions", "auto-backup"}
		if req.Once {
			args = append(args, "--once")
		}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		if path != "" {
			args = append(args, "--file", path)
		}
		if req.Interval != "" {
			args = append(args, "--interval", req.Interval)
		}
		args = appendColorFlags(args, req)
		args = appendSpaceFlags(args, req)
		args = append(args, req.ExtraArgs...)

		code := iterm2.RunForTest(args, &stdout, &stderr, req.WorkingDir)
		resp := &Response{
			Stdout:       stdout.String(),
			Stderr:       stderr.String(),
			ExitCode:     code,
			ResolvedPath: path,
		}
		if path != "" {
			if b, err := os.ReadFile(path); err == nil {
				resp.FileJSON = string(b)
				var doc iterm2.SaveDocument
				if json.Unmarshal(b, &doc) == nil {
					resp.Doc = &doc
				}
			}
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}
```
