package jsontool

import (
	"encoding/json"
	"os"

	"github.com/xhd2015/kool/pkgs/jsondecode"
	"github.com/xhd2015/kool/pkgs/terminal"
)

func HandlePretty(args []string) error {
	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return err
	}

	v, err := jsondecode.UnmarshalSafeAny([]byte(data))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(v)
	if err != nil {
		return err
	}
	return nil
}
