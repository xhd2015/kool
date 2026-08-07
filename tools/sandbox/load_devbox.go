package sandbox

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/linkoverlay"
	"golang.org/x/term"
)

// loadedDevbox is one successfully loaded sandbox binary in walk order.
type loadedDevbox struct {
	Abs   string
	Files []PackFile
	Env   map[string]string
}

// walkLoadDevboxes loads sealed RuntimeLoadDevbox paths then adhoc CLI paths.
// First-seen cleaned abs identity; each load prints a notice and DFS-walks nested
// RuntimeLoadDevbox with cycle skip (already-seen abs → skip, no second notice).
func walkLoadDevboxes(primary *PackBlob, adhoc []string) ([]loadedDevbox, error) {
	seen := make(map[string]struct{})
	var loaded []loadedDevbox

	var loadOne func(path string) error
	loadOne = func(path string) error {
		if path == "" {
			return fmt.Errorf("Error: --load-devbox requires an absolute path (got empty)")
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("Error: --load-devbox requires an absolute path (got relative: %q)", path)
		}
		abs := filepath.Clean(path)
		if _, ok := seen[abs]; ok {
			return nil
		}
		seen[abs] = struct{}{}

		blob, err := openAndUnsealDevbox(abs)
		if err != nil {
			return err
		}
		printLoadDevboxNotice(abs)
		loaded = append(loaded, loadedDevbox{
			Abs:   abs,
			Files: blob.Files,
			Env:   blob.Env,
		})
		for _, nested := range blob.RuntimeLoadDevbox {
			if err := loadOne(nested); err != nil {
				return err
			}
		}
		return nil
	}

	// Seed list: primary sealed paths (pack order) then adhoc CLI (first-seen via seen).
	for _, p := range primary.RuntimeLoadDevbox {
		if err := loadOne(p); err != nil {
			return nil, err
		}
	}
	for _, p := range adhoc {
		if err := loadOne(p); err != nil {
			return nil, err
		}
	}
	return loaded, nil
}

// openAndUnsealDevbox reads a host path, finds the sealed payload, and unseals it.
func openAndUnsealDevbox(abs string) (*PackBlob, error) {
	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("Error: load-devbox %s: %w", abs, err)
	}
	sealed, err := findSealedPayload(raw)
	if err != nil {
		return nil, fmt.Errorf("Error: load-devbox %s: %w", abs, err)
	}
	blob, err := unseal(sealed)
	if err != nil {
		return nil, fmt.Errorf("Error: load-devbox %s: unseal failed: %w", abs, err)
	}
	return blob, nil
}

// printLoadDevboxNotice writes "notice: loading devbox <abs>" to stdout.
// Grey ANSI only when stdout is a terminal; plain text otherwise (tests are non-TTY).
func printLoadDevboxNotice(abs string) {
	msg := "notice: loading devbox " + abs
	if term.IsTerminal(int(os.Stdout.Fd())) {
		// Grey / dim for interactive terminals only.
		fmt.Fprintln(os.Stdout, "\x1b[90m"+msg+"\x1b[0m")
		return
	}
	fmt.Fprintln(os.Stdout, msg)
}

// mergeSandboxEnv unions primary then each load's Env. Same key from any two
// sandboxes is a hard conflict (even same value). Host env is not in the set.
// sourceLabels: primary is "current sandbox"; loads use their abs path.
// When homeLinked, HOME is policed separately (not claimed for conflicts).
func mergeSandboxEnv(primary *PackBlob, loads []loadedDevbox, homeLinked bool, absRoot string) (map[string]string, error) {
	claimed := make(map[string]string) // key -> source label
	merged := make(map[string]string)

	claim := func(key, value, source string) error {
		if homeLinked && key == "HOME" {
			// Policed by applyHomeLinkedHOMEPolicy; never apply packed HOME.
			return nil
		}
		if prev, ok := claimed[key]; ok {
			return fmt.Errorf("Error: incompatible env %s: %s vs %s", key, prev, source)
		}
		claimed[key] = source
		merged[key] = value
		return nil
	}

	for k, v := range primary.Env {
		if err := claim(k, v, "current sandbox"); err != nil {
			return nil, err
		}
	}
	for _, ld := range loads {
		for k, v := range ld.Env {
			if err := claim(k, v, ld.Abs); err != nil {
				return nil, err
			}
		}
	}

	if homeLinked {
		if err := applyHomeLinkedHOMEPolicy(primary, loads, absRoot); err != nil {
			return nil, err
		}
	}
	return merged, nil
}

// applyHomeLinkedHOMEPolicy applies packed HOME rules for primary and each load:
// equal to abs session root → notice on stderr and ignore; any other value → error.
func applyHomeLinkedHOMEPolicy(primary *PackBlob, loads []loadedDevbox, absRoot string) error {
	check := func(packedHome string) error {
		if packedHome == absRoot {
			fmt.Fprintln(os.Stderr, "Notice: --home-linked already sets HOME; ignoring packed HOME")
			return nil
		}
		return fmt.Errorf("Error: HOME cannot be set when --home-linked")
	}
	if packedHome, ok := primary.Env["HOME"]; ok {
		if err := check(packedHome); err != nil {
			return err
		}
	}
	for _, ld := range loads {
		if packedHome, ok := ld.Env["HOME"]; ok {
			if err := check(packedHome); err != nil {
				return err
			}
		}
	}
	return nil
}

// packFilesToOverlay converts PackFile entries to linkoverlay.File.
func packFilesToOverlay(files []PackFile) []linkoverlay.File {
	out := make([]linkoverlay.File, 0, len(files))
	for _, f := range files {
		out = append(out, linkoverlay.File{
			Path:    f.Path,
			Mode:    f.Mode,
			Content: f.Content,
		})
	}
	return out
}

// materializeWithLoads applies file layers via linkoverlay.Apply (later wins):
//
//	[ Layer{Dir: realHome} if homeLinked ]
//	→ Layer{Files: primary Files}
//	→ Layer{Files: each loaded Files in walk order}
func materializeWithLoads(absRoot, realHome string, homeLinked bool, primary *PackBlob, loads []loadedDevbox) error {
	var layers []linkoverlay.Layer
	if homeLinked {
		layers = append(layers, linkoverlay.Layer{Dir: realHome})
	}
	layers = append(layers, linkoverlay.Layer{Files: packFilesToOverlay(primary.Files)})
	for _, ld := range loads {
		layers = append(layers, linkoverlay.Layer{Files: packFilesToOverlay(ld.Files)})
	}
	return linkoverlay.Apply(absRoot, layers...)
}
