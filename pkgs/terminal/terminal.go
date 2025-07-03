package terminal

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func IsStdinTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func ReadOrTerminalData(args []string) (string, error) {
	if len(args) == 0 {
		isTTY := term.IsTerminal(int(os.Stdin.Fd()))
		if isTTY {
			return "", fmt.Errorf("no data")
		}
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(stdinData), nil
	}

	if len(args) > 1 {
		return "", fmt.Errorf("unrecognized extra arguments: %s", strings.Join(args[1:], ", "))
	}
	return args[0], nil
}
