# kool sandbox (P1 package + P2 sealed-binary run)

`kool sandbox build` packs sandbox files and env into a **cross-compiled sealed
binary**. One-time RSA keypair per build; bulk data under AES-256-GCM; DEK
wrapped with RSA-OAEP; private key + ciphertext embedded via `//go:embed`.
P1 asserts the **built artifact** and security bar. P2 asserts **unpack +
materialize + exec** of the sealed binary on the **host** GOOS/GOARCH (classic
TDD: run leaves are RED until the runner is implemented).

## Version

0.0.5

## DSN (Domain Specific Notion)

### Participants

- **User** — invokes `kool sandbox` help, `build`, or `inspect` with flags and
  paths; later executes the sealed binary with a guest command.
- **kool CLI router** — `main.go` dispatches `case "sandbox"` to
  `tools/sandbox`.
- **sandbox handler** — root/build help; `build` flag parse; merge input dir +
  flags; validate; seal; cross-compile write output binary; `inspect` reads a
  sealed binary without executing the sandbox.
- **Input sources** — config directory (`-i`: `meta.yaml`, `files/`, `env.yaml`)
  and/or repeatable `--file LOCAL=SANDBOX_REL` and `--env KEY=VALUE`. Flags win
  on path/env-key conflict. Optional `--home-linked` seals a home-linked pack.
  Optional repeatable `--runtime-load-devbox ABS` seals absolute remote paths of
  other sandbox binaries (strings only; pack does not open them).
- **Sealer** — per-build RSA keypair; AES-256-GCM for PackBlob; RSA-OAEP wrap of
  DEK; embed sealed blob in runner binary.
- **Output binary (sealed runner)** — path from `-o` / `--output`; target OS/arch
  from `--goos` / `--goarch` (default host runtime). At run time it unseals the
  payload, materializes files (optionally seed+overlay real home), merges optional
  load-devbox packs (files + env), applies env, execs the guest command.
- **Load-devbox target** — another host-built sealed sandbox binary (absolute
  path). Unsealed for Files + Env only; nested sealed `RuntimeLoadDevbox` paths
  are walked DFS with cycle skip on already-seen abs paths.
- **Materialize root** — session directory under Linux default
  `/dev/shm/kool-sandbox/<id>/`, or under env `KOOL_SANDBOX_ROOT` (required for
  macOS/doctest hosts), with temp-dir fallback. Child process `cwd` and env
  `SANDBOX_ROOT` are this absolute path. Removed best-effort when the process
  exits. When home-linked, guest `HOME` is also this absolute path.
- **Real home (home-linked)** — host `HOME` / `UserHomeDir` captured once before
  sandbox override; top-level entries seeded as symlinks into the materialize
  root, then packed files overlay (explode intermediate symlinks as needed).
- **Inspector** — `kool sandbox inspect <binary>`: name, file paths + hashes,
  env keys only; sealed `runtime-load-devbox` absolute paths when present.

### Behaviors

- **Root help** — `kool sandbox -h|--help` lists `build` (and `inspect` when
  present) and principal flags; exit 0; stdout ends with `\n`.
- **Build help** — `kool sandbox build -h|--help` documents `-o/-i/--file/--env/
  --goos/--goarch/--home-linked/--runtime-load-devbox`; exit 0; stdout ends
  with `\n`.
- **Build validation** — missing `-o`, empty pack (no files and no env after
  merge), `--env` without `=`, missing local path for `--file` → non-zero;
  message on stderr; no output binary required. Relative
  `--runtime-load-devbox` → non-zero; stderr `Error:` (absolute / relative /
  runtime-load-devbox). Load-devbox paths alone do **not** fill an empty pack.
- **Build from input dir** — `-i` with `files/` and/or `env.yaml` (+ optional
  `meta.yaml`) → exit 0; binary at `-o` with size > 0; stdout mentions sandbox
  name (or default), files/env counts, size; ends with `\n`.
- **Build from flags only** — no `-i`; only `--file` / `--env` → success when
  pack non-empty.
- **Build home-linked** — `--home-linked` accepted on build; bit sealed into
  PackBlob (`home_linked: true`); binary still builds when pack non-empty.
- **Build runtime-load-devbox** — repeatable `--runtime-load-devbox ABS` seals
  absolute path strings into PackBlob `runtime_load_devbox` (no local FS open);
  non-empty pack still requires ≥1 file or env; `inspect` lists sealed paths.
- **Merge** — dir + flags; flag value wins same relative path or env key;
  `inspect` shows winning paths/keys (not env secret values).
