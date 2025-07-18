package jsontool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xhd2015/kool/pkgs/terminal"
)

func HandleSample(args []string) error {
	var match string
	var data string
	if len(args) > 0 {
		match = args[0]
		args = args[1:]

		var err error
		data, err = terminal.ReadOrTerminalData(args)
		if err != nil {
			return err
		}
	}

	return sampleJSON([]byte(data), match)
}

func sampleJSON(data []byte, match string) error {
	v, err := decodeJSON(data)
	if err != nil {
		return err
	}

	_, sample := traverseSample(v, match)
	sampleData, err := prettyJSON(sample)
	if err != nil {
		return err
	}
	fmt.Println(string(sampleData))
	return nil
}
func prettyJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func traverseSample(v interface{}, match string) (bool, interface{}) {
	if v == nil {
		return match == "", nil
	}
	switch v := v.(type) {
	case []interface{}:
		var newV []interface{}
		var hasMatch bool
		for _, e := range v {
			ok, x := traverseSample(e, match)
			if !ok {
				continue
			}
			hasMatch = true
			newV = append(newV, x)
			if match == "" && len(newV) >= 2 {
				break
			}
		}
		return hasMatch, newV
	case map[string]interface{}:
		var hasAnyMatch bool
		newMap := make(map[string]interface{}, len(v))
		for k, e := range v {
			ok, x := traverseSample(e, match)
			if ok {
				hasAnyMatch = true
			}
			newMap[k] = x
		}
		return hasAnyMatch, newMap
	case string:
		hasMatch := match == "" || strings.Contains(v, match)
		return hasMatch, v
	default:
		return match == "", v
	}
}

func decodeJSON(data []byte) (interface{}, error) {
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func enclosedBy(data []byte, pairs [][2]byte) bool {
	if len(data) < 2 {
		return false
	}
	i := 0
	n := len(data)
	for ; i < n && isSpace(data[i]); i++ {
	}
	if i >= n {
		return false
	}
	var match [2]byte
	var found bool
	for _, pair := range pairs {
		if data[i] == pair[0] {
			match = pair
			found = true
			break
		}
	}
	if !found {
		return false
	}
	j := n - 1
	for ; j > i && isSpace(data[j]); j-- {
	}
	if j <= i {
		return false
	}
	return data[j] == match[1]
}
func isSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}
