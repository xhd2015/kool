package debug

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/kool/tools/dlv"
	"github.com/xhd2015/kool/tools/go/run"
)

const help = `
kool debug

Debug tools

Examples:
  kool debug go run ./main.go
`

func Handle(args []string) error {
	if len(args) == 0 {
		return errors.New("requires command, try 'kool debug --help'")
	}
	cmd := args[0]
	args = args[1:]

	if cmd == "--help" || cmd == "help" {
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	}
	switch cmd {
	case "go":
		return handleGo(args)
	default:
		// check if cmd is a file
		fileStat, err := os.Stat(cmd)
		if err == nil {
			if fileStat.IsDir() {
				return fmt.Errorf("%s refers to a directory, want a debug binary", cmd)
			}
			binary := cmd
			// check if binary is a go-built binary
			if dlv.HasMainMain(binary) {
				return run.DebugBinary(binary, args, run.DebugOptions{
					PassStdin: false,
				})
			}
			return fmt.Errorf("expected a go-built executable binary, got %s", binary)
		}
	}

	return fmt.Errorf("unsupported command: %s", cmd)
}

const goHelp = `
kool debug go

Debug go tools
`

func handleGo(args []string) error {
	if len(args) == 0 {
		return errors.New("requires command, try 'kool debug go --help'")
	}
	cmd := args[0]
	args = args[1:]

	if cmd == "--help" || cmd == "help" {
		fmt.Println(strings.TrimPrefix(goHelp, "\n"))
		return nil
	}

	if cmd == "run" {
		return run.HandleOpts(args, run.Options{
			AcceptDebugFlag: false,
			IsDebug:         true,
		})
	}

	return fmt.Errorf("unsupported command: %s", cmd)
}
