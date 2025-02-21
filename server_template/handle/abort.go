package handle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func AbortWithErr(w http.ResponseWriter, err error) {
	AbortWithErrCode(w, http.StatusInternalServerError, err)
}

func AbortWithErrCode(w http.ResponseWriter, code int, err error) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, writeErr := w.Write([]byte(fmt.Sprintf(`{"code":%d, "msg":%q}`, code, err.Error())))
	if writeErr != nil {
		fmt.Fprintf(os.Stderr, "write error: %v", writeErr)
	}
}

func ResponseJSON(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		AbortWithErr(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write(jsonData)
	if writeErr != nil {
		fmt.Fprintf(os.Stderr, "write error: %v", writeErr)
	}
}
