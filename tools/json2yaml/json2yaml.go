package json2yaml

import (
	"fmt"

	"github.com/xhd2015/kool/pkgs/jsondecode"
	"github.com/xhd2015/kool/pkgs/terminal"
	"gopkg.in/yaml.v3"
)

func Handle(args []string) error {
	data, err := terminal.ReadOrTerminalDataOrFile(args)
	if err != nil {
		return err
	}

	ymlData, err := jsonToYaml([]byte(data))
	if err != nil {
		return err
	}
	fmt.Println(string(ymlData))
	return nil
}

func jsonToYaml(jsonData []byte) ([]byte, error) {
	v, err := jsondecode.UnmarshalSafeAny(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Unmarshal YAML to a generic interface
	return yaml.Marshal(v)
}