- **Cross-compile** — `--goos linux --goarch amd64` produces a binary; when
  `file` is available, optional ELF check.
- **Security bar** — unique secret string in packed file content must not appear
  as plaintext in `strings` of the output binary; two builds of the same input
  must not produce byte-identical binaries (fresh key/ciphertext per build).
- **Empty pack rejected** — neither files nor env after merge → error.
- **Sealed run (P2)** — `./sandbox.bin [runner-flags] [--] <command> [args…]`:
  unseal → materialize under session root (honor `KOOL_SANDBOX_ROOT`) → write
  files with modes → apply packed env → set `SANDBOX_ROOT` → `cwd =
  SANDBOX_ROOT` → exec command (inherit stdio) → exit code = child exit →
  remove materialize dir on exit. Missing command → non-zero, stderr `Error:`
  style (mentions command/usage). No ssh-agent special case. Runner flags use
  less-flags **`StopOnFirstArg()`** so guest argv may follow without a `--`.
- **Sealed run home-linked** — when PackBlob `home_linked` is true: capture real
  home from process `HOME` before override; seed top-level names from real home
  as absolute symlinks into session root; overlay packed files (explode
  intermediate symlink components; leaf replace; mid-path type conflict
  replace); set guest `HOME = SANDBOX_ROOT` (critical); then apply other packed
  env. Packed env `HOME` equal to abs `SANDBOX_ROOT` → notice on stderr,
  continue; any other packed `HOME` → error, non-zero (mentions HOME /
  home-linked). When `home_linked` is false: no HOME force, no seed (unchanged).
  Same HOME policy applies to env keys contributed by load-devbox packs.
- **Sealed run load-devbox** — repeatable **`--load-devbox ABS`** at runtime
  (StringSlice). Path list = sealed `RuntimeLoadDevbox` (pack order) then adhoc
  CLI; **first-seen dedupe**. Each path must be absolute and a sealed sandbox
  binary; unseal Files + Env; recurse nested sealed `RuntimeLoadDevbox` (DFS);
  cycle → skip already-seen abs paths. File layers via `linkoverlay.Apply`
  (later wins): optional real HOME dir if primary home-linked → primary pack
  Files → each loaded box Files in walk order. Env merge: union of primary +
  each load; **same key from any two sandboxes → hard error** (even same value);
  labels `current sandbox` vs abs load path; host env not in conflict set.
  Runner injects `SANDBOX_ROOT` / `PWD` / (`HOME` if home-linked) after merge.
  Per successful load: stdout `notice: loading devbox <abs>\n` (grey ANSI only
  when stdout is TTY; doctest capture is non-TTY → no ANSI). Relative path,
  missing path, not sealed, or unseal fail → hard `Error:` non-zero. Nested
  adhoc flags from secondary binaries are not applied.

### Pack / seal model (conceptual)

```text
PackBlob: Version, Name, CreatedAt, ExpiresAt?, HomeLinked?,
          RuntimeLoadDevbox?, Files[{Path,Mode,Content}], Env
Sealed: RSA private (one-time) + RSA-OAEP(AES-256 DEK) + AES-GCM(PackBlob)
```

### Inspect CLI (P1 helper surface)

```text
kool sandbox inspect <binary>
  -> exit 0; stdout lists name, file paths (+ content hashes), env keys only,
     and sealed runtime-load-devbox absolute paths when present
```

### Sealed binary CLI (P2)

```text
./sandbox.bin [--load-devbox ABS]... [--] <command> [args...]
  env KOOL_SANDBOX_ROOT=<parent>   # force materialize parent (tests / macOS)
  env HOME=<fake-real-home>        # home-linked tests: inject "real" home
  -> unseal primary; merge sealed+adhoc load-devbox Files/Env;
     materialize under <parent>/<session>/; exec command; cleanup
```

## Decision Tree

