package stringtool

import (
	"fmt"
	"strconv"

	"github.com/xhd2015/kool/pkgs/terminal"
)

func HandleQuote(args []string) error {
	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return err
	}
	quoteStr := strconv.Quote(data)
	fmt.Println(quoteStr)
	return nil
}

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
