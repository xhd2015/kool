# kool sandbox (P1 package + P2 sealed-binary run)

`kool sandbox build` packs sandbox files and env into a **cross-compiled sealed
binary**. One-time RSA keypair per build; bulk data under AES-256-GCM; DEK
wrapped with RSA-OAEP; private key + ciphertext embedded via `//go:embed`.
P1 asserts the **built artifact** and security bar. P2 asserts **unpack +
materialize + exec** of the sealed binary on the **host** GOOS/GOARCH (classic
TDD: run leaves are RED until the runner is implemented).

## Version

0.0.2

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
  on path/env-key conflict.
- **Sealer** — per-build RSA keypair; AES-256-GCM for PackBlob; RSA-OAEP wrap of
  DEK; embed sealed blob in runner binary.
- **Output binary (sealed runner)** — path from `-o` / `--output`; target OS/arch
  from `--goos` / `--goarch` (default host runtime). At run time it unseals the
  payload, materializes files, applies env, execs the guest command.
- **Materialize root** — session directory under Linux default
  `/dev/shm/kool-sandbox/<id>/`, or under env `KOOL_SANDBOX_ROOT` (required for
  macOS/doctest hosts), with temp-dir fallback. Child process `cwd` and env
  `SANDBOX_ROOT` are this absolute path. Removed best-effort when the process
  exits.
- **Inspector** — `kool sandbox inspect <binary>`: name, file paths + hashes,
  env keys only.

### Behaviors

- **Root help** — `kool sandbox -h|--help` lists `build` (and `inspect` when
  present) and principal flags; exit 0; stdout ends with `\n`.
- **Build help** — `kool sandbox build -h|--help` documents `-o/-i/--file/--env/
  --goos/--goarch`; exit 0; stdout ends with `\n`.
- **Build validation** — missing `-o`, empty pack (no files and no env after
  merge), `--env` without `=`, missing local path for `--file` → non-zero;
  message on stderr; no output binary required.
- **Build from input dir** — `-i` with `files/` and/or `env.yaml` (+ optional
  `meta.yaml`) → exit 0; binary at `-o` with size > 0; stdout mentions sandbox
  name (or default), files/env counts, size; ends with `\n`.
- **Build from flags only** — no `-i`; only `--file` / `--env` → success when
  pack non-empty.
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
  style (mentions command/usage). No ssh-agent special case.
- **Runner run stub (pre-P2 implement)** — until runtime lands, sealed binary may
  print `Error: run not implemented` and exit non-zero; P2 leaves stay **RED**.

### Pack / seal model (conceptual)

```text
PackBlob: Version, Name, CreatedAt, ExpiresAt?, Files[{Path,Mode,Content}], Env
Sealed: RSA private (one-time) + RSA-OAEP(AES-256 DEK) + AES-GCM(PackBlob)
```

### Inspect CLI (P1 helper surface)

```text
kool sandbox inspect <binary>
  -> exit 0; stdout lists name, file paths (+ content hashes), env keys only
```

### Sealed binary CLI (P2)

```text
./sandbox.bin [--] <command> [args...]
  env KOOL_SANDBOX_ROOT=<parent>   # force materialize parent (tests / macOS)
  -> unseal, materialize under <parent>/<session>/, exec command, cleanup
```

## Decision Tree

```
sandbox/
├── help/                                   [usage; exit 0; no build]
│   ├── root/                               kool sandbox --help
│   └── build/                              kool sandbox build --help
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
    ├── cleanup/                            [session dir lifecycle]
    │   └── removes-materialize-dir/        parent empty after successful run
    └── exit-code/                          [propagate child status]
        └── child-nonzero/                  sh -c 'exit 42' → exit 42
```

## Test Index

