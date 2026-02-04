package cookie

import "strings"

func GetToken(cookie string) string {
	return Get(cookie, "token")
}

func Get(cookie string, key string) string {
	n := len(key) + len("=")
	idx := strings.Index(cookie, key+"=")
	if idx == -1 {
		return ""
	}
	base := idx + n
	keyEnd := strings.Index(cookie[base:], ";")
	if keyEnd == -1 {
		return cookie[base:]
	}
	keyEnd += base
	return cookie[base:keyEnd]
}
