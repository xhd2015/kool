package model

import (
	"strconv"
	"strings"
)

type OptionalNumber string

func (c *OptionalNumber) UnmarshalJSON(data []byte) error {
	*c = OptionalNumber(data)
	return nil
}

func (c OptionalNumber) Int64() (int64, error) {
	if len(c) == 0 {
		return 0, nil
	}
	var s string
	if strings.HasPrefix(string(c), "\"") {
		var err error
		s, err = strconv.Unquote(string(c))
		if err != nil {
			return 0, err
		}
		if s == "" {
			return 0, nil
		}
	} else {
		s = string(c)
		if s == "null" {
			return 0, nil
		}
	}
	return strconv.ParseInt(s, 10, 64)
}