| Leaf | Description |
|------|-------------|
| `help/root/` | Root `--help` exit 0; mentions build and key flags; trailing `\n` |
| `help/build/` | Build `--help` exit 0; documents `-o/-i/--file/--env/--goos/--goarch` |
| `build/validation/missing-output/` | Build without `-o` → non-zero; stderr mentions output/`-o` |
| `build/validation/empty-pack/` | No input content → non-zero; empty pack rejected |
| `build/validation/bad-env-flag/` | `--env NOTVALID` → non-zero; message mentions env/`=` |
| `build/validation/missing-file-source/` | `--file` local path missing → non-zero |
| `build/from-input-dir/files-and-env/` | Fixture dir → exit 0; binary size > 0; stdout files/env counts |
| `build/from-input-dir/meta-name/` | `meta.yaml` name appears in stdout |
| `build/from-flags/file-and-env-flags-only/` | Flags only → success binary |
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
| `run/cleanup/removes-materialize-dir/` | After exit 0, no session children under materialize parent |
| `run/exit-code/child-nonzero/` | Guest `exit 42` → sealed binary exit code 42 |

## How to Run

```sh
doctest vet ./tests/sandbox
doctest test ./tests/sandbox
```

Classic TDD: P1 leaves are **GREEN** once build/inspect lands. P2 `run/*` leaves
are **RED** until the sealed runner implements unseal/materialize/exec (not
stubbed with `Error: run not implemented`). `Run` session-builds `kool` from the
module root so leaves exercise the workspace binary. Run leaves build for **host
GOOS/GOARCH** (no `--goos linux`) so the sealed binary can execute on the doctest
machine, and force materialize via `KOOL_SANDBOX_ROOT`.

```go
import (
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

	// BuildTwice: run build twice with same inputs to two different -o paths
	// (security-bar/two-builds-differ).
	BuildTwice bool

	// AfterBuildInspect: on first build exit 0, run `kool sandbox inspect <Output>`.
	AfterBuildInspect bool

	// AfterBuildRun: on first build exit 0 + binary exists, execute the sealed
	// binary under WorkingDir with KOOL_SANDBOX_ROOT (P2 run leaves).
	AfterBuildRun bool
	// SealedArgs is the argv passed to the sealed binary (command + args).
	// Empty means invoke the binary with no args (missing-command).
	SealedArgs []string
	// SealedDoubleDash inserts `--` before SealedArgs (ends runner flags).
	SealedDoubleDash bool
	// SandboxRootParent is the absolute (or WorkingDir-relative) path exported as
	// KOOL_SANDBOX_ROOT for the sealed process. Empty → WorkingDir/kool-sandbox-root.
	SandboxRootParent string

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
	RunExecuted        bool
	RunStdout          string
	RunStderr          string
	RunExitCode        int
	SandboxRootParent  string   // absolute KOOL_SANDBOX_ROOT used
	MaterializeRemaining []string // entry names still under parent after process exit
	MaterializeEmpty   bool     // true when parent has no remaining children
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
		cmd := exec.Command("go", "build", "-o", bin, ".")
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
	}
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

// runSealedBinary executes a host-built sealed sandbox binary with KOOL_SANDBOX_ROOT set.
func runSealedBinary(t *testing.T, binPath, workingDir, sandboxRootParent string, timeout time.Duration, args []string) (stdout, stderr string, exitCode int, runErr error) {
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
	cmd.Env = append(os.Environ(), "KOOL_SANDBOX_ROOT="+sandboxRootParent)
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

	if req.AfterBuildRun && resp.ExitCode == 0 && resp.OutputExists {
		parent := req.SandboxRootParent
		if parent == "" {
			parent = filepath.Join(req.WorkingDir, "kool-sandbox-root")
		}
		parent = resolvePath(req.WorkingDir, parent)
		resp.SandboxRootParent = parent

		sealedArgs := append([]string(nil), req.SealedArgs...)
		if req.SealedDoubleDash {
			sealedArgs = append([]string{"--"}, sealedArgs...)
		}
		rOut, rErr, rCode, rRunErr := runSealedBinary(t, resp.OutputPath, req.WorkingDir, parent, timeout, sealedArgs)
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
