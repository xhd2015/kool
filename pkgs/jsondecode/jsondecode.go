package jsondecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

func UnmarshalSafe(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	err := dec.Decode(&v)
	if err != nil {
		return fmt.Errorf("cannot compress body, invalid json: %v", err)
	}

	var noMore interface{}
	noMoreErr := dec.Decode(&noMore)
	if noMoreErr != nil {
		if noMoreErr != io.EOF {
			return fmt.Errorf("invalid json: %v", noMoreErr)
		}
	} else {
		err := json.Unmarshal([]byte(data), &noMore)
		if err != nil {
			return fmt.Errorf("invalid json: %v", err)
		} else {
			return fmt.Errorf("invalid json: multiple json object found")
		}
	}
	return nil
}

func UnmarshalSafeAny(data []byte) (interface{}, error) {
	var v interface{}

	err := UnmarshalSafe(data, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func SafeCompress(data []byte) ([]byte, error) {
	jsonObj, err := UnmarshalSafeAny(data)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func MustPrettyString(s string) string {
	if s == "" {
		return ""
	}
	jsonObj, err := UnmarshalSafeAny([]byte(s))
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func PrettyObjectOneLevel(data []byte) ([]byte, error) {
	var v map[string]json.RawMessage
	err := UnmarshalSafe(data, &v)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return []byte("null"), nil
	}
	keys := make([]string, 0, len(v))
	for key := range v {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var buf bytes.Buffer
	n := len(keys)
	fmt.Fprintf(&buf, "{\n")
	for i, key := range keys {
		fmt.Fprintf(&buf, "  %q: %s", key, string(v[key]))
		if i < n-1 {
			fmt.Fprintf(&buf, ",")
		}
		fmt.Fprintf(&buf, "\n")
	}
	fmt.Fprintf(&buf, "}")
	return buf.Bytes(), nil
}

func MustPrettyObjectOneLevelToString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	bytes, err := PrettyObjectOneLevel(data)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
