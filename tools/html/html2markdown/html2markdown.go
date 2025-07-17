package html2markdown

import (
	"fmt"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/xhd2015/kool/pkgs/terminal"
)

func Handle(args []string) error {
	html, err := terminal.ReadOrTerminalDataOrFile(args)
	if err != nil {
		return err
	}
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(
				commonmark.WithStrongDelimiter("__"),
				// ...additional configurations for the plugin
			),

			// ...additional plugins (e.g. table)
		),
	)

	markdown, err := conv.ConvertString(html)
	if err != nil {
		return err
	}
	fmt.Println(markdown)
	return nil
}
