package terminal

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/kool/pkgs/ioread"
	"golang.org/x/term"
)

func IsStdinTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func ReadOrTerminalData(args []string) (string, error) {
	return readOrTerminalData(args, false)
}

func ReadOrTerminalDataOrFile(args []string) (string, error) {
	return readOrTerminalData(args, true)
}

func readOrTerminalData(args []string, tryFile bool) (string, error) {
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
	if tryFile {
		return ioread.ReadOrContent(args[0])
	}
	return args[0], nil
}
