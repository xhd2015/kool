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

const buildHelp = `
kool sandbox build - Pack files/env into a sealed cross-compiled binary

Usage:
  kool sandbox build -o OUTPUT [OPTIONS]

Options:
  -o,--output PATH                 output sealed binary path (required)
  -i,--input DIR                   input directory (meta.yaml, files/, env.yaml)
  --file LOCAL=SANDBOX_REL         pack a local file (repeatable; overrides -i path)
  --env KEY=VALUE                  pack an environment variable (repeatable; overrides -i key)
  --goos OS                        target GOOS (default: host)
  --goarch ARCH                    target GOARCH (default: host)
  -h,--help                        show help message

Input directory layout:
  <dir>/
    meta.yaml     # optional: name, comment, expires_at
    files/        # tree of files to pack
    env.yaml      # KEY: value map

Crypto:
  One-time RSA keypair per build; AES-256-GCM bulk; RSA-OAEP wrap of DEK;
  sealed payload embedded in the output binary.

Examples:
  kool sandbox build -o sandbox.bin -i ./pack
  kool sandbox build -o sandbox.bin --file secret.txt=app/secret.txt --env TOKEN=x
  kool sandbox build -o sandbox.bin --goos linux --goarch amd64 --env X=1
`

func handleBuild(args []string) error {
	var opts buildOpts
	var files []string
	var envs []string

	remain, err := lessflags.
		String("-o,--output", &opts.Output).
		String("-i,--input", &opts.Input).
		StringSlice("--file", &files).
		StringSlice("--env", &envs).
		String("--goos", &opts.Goos).
		String("--goarch", &opts.Goarch).
		Help("-h,--help", buildHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(remain, " "))
	}

	opts.Files = files
	opts.Env = envs

	if opts.Output == "" {
		return fmt.Errorf("requires -o/--output")
	}
	if opts.Goos == "" {
		opts.Goos = runtime.GOOS
	}
	if opts.Goarch == "" {
		opts.Goarch = runtime.GOARCH
	}

	// Validate --env forms early (also done in merge, but clear error).
	for _, e := range opts.Env {
		if _, _, err := splitEnvFlag(e); err != nil {
			return err
		}
	}

	blob, err := mergePack(&opts)
	if err != nil {
		return err
	}

	packJSON, err := marshalPackBlob(blob)
	if err != nil {
		return fmt.Errorf("marshal pack: %w", err)
	}
	sealed, err := seal(packJSON)
	if err != nil {
		return fmt.Errorf("seal pack: %w", err)
	}

	outPath := opts.Output
	if !filepath.IsAbs(outPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		outPath = filepath.Join(cwd, outPath)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	if err := buildSealedBinary(outPath, opts.Goos, opts.Goarch, sealed); err != nil {
		return err
	}

	st, err := os.Stat(outPath)
	if err != nil {
		return fmt.Errorf("stat output: %w", err)
	}

	printBuildSummary(blob, opts.Goos, opts.Goarch, st.Size())
	return nil
}

func printBuildSummary(blob *PackBlob, goos, goarch string, size int64) {
	name := blob.Name
	if name == "" {
		name = "sandbox"
	}
	fmt.Printf("  name      %s\n", name)
	fmt.Printf("  goos      %s\n", goos)
	fmt.Printf("  goarch    %s\n", goarch)
	fmt.Printf("  files     %d\n", len(blob.Files))
	fmt.Printf("  env       %d\n", len(blob.Env))
	fmt.Printf("  size      %d\n", size)
}

// buildSealedBinary writes a temp Go module with embedded payload and cross-compiles it.
// The runner imports github.com/xhd2015/kool/tools/sandbox via a replace to the
// local kool module so sealed binaries share unseal/materialize/exec with the CLI.
func buildSealedBinary(outPath, goos, goarch string, sealed []byte) error {
	tmp, err := os.MkdirTemp("", "kool-sandbox-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	modRoot, err := findKoolModuleRoot()
	if err != nil {
		return err
	}

	// Replace path must be absolute for go.mod when the build runs outside the module.
	absModRoot, err := filepath.Abs(modRoot)
	if err != nil {
		return err
	}

	mod := fmt.Sprintf("module sandbox-runner\n\ngo 1.22\n\nrequire github.com/xhd2015/kool v0.0.0\n\nreplace github.com/xhd2015/kool => %s\n",
		filepath.ToSlash(absModRoot))
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(mod), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, "payload.seal"), sealed, 0644); err != nil {
		return err
	}
	// Runner must *use* payload so the linker keeps //go:embed data.
	mainSrc := `package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/kool/tools/sandbox"
)

//go:embed payload.seal
var payload []byte

func main() {
	if len(payload) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing sealed payload")
		os.Exit(1)
	}
	os.Exit(sandbox.RunEmbedded(payload, os.Args[1:]))
}
`
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte(mainSrc), 0644); err != nil {
		return err
	}

	// Resolve transitive deps of the replaced kool module into this temp module.
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tmp
	tidy.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := tidy.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy sealed runner: %w\n%s", err, out)
	}

	cmd := exec.Command("go", "build", "-o", outPath, ".")
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
		"CGO_ENABLED=0",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build sealed binary: %w\n%s", err, out)
	}
	return nil
}

// findKoolModuleRoot locates the local github.com/xhd2015/kool module directory
// so the sealed-runner temp module can replace-import tools/sandbox.
func findKoolModuleRoot() (string, error) {
	var candidates []string

	if _, file, _, ok := runtime.Caller(0); ok && file != "" && filepath.IsAbs(file) {
		// tools/sandbox/build.go → module root ../..
		candidates = append(candidates, filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..")))
	}

	if cwd, err := os.Getwd(); err == nil {
		p := cwd
		for i := 0; i < 16; i++ {
			candidates = append(candidates, p)
			parent := filepath.Dir(p)
			if parent == p {
				break
			}
			p = parent
		}
	}

	// Prefer go list when the toolchain can resolve the main module.
	list := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/xhd2015/kool")
	if out, err := list.Output(); err == nil {
		dir := strings.TrimSpace(string(out))
		if dir != "" {
			candidates = append([]string{dir}, candidates...)
		}
	}

	seen := map[string]bool{}
	for _, c := range candidates {
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		if isKoolModuleRoot(c) {
			return c, nil
		}
	}
	return "", fmt.Errorf("cannot locate github.com/xhd2015/kool module root (needed to build sealed runner)")
}

func isKoolModuleRoot(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	// First non-empty line should be the module path (allow trailing comments/space).
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		return line == "module github.com/xhd2015/kool" || strings.HasPrefix(line, "module github.com/xhd2015/kool ")
	}
	return false
}

