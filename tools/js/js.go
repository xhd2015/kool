package js

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/kool/pkgs/terminal"
	"github.com/xhd2015/xgo/support/cmd"
)

const help = `
kool js evaluate js script with stdin data

Usage: kool js <cmd> [OPTIONS]

Examples:
  echo '["A","B"]' | kool js 'console.log(data.join("\n"))' 
`

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires expr")
	}

	expr := args[0]
	args = args[1:]

	if expr == "--help" || expr == "help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}

	if expr == "" {
		return fmt.Errorf("requires expr")
	}

	if terminal.IsStdinTerminal() {
		nodeArgs := []string{"-e", expr}
		nodeArgs = append(nodeArgs, args...)
		return cmd.New().Stdin(os.Stdin).Run("node", nodeArgs...)
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	script := fmt.Sprintf("(function(data){\n%s\n})(%s)", expr, string(data))
	nodeArgs := []string{"-e", script}
	nodeArgs = append(nodeArgs, args...)
	return cmd.Run("node", nodeArgs...)
}