```
sandbox/
├── help/                                   [usage; exit 0; no build]
│   ├── root/                               kool sandbox --help
│   └── build/                              kool sandbox build --help (+ --home-linked, --runtime-load-devbox)
├── build/                                  [package sealed binary]
│   ├── validation/                         [errors before artifact]
│   │   ├── missing-output/                 no -o
│   │   ├── empty-pack/                     no -i content and no flags
│   │   ├── bad-env-flag/                   --env without =
│   │   └── missing-file-source/            --file local path missing
│   ├── from-input-dir/                     [-i with layout]
│   │   ├── files-and-env/                  files/ + env.yaml → binary + counts
│   │   └── meta-name/                      meta.yaml name on stdout
│   ├── from-flags/                         [no -i]
│   │   └── file-and-env-flags-only/        --file + --env only
│   ├── home-linked/                        [--home-linked pack bit]
│   │   └── packs-ok/                       --home-linked + file/env → binary
│   ├── runtime-load-devbox/                [--runtime-load-devbox seal]
│   │   ├── packs-ok/                       abs paths sealed; inspect lists them
│   │   └── relative-rejected/              relative path → non-zero Error:
│   ├── merge/                              [dir + flags; flags win]
│   │   └── flag-overrides-dir/             build + inspect winning keys/paths
│   ├── cross-compile/                      [--goos/--goarch]
│   │   └── linux-amd64/                    linux/amd64 binary exists
│   └── security-bar/                       [crypto / no leak]
│       ├── no-plaintext-secret/            secret not in strings(binary)
│       └── two-builds-differ/              same input → different binaries
└── run/                                    [P2: unpack + materialize + exec]
    ├── validation/                         [runner arg errors]
    │   └── missing-command/                no command → non-zero; Error: usage
    ├── happy/                              [materialize + guest sees pack]
    │   ├── pwd-is-sandbox-root/            pwd under KOOL_SANDBOX_ROOT session
    │   ├── file-visible/                   cat packed relative path
    │   ├── env-visible/                    packed --env visible to child
    │   ├── sandbox-root-env/               $SANDBOX_ROOT == cwd abs path
    │   └── relative-path-from-cwd/         cat nested relative path works
    ├── home-linked/                        [seed real home + overlay pack]
    │   ├── happy/                          [HOME + seed + explode]
    │   │   ├── home-equals-sandbox-root/   $HOME == $SANDBOX_ROOT == session
    │   │   ├── top-level-seed-visible/     real-home seed.txt via symlink
    │   │   ├── direct-child-replace/       packed leaf replaces seed link
    │   │   └── explode-deep-path/          explode .config; sibling re-link
    │   └── env-policy/                     [packed HOME vs home-linked]
    │       └── packed-home-conflict/       packed HOME≠sandbox → non-zero
    ├── cleanup/                            [session dir lifecycle]
    │   └── removes-materialize-dir/        parent empty after successful run
    ├── exit-code/                          [propagate child status]
    │   └── child-nonzero/                  sh -c 'exit 42' → exit 42
    └── load-devbox/                        [P2: --load-devbox runtime merge]
        ├── happy/                          [files + notices + layer order]
        │   ├── adhoc-only/                 CLI --load-devbox; guest sees secondary file
        │   ├── sealed-only/                pack --runtime-load-devbox only
        │   ├── sealed-and-adhoc/           both sources; notices ×2; both files
        │   ├── dedupe/                     same path sealed+CLI → one notice
        │   ├── later-wins-over-pack/       same relpath → load content wins
        │   ├── with-home-linked/           HOME seed < pack < load layer order
        │   ├── stop-on-first-arg/          no -- before guest still works
        │   └── nested-sealed/              B seals A; primary loads B → A files
        ├── env/                            [env merge / conflict]
        │   ├── merge-disjoint/             primary A=1 load B=2 both visible
        │   ├── conflict-with-primary/      same key primary+load → Error
        │   └── conflict-between-loads/     two loads same key → Error
        ├── notice/                         [stdout notice contract]
        │   └── on-stdout/                  notice: loading devbox <abs>; no ANSI
        └── validation/                     [runtime path errors]
            ├── missing-path/               abs path missing → Error
            ├── not-sealed/                 plain file not sealed → Error
            └── relative-path/              relative --load-devbox → Error
```

## Test Index

