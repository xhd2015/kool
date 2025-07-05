package jsontool

import (
	"encoding/json"
	"os"

	"github.com/xhd2015/kool/pkgs/jsondecode"
	"github.com/xhd2015/kool/pkgs/terminal"
)

func HandleCompress(args []string) error {
	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return err
	}

	v, err := jsondecode.SafeUnmarshalJson([]byte(data))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(v)
	if err != nil {
		return err
	}
	return nil
}
