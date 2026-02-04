package line

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: kool git line history <file> <line1> [line2]

Show the history of specific lines in a file.

Examples:
  kool git line history src/main.go 10          # show history of line 10
  kool git line history src/main.go 10 20       # show history of lines 10-20
  kool git line history src/main.go 10-20       # show history of lines 10-20
`

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires subcommand: history")
	}

	subcmd := args[0]
	subArgs := args[1:]

	switch subcmd {
	case "-h", "--help", "help":
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	case "history":
		return handleHistory(subArgs)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcmd)
	}
}

const historyHelp = `
Usage: kool git line history <file> <line1> [line2]

Show the history of specific lines in a file.

Examples:
  kool git line history src/main.go 10          # show history of line 10
  kool git line history src/main.go 10 20       # show history of lines 10-20
  kool git line history src/main.go 10-20       # show history of lines 10-20
`

func handleHistory(args []string) error {
	var verbose bool
	// "github.com/xhd2015/less-gen/flags"
	args, err := flags.Help("-h,--help", historyHelp).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: kool git line history <file> <line1> [line2]")
	}

	file := args[0]
	lineSpec := args[1]

	args = args[2:]

	var line1, line2 int

	// Check if lineSpec contains hyphen (e.g., "10-20")
	if strings.Contains(lineSpec, "-") {
		parts := strings.Split(lineSpec, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid line range: %s", lineSpec)
		}
		line1, err = strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid line number: %s", parts[0])
		}
		line2, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid line number: %s", parts[1])
		}
	} else {
		// Single line number
		line1, err = strconv.Atoi(lineSpec)
		if err != nil {
			return fmt.Errorf("invalid line number: %s", lineSpec)
		}
		// Check if there's a second argument for line2
		if len(args) > 0 {
			line2, err = strconv.Atoi(args[0])
			args = args[1:]
			if err != nil {
				return fmt.Errorf("invalid line number: %s", args[0])
			}
		} else {
			line2 = line1
		}
	}


	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	lineRange := fmt.Sprintf("%d,%d", line1, line2)

	// Build git log -L command
	cmd := exec.Command("git", "log", "-L", lineRange+":"+file, "-p")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
