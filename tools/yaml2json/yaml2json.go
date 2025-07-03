package yaml2json

import (
	"encoding/json"
	"fmt"

	"github.com/xhd2015/kool/pkgs/terminal"
	"gopkg.in/yaml.v3"
)

func Handle(args []string) error {
	data, err := terminal.ReadOrTerminalDataOrFile(args)
	if err != nil {
		return err
	}

	json, err := yamlToJSON([]byte(data))
	if err != nil {
		return err
	}
	fmt.Println(string(json))
	return nil
}

func yamlToJSON(yamlData []byte) ([]byte, error) {
	// Unmarshal YAML to a generic interface
	var data interface{}
	err := yaml.Unmarshal([]byte(yamlData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	// Marshal to JSON
	return json.MarshalIndent(data, "", "  ")
}
