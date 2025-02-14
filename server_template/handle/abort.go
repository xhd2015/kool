package handle

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func AbortWithErr(w http.ResponseWriter, err error) {
	AbortWithErrCode(w, http.StatusInternalServerError, err)
}

func AbortWithErrCode(w http.ResponseWriter, code int, err error) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf(`{"code":%d, "msg":%q}`, code, err.Error())))
}

func ResponseJSON(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		AbortWithErr(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
