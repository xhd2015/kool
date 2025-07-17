package html2text

import (
	"fmt"

	"github.com/k3a/html2text"
	"github.com/xhd2015/kool/pkgs/terminal"
)

func Handle(args []string) error {
	html, err := terminal.ReadOrTerminalDataOrFile(args)
	if err != nil {
		return err
	}
	text := html2text.HTML2Text(html)
	fmt.Println(text)
	return nil
}
