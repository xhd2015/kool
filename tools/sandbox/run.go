package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RunEmbedded decrypts a sealed pack payload, materializes files under a session
// directory, applies packed env + SANDBOX_ROOT, executes the guest command with
// cwd at the materialize root, then removes the session directory.
//
// args are the sealed binary's argv after the program name (os.Args[1:]).
// A leading "--" is skipped. Returns the guest process exit code (or a non-zero
// code on runner errors).
func RunEmbedded(sealed []byte, args []string) int {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing command")
		fmt.Fprintln(os.Stderr, "Usage: <sandbox.bin> [--] <command> [args...]")
		return 1
	}

	blob, err := unseal(sealed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unseal failed: %v\n", err)
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

	if err := materializeFiles(root, blob); err != nil {
		fmt.Fprintf(os.Stderr, "Error: write files: %v\n", err)
		return 1
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: resolve SANDBOX_ROOT: %v\n", err)
		return 1
	}

	env := append([]string{}, os.Environ()...)
	for k, v := range blob.Env {
		env = append(env, k+"="+v)
	}
	env = append(env, "SANDBOX_ROOT="+absRoot)
	// Keep logical path for pwd(1) on macOS (/var → /private/var): without PWD,
	// shells report the physical path and SANDBOX_ROOT would disagree with cwd.
	env = append(env, "PWD="+absRoot)

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
		rel := filepath.Clean(filepath.FromSlash(f.Path))
		if rel == "." || rel == "" {
			return fmt.Errorf("invalid empty path")
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("invalid path: %s", f.Path)
		}
		// Reject absolute paths after FromSlash/Clean.
		if filepath.IsAbs(rel) {
			return fmt.Errorf("invalid absolute path: %s", f.Path)
		}

		dest := filepath.Join(root, rel)
		// Ensure dest stays under root.
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
			return fmt.Errorf("path escapes sandbox root: %s", f.Path)
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return err
		}
		mode := os.FileMode(f.Mode) & os.ModePerm
		if mode == 0 {
			mode = 0o644
		}
		if err := os.WriteFile(dest, f.Content, mode); err != nil {
			return err
		}
		if err := os.Chmod(dest, mode); err != nil {
			return err
		}
	}
	return nil
}
