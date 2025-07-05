package stringtool

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/xhd2015/kool/pkgs/terminal"
)

func HandleStrCase(args []string) error {
	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return err
	}

	line := strings.TrimSpace(data)
	fmt.Println(strcase.ToSnake(line))
	fmt.Println(strcase.ToCamel(line))
	fmt.Println(strcase.ToLowerCamel(line))
	fmt.Println(strcase.ToScreamingSnake(line))
	fmt.Println(strcase.ToKebab(line))
	return nil
}
