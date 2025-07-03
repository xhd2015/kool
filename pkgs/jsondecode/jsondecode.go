package jsondecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func SafeUnmarshalJson(data []byte) (interface{}, error) {
	var v interface{}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err := dec.Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("cannot compress body, invalid json: %v", err)
	}

	var noMore interface{}
	noMoreErr := dec.Decode(&noMore)
	if noMoreErr != nil {
		if noMoreErr != io.EOF {
			return nil, fmt.Errorf("invalid json: %v", noMoreErr)
		}
	} else {
		err := json.Unmarshal([]byte(data), &noMore)
		if err != nil {
			return nil, fmt.Errorf("invalid json: %v", err)
		} else {
			return nil, fmt.Errorf("invalid json: multiple json object found")
		}
	}
	return v, nil
}
