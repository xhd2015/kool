package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

// RunEmbedded decrypts a sealed pack payload, materializes files under a session
// directory, applies packed env + SANDBOX_ROOT, executes the guest command with
// cwd at the materialize root, then removes the session directory.
//
// When the pack has HomeLinked set, the runner captures the host real home once,
// seeds top-level absolute symlinks into the session root, overlays packed files
// with explode-on-demand for intermediate symlinks, and forces guest HOME to the
// session root (packed HOME is policed).
//
// Optional --load-devbox ABS (repeatable, StopOnFirstArg) merges additional sealed
// sandbox binaries (Files later-wins overlay + hard env conflict merge).
//
// args are the sealed binary's argv after the program name (os.Args[1:]).
// Returns the guest process exit code (or a non-zero code on runner errors).
func RunEmbedded(sealed []byte, args []string) int {
	var loadDevboxPaths []string
	remain, err := lessflags.
		StringSlice("--load-devbox", &loadDevboxPaths).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	args = remain

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing command")
		fmt.Fprintln(os.Stderr, "Usage: <sandbox.bin> [--load-devbox ABS]... [--] <command> [args...]")
		return 1
	}

	blob, err := unseal(sealed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unseal failed: %v\n", err)
		return 1
	}

	// Capture real home before any guest HOME override (process env is not
	// mutated; cmd.Env is built separately). Prefer $HOME so tests can inject
	// a fake real home via child process environment.
	var realHome string
	if blob.HomeLinked {
		realHome, err = captureRealHome()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: resolve real home: %v\n", err)
			return 1
		}
	}

	// Walk sealed RuntimeLoadDevbox then adhoc --load-devbox (notices on stdout).
	loads, err := walkLoadDevboxes(blob, loadDevboxPaths)
	if err != nil {
		// walkLoadDevboxes already prefixes Error: on returned errors.
		msg := err.Error()
		if !strings.HasPrefix(msg, "Error:") {
			msg = "Error: " + msg
		}
		fmt.Fprintln(os.Stderr, msg)
		return 1
	}

	root, err := createMaterializeRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: materialize: %v\n", err)
		return 1
	}
	defer func() {
		_ = os.RemoveAll(root)
	}()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: resolve SANDBOX_ROOT: %v\n", err)
		return 1
	}

	// Env merge (primary + loads) with hard conflicts; HOME policy when home-linked.
	mergedEnv, err := mergeSandboxEnv(blob, loads, blob.HomeLinked, absRoot)
	if err != nil {
		msg := err.Error()
		if !strings.HasPrefix(msg, "Error:") {
			msg = "Error: " + msg
		}
		fmt.Fprintln(os.Stderr, msg)
		return 1
	}

	// Materialize files: home-linked and/or load-devbox use linkoverlay.Apply;
	// plain primary-only packs keep the existing materialize path.
	if blob.HomeLinked || len(loads) > 0 {
		if err := materializeWithLoads(absRoot, realHome, blob.HomeLinked, blob, loads); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write files: %v\n", err)
			return 1
		}
	} else {
		if err := materializeFiles(root, blob); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write files: %v\n", err)
			return 1
		}
	}

	env := append([]string{}, os.Environ()...)
	for k, v := range mergedEnv {
		if blob.HomeLinked && k == "HOME" {
			// Policed above; never apply packed HOME when home-linked.
			continue
		}
		env = append(env, k+"="+v)
	}
	env = append(env, "SANDBOX_ROOT="+absRoot)
	// Keep logical path for pwd(1) on macOS (/var → /private/var): without PWD,
	// shells report the physical path and SANDBOX_ROOT would disagree with cwd.
	env = append(env, "PWD="+absRoot)
	if blob.HomeLinked {
		env = append(env, "HOME="+absRoot)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = absRoot
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	fmt.Fprintf(os.Stderr, "Error: exec %s: %v\n", args[0], err)
	return 1
}

// captureRealHome returns the host home directory used for home-linked seeding.
// Prefer $HOME (injectable by tests) then os.UserHomeDir().
func captureRealHome() (string, error) {
	if home := strings.TrimSpace(os.Getenv("HOME")); home != "" {
		return home, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", fmt.Errorf("empty home directory")
	}
	return home, nil
}

// createMaterializeRoot picks a session directory:
//  1. unique child of KOOL_SANDBOX_ROOT when set
//  2. /dev/shm/kool-sandbox/<id> on Linux when writable
//  3. os.MkdirTemp fallback (warns on stderr)
func createMaterializeRoot() (string, error) {
	if parent := strings.TrimSpace(os.Getenv("KOOL_SANDBOX_ROOT")); parent != "" {
		if err := os.MkdirAll(parent, 0o700); err != nil {
			return "", fmt.Errorf("KOOL_SANDBOX_ROOT: %w", err)
		}
		dir, err := os.MkdirTemp(parent, "kool-sandbox-*")
		if err != nil {
			return "", err
		}
		// MkdirTemp is 0700; re-chmod for explicit contract.
		if err := os.Chmod(dir, 0o700); err != nil {
			_ = os.RemoveAll(dir)
			return "", err
		}
		return dir, nil
	}

	if runtime.GOOS == "linux" {
		if st, err := os.Stat("/dev/shm"); err == nil && st.IsDir() {
			base := "/dev/shm/kool-sandbox"
			if err := os.MkdirAll(base, 0o700); err == nil {
				if dir, err := os.MkdirTemp(base, "sess-*"); err == nil {
					_ = os.Chmod(dir, 0o700)
					return dir, nil
				}
			}
		}
	}

	dir, err := os.MkdirTemp("", "kool-sandbox-*")
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr, "Warning: using system temp dir for sandbox (set KOOL_SANDBOX_ROOT to override)")
	return dir, nil
}

func materializeFiles(root string, blob *PackBlob) error {
	for _, f := range blob.Files {
		rel, err := sanitizeRelPath(f.Path)
		if err != nil {
			return err
		}
		dest := filepath.Join(root, rel)
		if err := ensureUnderRoot(root, dest, f.Path); err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return err
		}
		if err := writePackedFile(dest, f); err != nil {
			return err
		}
	}
	return nil
}

func writePackedFile(dest string, f PackFile) error {
	mode := os.FileMode(f.Mode) & os.ModePerm
	if mode == 0 {
		mode = 0o644
	}
	if err := os.WriteFile(dest, f.Content, mode); err != nil {
		return err
	}
	return os.Chmod(dest, mode)
}

func sanitizeRelPath(path string) (string, error) {
	rel := filepath.Clean(filepath.FromSlash(path))
	if rel == "." || rel == "" {
		return "", fmt.Errorf("invalid empty path")
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path: %s", path)
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid absolute path: %s", path)
	}
	return rel, nil
}

func ensureUnderRoot(root, dest, orig string) error {
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	relCheck, err := filepath.Rel(absRoot, absDest)
	if err != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes sandbox root: %s", orig)
	}
	return nil
}
