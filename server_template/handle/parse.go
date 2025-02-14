package handle

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func ParseRequest(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	if len(body) == 0 {
		return true
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	err = dec.Decode(req)
	if err != nil {
		AbortWithErrCode(w, http.StatusBadRequest, err)
		return false
	}

	return true
}