| Leaf | Description |
|------|-------------|
| `help/root/` | Root `--help` exit 0; mentions build and key flags; trailing `\n` |
| `help/build/` | Build `--help` exit 0; documents `-o/-i/--file/--env/--goos/--goarch/--home-linked/--runtime-load-devbox` |
| `build/validation/missing-output/` | Build without `-o` → non-zero; stderr mentions output/`-o` |
| `build/validation/empty-pack/` | No input content → non-zero; empty pack rejected |
| `build/validation/bad-env-flag/` | `--env NOTVALID` → non-zero; message mentions env/`=` |
| `build/validation/missing-file-source/` | `--file` local path missing → non-zero |
| `build/from-input-dir/files-and-env/` | Fixture dir → exit 0; binary size > 0; stdout files/env counts |
| `build/from-input-dir/meta-name/` | `meta.yaml` name appears in stdout |
| `build/from-flags/file-and-env-flags-only/` | Flags only → success binary |
| `build/home-linked/packs-ok/` | `--home-linked` + file/env → exit 0; binary size > 0 |
| `build/runtime-load-devbox/packs-ok/` | `--runtime-load-devbox` abs paths sealed; inspect lists them |
| `build/runtime-load-devbox/relative-rejected/` | Relative `--runtime-load-devbox` → non-zero; stderr `Error:` |
| `build/merge/flag-overrides-dir/` | Flag wins path/env; inspect shows winning path + env key |
| `build/cross-compile/linux-amd64/` | `--goos linux --goarch amd64` → binary; optional ELF |
| `build/security-bar/no-plaintext-secret/` | Unique secret not in `strings` of sealed binary |
| `build/security-bar/two-builds-differ/` | Two builds same input → binaries not byte-identical |
| `run/validation/missing-command/` | Sealed bin with no args → non-zero; stderr command/usage |
| `run/happy/pwd-is-sandbox-root/` | `pwd` is session materialize abs path under `KOOL_SANDBOX_ROOT` |
| `run/happy/file-visible/` | Packed `hello.txt` content via `sh -c 'cat hello.txt'` |
| `run/happy/env-visible/` | Packed `FOO=bar` visible in child env |
| `run/happy/sandbox-root-env/` | `$SANDBOX_ROOT` equals materialize cwd abs path |
| `run/happy/relative-path-from-cwd/` | Nested packed file readable via relative path |
| `run/home-linked/happy/home-equals-sandbox-root/` | Guest `$HOME` equals `$SANDBOX_ROOT` (session root; not fake real home) |
| `run/home-linked/happy/top-level-seed-visible/` | Top-level seed from fake real home visible via symlink |
| `run/home-linked/happy/direct-child-replace/` | Packed leaf replaces same-name seed symlink |
| `run/home-linked/happy/explode-deep-path/` | Explode intermediate symlink; packed path + sibling seed both work |
| `run/home-linked/env-policy/packed-home-conflict/` | Packed `HOME` other than sandbox root → non-zero; stderr HOME/home-linked |
| `run/cleanup/removes-materialize-dir/` | After exit 0, no session children under materialize parent |
| `run/exit-code/child-nonzero/` | Guest `exit 42` → sealed binary exit code 42 |
| `run/load-devbox/happy/adhoc-only/` | CLI `--load-devbox` secondary; guest sees secondary file + notice |
| `run/load-devbox/happy/sealed-only/` | Pack `--runtime-load-devbox` only; guest sees load file + notice |
| `run/load-devbox/happy/sealed-and-adhoc/` | Sealed + adhoc distinct paths; two notices; both files |
| `run/load-devbox/happy/dedupe/` | Same abs path sealed+CLI → single notice |
| `run/load-devbox/happy/later-wins-over-pack/` | Same relpath primary vs load → load content |
| `run/load-devbox/happy/with-home-linked/` | Home seed < pack < load; guest sees load; HOME==SANDBOX_ROOT |
| `run/load-devbox/happy/stop-on-first-arg/` | `--load-devbox ABS sh -c …` without `--` works |
| `run/load-devbox/happy/nested-sealed/` | Primary loads B; B seals A; guest sees A's file |
| `run/load-devbox/env/merge-disjoint/` | Disjoint env keys from primary + load both visible |
| `run/load-devbox/env/conflict-with-primary/` | Same env key primary+load → non-zero Error incompatible |
| `run/load-devbox/env/conflict-between-loads/` | Two loads same env key → non-zero Error |
| `run/load-devbox/notice/on-stdout/` | `notice: loading devbox <abs>` on RunStdout; no ANSI |
| `run/load-devbox/validation/missing-path/` | Missing abs load path → non-zero Error |
| `run/load-devbox/validation/not-sealed/` | Non-sealed file at abs path → non-zero Error |
| `run/load-devbox/validation/relative-path/` | Relative `--load-devbox` → non-zero Error |

## How to Run

```sh
doctest vet ./tests/sandbox
doctest test ./tests/sandbox
```

Classic TDD: P1 pack/inspect (including `build/runtime-load-devbox/*`, help) is
**GREEN**. Baseline P2 sealed run + home-linked are **GREEN** when unseal/
materialize/exec land. New **`run/load-devbox/*`** leaves are **RED** until the
sealed runner implements runtime `--load-devbox` merge (files, env conflict,
notices, validation). `Run` session-builds `kool` from the module root so leaves
exercise the workspace binary. Run leaves build for **host GOOS/GOARCH** (no
`--goos linux`) so the sealed binary can execute on the doctest machine, and
force materialize via `KOOL_SANDBOX_ROOT`. Home-linked sealed runs also inject a
**fake real home** via child `HOME` (`Request.SealedHome` / `SealedEnv`) — no
process-global `Setenv`. Load-devbox leaves pre-build **secondary** sealed
binaries under `WorkingDir` via session `kool` (`SecondaryPacks`) and pass
`--load-devbox` via `SealedLoadDevbox`.

