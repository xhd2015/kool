package stringtool

import (
	"fmt"
	"strconv"

	"github.com/xhd2015/kool/pkgs/terminal"
)

func HandleUnquote(args []string) error {
	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return err
	}
	unquoteStr, err := strconv.Unquote(data)
	if err != nil {
		return err
	}
	fmt.Println(unquoteStr)
	return nil
}
