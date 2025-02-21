package handle

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func ParseRequest(w http.ResponseWriter, r *http.Request, req interface{}) error {
	var body []byte
	if r.Method == http.MethodGet {
		// query
		m := make(map[string]string)
		for k, v := range r.URL.Query() {
			if len(v) == 0 {
				m[k] = ""
			} else {
				// take last
				m[k] = v[len(v)-1]
			}
		}
		var err error
		body, err = json.Marshal(m)
		if err != nil {
			return err
		}
	} else {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return err
		}
	}

	if len(body) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	err := dec.Decode(req)
	if err != nil {
		return err
	}

	return nil
}