```go
import (
	"runtime"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

// SecondaryPack describes an extra host-built sealed sandbox used as a
// --load-devbox / --runtime-load-devbox target. Built under WorkingDir with
// session kool before the primary sealed run (host GOOS/GOARCH).
type SecondaryPack struct {
	// Output is -o path (absolute or WorkingDir-relative), e.g. "load-a.bin".
	Output string
	// ExtraFiles are --file LOCAL=SANDBOX_REL entries (paths relative to WorkingDir).
	ExtraFiles []string
	// ExtraEnv are --env KEY=VALUE entries.
	ExtraEnv []string
	// HomeLinked passes --home-linked on the secondary build.
	HomeLinked bool
	// RuntimeLoadDevbox seals nested load paths into the secondary PackBlob.
	RuntimeLoadDevbox []string
}

// Request drives one or more kool sandbox invocations for a leaf, and optionally
// a follow-up sealed-binary run (P2).
type Request struct {
	// HelpAtRoot runs `kool sandbox --help` (ignores other command fields).
	HelpAtRoot bool
	// HelpBuild runs `kool sandbox build --help`.
	HelpBuild bool

	// Subcommand is the first positional after "sandbox" when not help
	// (typically "build"). Empty with neither help flag = bare `kool sandbox`.
	Subcommand string

	// Build flags.
	Output    string // -o path (absolute or relative to WorkingDir)
	OutputSet bool   // pass -o even when Output == ""
	Input     string // -i path
	InputSet  bool
	// ExtraFiles are --file LOCAL=SANDBOX_REL entries (repeatable).
	ExtraFiles []string
	// ExtraEnv are --env KEY=VALUE entries (repeatable). May be invalid for validation leaves.
	ExtraEnv []string
	Goos     string
	Goarch   string
	// HomeLinked passes --home-linked on `kool sandbox build` (sealed into PackBlob).
	HomeLinked bool
	// RuntimeLoadDevbox are --runtime-load-devbox PATH entries (repeatable).
	// Values are passed to the CLI as-is (absolute for seal-ok leaves; relative
	// for validation). Pack seals strings only — paths need not exist locally.
	RuntimeLoadDevbox []string

	// BuildTwice: run build twice with same inputs to two different -o paths
	// (security-bar/two-builds-differ).
	BuildTwice bool

	// AfterBuildInspect: on first build exit 0, run `kool sandbox inspect <Output>`.
	AfterBuildInspect bool

	// AfterBuildRun: on first build exit 0 + binary exists, execute the sealed
	// binary under WorkingDir with KOOL_SANDBOX_ROOT (P2 run leaves).
	AfterBuildRun bool
	// SealedArgs is the guest argv passed to the sealed binary (command + args).
	// Empty means invoke the binary with no guest args (missing-command).
	// Runner flags such as --load-devbox are not placed here — use SealedLoadDevbox.
	SealedArgs []string
	// SealedLoadDevbox are absolute (or as-is relative for validation) paths
	// passed as repeatable `--load-devbox` before SealedArgs on the sealed run.
	SealedLoadDevbox []string
	// SealedDoubleDash inserts `--` after runner flags (SealedLoadDevbox) and
	// before SealedArgs (ends runner flags). StopOnFirstArg leaves set false.
	SealedDoubleDash bool
	// SecondaryPacks are host-built sealed sandboxes for load-devbox targets.
	// Built before the primary sealed run; absolute -o paths recorded on Response.
	SecondaryPacks []SecondaryPack
	// SandboxRootParent is the absolute (or WorkingDir-relative) path exported as
	// KOOL_SANDBOX_ROOT for the sealed process. Empty → WorkingDir/kool-sandbox-root.
	SandboxRootParent string
	// SealedHome, when non-empty, is injected as HOME for the sealed process
	// (fake real home for home-linked leaves). Absolute or WorkingDir-relative.
	// No process-global Setenv — child cmd.Env only.
	SealedHome string
	// SealedEnv are additional KEY=VALUE pairs for the sealed process env
	// (appended after KOOL_SANDBOX_ROOT and optional HOME from SealedHome).
	SealedEnv []string

	// WorkingDir is the kool process cwd (isolation). Set by root Setup.
	WorkingDir string

	// ProcessTimeout bounds each kool / sealed subprocess wall clock (default 3m for build).
	ProcessTimeout time.Duration

	// SecretProbe is the unique string Assert checks must not appear in binary
	// strings (security-bar/no-plaintext-secret). Set by leaf Setup when packing.
	SecretProbe string
}

// Response is CLI capture after Run.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int

	// OutputPath is the resolved first -o path when set.
	OutputPath   string
	OutputExists bool
	OutputSize   int64

	// Second build (BuildTwice).
	SecondOutputPath   string
	SecondOutputExists bool
	SecondOutputSize   int64
	BinariesIdentical  bool // true if both exist and SHA-256 digests match

	// Inspect capture (AfterBuildInspect).
	InspectStdout   string
	InspectStderr   string
	InspectExitCode int
	InspectRan      bool

	// Sealed-binary run capture (AfterBuildRun).
	RunExecuted          bool
	RunStdout            string
	RunStderr            string
	RunExitCode          int
	SandboxRootParent    string   // absolute KOOL_SANDBOX_ROOT used
	SealedHome           string   // absolute HOME injected into sealed process (if any)
	MaterializeRemaining []string // entry names still under parent after process exit
	MaterializeEmpty     bool     // true when parent has no remaining children
	// SecondaryPaths are absolute -o paths of SecondaryPacks, same order as request.
	SecondaryPaths []string
}

// moduleRoot returns the kool module root from the doctest tree root path.
func moduleRoot(doctestRoot string) string {
	return filepath.Clean(filepath.Join(doctestRoot, "..", ".."))
}

// sessionCacheDir is keyed by the doctest session id.
func sessionCacheDir(sessionID string) string {
	return filepath.Join(os.TempDir(), "kool-sandbox-doctest-"+sessionID)
}

func withFileLock(t *testing.T, lockPath string, fn func() error) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

// ensureKoolBinary builds kool once per doctest session into the session cache.
// One-process mode: use d.DOCTEST_ROOT / d.DOCTEST_SESSION_ID (no bare free ids).
func ensureKoolBinary(t *testing.T, d *session.Doctest) (string, error) {
	t.Helper()
	cacheDir := sessionCacheDir(d.DOCTEST_SESSION_ID)
	lock := filepath.Join(cacheDir, "build.lock")
	ready := filepath.Join(cacheDir, "binaries.ready")
	bin := filepath.Join(cacheDir, "kool")
	modRoot := moduleRoot(d.DOCTEST_ROOT)
	err := withFileLock(t, lock, func() error {
		if st, err := os.Stat(ready); err == nil && !st.IsDir() {
			if st2, err2 := os.Stat(bin); err2 == nil && !st2.IsDir() {
				return nil
			}
		}
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return err
		}
		cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", bin, ".")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("go build kool: %w\n%s", err, out)
		}
		return os.WriteFile(ready, []byte("ok\n"), 0644)
	})
	if err != nil {
		return "", err
	}
	return bin, nil
}

func resolvePath(workingDir, p string) string {
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(workingDir, p)
}

func buildSandboxArgs(req *Request, outputOverride string) []string {
	if req.HelpAtRoot {
		return []string{"sandbox", "--help"}
	}
	if req.HelpBuild {
		return []string{"sandbox", "build", "--help"}
	}
	args := []string{"sandbox"}
	if req.Subcommand != "" {
		args = append(args, req.Subcommand)
	}
	if req.Subcommand == "build" {
		out := req.Output
		if outputOverride != "" {
			out = outputOverride
		}
		if req.OutputSet || out != "" {
			args = append(args, "-o", out)
		}
		if req.InputSet || req.Input != "" {
			args = append(args, "-i", req.Input)
		}
		for _, f := range req.ExtraFiles {
			args = append(args, "--file", f)
		}
		for _, e := range req.ExtraEnv {
			args = append(args, "--env", e)
		}
		if req.Goos != "" {
			args = append(args, "--goos", req.Goos)
		}
		if req.Goarch != "" {
			args = append(args, "--goarch", req.Goarch)
		}
		if req.HomeLinked {
			args = append(args, "--home-linked")
		}
		for _, p := range req.RuntimeLoadDevbox {
			args = append(args, "--runtime-load-devbox", p)
		}
	}
	return args
}

// buildSecondaryArgs builds `kool sandbox build` argv for a SecondaryPack.
// Host GOOS/GOARCH (no --goos/--goarch) so the binary is runnable/unsealable here.
func buildSecondaryArgs(sp SecondaryPack) []string {
	args := []string{"sandbox", "build", "-o", sp.Output}
	for _, f := range sp.ExtraFiles {
		args = append(args, "--file", f)
	}
	for _, e := range sp.ExtraEnv {
		args = append(args, "--env", e)
	}
	if sp.HomeLinked {
		args = append(args, "--home-linked")
	}
	for _, p := range sp.RuntimeLoadDevbox {
		args = append(args, "--runtime-load-devbox", p)
	}
	return args
}

// buildSecondaryPacks builds each SecondaryPack under WorkingDir with session kool.
// Returns absolute output paths in request order. Fails if any build is non-zero
// or the artifact is missing.
func buildSecondaryPacks(t *testing.T, koolBin string, workingDir string, timeout time.Duration, packs []SecondaryPack) ([]string, error) {
	t.Helper()
	paths := make([]string, 0, len(packs))
	for i, sp := range packs {
		if sp.Output == "" {
			return nil, fmt.Errorf("SecondaryPacks[%d]: empty Output", i)
		}
		args := buildSecondaryArgs(sp)
		stdout, stderr, code, runErr := runKool(t, koolBin, workingDir, timeout, args)
		if runErr != nil {
			return nil, fmt.Errorf("SecondaryPacks[%d] build: %w", i, runErr)
		}
		if code != 0 {
			return nil, fmt.Errorf("SecondaryPacks[%d] build exit=%d stdout=%q stderr=%q", i, code, stdout, stderr)
		}
		outPath := resolvePath(workingDir, sp.Output)
		exists, size := statOutput(outPath)
		if !exists || size <= 0 {
			return nil, fmt.Errorf("SecondaryPacks[%d]: missing binary at %q", i, outPath)
		}
		paths = append(paths, outPath)
	}
	return paths, nil
}

// sealedRunArgs composes sealed-binary argv: --load-devbox… [ -- ] guest…
func sealedRunArgs(req *Request) []string {
	var args []string
	for _, p := range req.SealedLoadDevbox {
		args = append(args, "--load-devbox", p)
	}
	if req.SealedDoubleDash {
		args = append(args, "--")
	}
	args = append(args, req.SealedArgs...)
	return args
}

func runKool(t *testing.T, koolBin string, workingDir string, timeout time.Duration, args []string) (stdout, stderr string, exitCode int, runErr error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 3 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, koolBin, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, -1, fmt.Errorf("kool sandbox exceeded process timeout %v; stderr=%q", timeout, stderr)
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, stderr, exitErr.ExitCode(), nil
		}
		return stdout, stderr, -1, fmt.Errorf("run kool: %w", err)
	}
	return stdout, stderr, 0, nil
}

// runSealedBinary executes a host-built sealed sandbox binary with KOOL_SANDBOX_ROOT
// set. Optional sealedHome (absolute) is injected as HOME for home-linked tests;
// sealedEnv are extra KEY=VALUE pairs. No process-global env mutation.
func runSealedBinary(t *testing.T, binPath, workingDir, sandboxRootParent, sealedHome string, sealedEnv []string, timeout time.Duration, args []string) (stdout, stderr string, exitCode int, runErr error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 1 * time.Minute
	}
	if err := os.MkdirAll(sandboxRootParent, 0755); err != nil {
		return "", "", -1, fmt.Errorf("mkdir KOOL_SANDBOX_ROOT: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	// Force materialize parent for cleanup asserts and macOS hosts.
	// Optional HOME=fake-real-home so home-linked runtime captures injected home.
	env := append(os.Environ(), "KOOL_SANDBOX_ROOT="+sandboxRootParent)
	if sealedHome != "" {
		env = append(env, "HOME="+sealedHome)
	}
	for _, e := range sealedEnv {
		if e != "" {
			env = append(env, e)
		}
	}
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, -1, fmt.Errorf("sealed binary exceeded process timeout %v; stderr=%q", timeout, stderr)
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, stderr, exitErr.ExitCode(), nil
		}
		return stdout, stderr, -1, fmt.Errorf("run sealed binary: %w", err)
	}
	return stdout, stderr, 0, nil
}

func listDirNames(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		names = append(names, e.Name())
	}
	return names, nil
}

func fileSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func statOutput(path string) (exists bool, size int64) {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return false, 0
	}
	return true, st.Size()
}

// Run executes kool sandbox…, captures stdout/stderr/exit, records output binary
// stats, and optionally runs inspect and/or the sealed binary (P2).
// Author-named d *session.Doctest is required in one-process mode (blank _ is not enough).
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	koolBin, err := ensureKoolBinary(t, d)
	if err != nil {
		return nil, err
	}
	timeout := req.ProcessTimeout
	if timeout <= 0 {
		timeout = 3 * time.Minute
	}

	resp := &Response{}
	args := buildSandboxArgs(req, "")
	stdout, stderr, code, runErr := runKool(t, koolBin, req.WorkingDir, timeout, args)
	if runErr != nil {
		resp.Stdout = stdout
		resp.Stderr = stderr
		resp.ExitCode = code
		return resp, runErr
	}
	resp.Stdout = stdout
	resp.Stderr = stderr
	resp.ExitCode = code

	if req.Subcommand == "build" && (req.OutputSet || req.Output != "") {
		outPath := resolvePath(req.WorkingDir, req.Output)
		resp.OutputPath = outPath
		resp.OutputExists, resp.OutputSize = statOutput(outPath)
	}

	if req.BuildTwice && req.Subcommand == "build" {
		// Second -o under WorkingDir (same inputs, fresh seal expected).
		secondRel := req.Output + ".second"
		if req.Output == "" {
			secondRel = "out.second"
		}
		secondPath := resolvePath(req.WorkingDir, secondRel)
		args2 := buildSandboxArgs(req, secondRel)
		_, stderr2, code2, runErr2 := runKool(t, koolBin, req.WorkingDir, timeout, args2)
		if runErr2 != nil {
			return resp, runErr2
		}
		if code2 != 0 {
			resp.Stderr = resp.Stderr + "\n[second-build]\n" + stderr2
			if resp.ExitCode == 0 {
				resp.ExitCode = code2
			}
		}
		resp.SecondOutputPath = secondPath
		resp.SecondOutputExists, resp.SecondOutputSize = statOutput(secondPath)
		if resp.OutputExists && resp.SecondOutputExists {
			h1, err1 := fileSHA256(resp.OutputPath)
			h2, err2 := fileSHA256(resp.SecondOutputPath)
			if err1 == nil && err2 == nil {
				resp.BinariesIdentical = h1 == h2
			}
		}
	}

	if req.AfterBuildInspect && resp.ExitCode == 0 && resp.OutputExists {
		inspArgs := []string{"sandbox", "inspect", resp.OutputPath}
		iOut, iErr, iCode, iRunErr := runKool(t, koolBin, req.WorkingDir, timeout, inspArgs)
		if iRunErr != nil {
			return resp, iRunErr
		}
		resp.InspectRan = true
		resp.InspectStdout = iOut
		resp.InspectStderr = iErr
		resp.InspectExitCode = iCode
	}

	// Secondary packs for load-devbox targets: build before sealed run so abs
	// paths exist as real host-GOOS sealed binaries (even if primary only seals
	// path strings via --runtime-load-devbox).
	if len(req.SecondaryPacks) > 0 {
		secPaths, secErr := buildSecondaryPacks(t, koolBin, req.WorkingDir, timeout, req.SecondaryPacks)
		if secErr != nil {
			return resp, secErr
		}
		resp.SecondaryPaths = secPaths
	}

	if req.AfterBuildRun && resp.ExitCode == 0 && resp.OutputExists {
		parent := req.SandboxRootParent
		if parent == "" {
			parent = filepath.Join(req.WorkingDir, "kool-sandbox-root")
		}
		parent = resolvePath(req.WorkingDir, parent)
		resp.SandboxRootParent = parent

		sealedHome := ""
		if req.SealedHome != "" {
			sealedHome = resolvePath(req.WorkingDir, req.SealedHome)
			if err := os.MkdirAll(sealedHome, 0755); err != nil {
				return resp, fmt.Errorf("mkdir SealedHome: %w", err)
			}
			resp.SealedHome = sealedHome
		}

		sealedArgs := sealedRunArgs(req)
		rOut, rErr, rCode, rRunErr := runSealedBinary(t, resp.OutputPath, req.WorkingDir, parent, sealedHome, req.SealedEnv, timeout, sealedArgs)
		if rRunErr != nil {
			resp.RunExecuted = true
			resp.RunStdout = rOut
			resp.RunStderr = rErr
			resp.RunExitCode = rCode
			return resp, rRunErr
		}
		resp.RunExecuted = true
		resp.RunStdout = rOut
		resp.RunStderr = rErr
		resp.RunExitCode = rCode

		// Snapshot remaining session materialize children for cleanup asserts.
		names, listErr := listDirNames(parent)
		if listErr != nil {
			return resp, fmt.Errorf("list KOOL_SANDBOX_ROOT after run: %w", listErr)
		}
		resp.MaterializeRemaining = names
		resp.MaterializeEmpty = len(names) == 0
	}

	return resp, nil
}
```
