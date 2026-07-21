package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
)

const inspectHelp = `
kool sandbox inspect - Show packed paths, content hashes, and env keys

Usage:
  kool sandbox inspect <binary>
  kool sandbox inspect -h|--help

Output:
  name, file paths with content SHA-256 hashes, and env keys only
  (never secret env values).
`

func handleInspect(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires binary path: kool sandbox inspect <binary>")
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		txt := strings.TrimPrefix(inspectHelp, "\n")
		fmt.Print(txt)
		if !strings.HasSuffix(txt, "\n") {
			fmt.Println()
		}
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args[1:], " "))
	}

	path := args[0]
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read binary: %w", err)
	}
	sealed, err := findSealedPayload(raw)
	if err != nil {
		return err
	}
	blob, err := unseal(sealed)
	if err != nil {
		return fmt.Errorf("unseal: %w", err)
	}

	name := blob.Name
	if name == "" {
		name = "sandbox"
	}
	fmt.Printf("name: %s\n", name)
	fmt.Printf("files: %d\n", len(blob.Files))
	for _, f := range blob.Files {
		sum := sha256.Sum256(f.Content)
		fmt.Printf("  %s  %s\n", f.Path, hex.EncodeToString(sum[:]))
	}
	keys := make([]string, 0, len(blob.Env))
	for k := range blob.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("env: %d\n", len(keys))
	for _, k := range keys {
		fmt.Printf("  %s\n", k)
	}
	return nil
}
