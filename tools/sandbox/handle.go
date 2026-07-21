package sandbox

import (
	"fmt"
	"strings"
)

const rootHelp = `
kool sandbox - Pack files/env into a sealed sandbox binary

Usage:
  kool sandbox build -o OUTPUT [OPTIONS]
  kool sandbox inspect <binary>
  kool sandbox -h|--help

Commands:
  build     pack files and env into a sealed cross-compiled binary
  inspect   list packed paths, content hashes, and env keys (no secret values)

Build options:
  -o,--output PATH                 output sealed binary path (required)
  -i,--input DIR                   input directory (meta.yaml, files/, env.yaml)
  --file LOCAL=SANDBOX_REL         pack a local file (repeatable)
  --env KEY=VALUE                  pack an environment variable (repeatable)
  --goos OS                        target GOOS (default: host)
  --goarch ARCH                    target GOARCH (default: host)

Examples:
  kool sandbox build -o sandbox.bin -i ./pack
  kool sandbox build -o sandbox.bin --file cfg.txt=app/cfg.txt --env TOKEN=x
  kool sandbox build -o sandbox.bin --goos linux --goarch amd64 --env X=1
  kool sandbox inspect ./sandbox.bin
`

// Handle is the production entry for kool sandbox.
func Handle(args []string) error {
	if len(args) == 0 {
		// Bare `kool sandbox` → help-ish usage error with guidance.
		txt := strings.TrimPrefix(rootHelp, "\n")
		fmt.Print(txt)
		if !strings.HasSuffix(txt, "\n") {
			fmt.Println()
		}
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		txt := strings.TrimPrefix(rootHelp, "\n")
		fmt.Print(txt)
		if !strings.HasSuffix(txt, "\n") {
			fmt.Println()
		}
		return nil
	case "build":
		return handleBuild(args[1:])
	case "inspect":
		return handleInspect(args[1:])
	default:
		return fmt.Errorf("unrecognized command: %s", args[0])
	}
}
