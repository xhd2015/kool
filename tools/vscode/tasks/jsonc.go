package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// StripJSONC removes // line comments and trailing commas so data can be
// parsed as standard JSON (VS Code tasks.json style).
func StripJSONC(src []byte) []byte {
	var out bytes.Buffer
	inString := false
	escaped := false
	n := len(src)
	for i := 0; i < n; i++ {
		c := src[i]
		if inString {
			out.WriteByte(c)
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		// not in string
		if c == '"' {
			inString = true
			out.WriteByte(c)
			continue
		}
		// // line comment
		if c == '/' && i+1 < n && src[i+1] == '/' {
			i += 2
			for i < n && src[i] != '\n' && src[i] != '\r' {
				i++
			}
			if i < n {
				// keep newline
				out.WriteByte(src[i])
			}
			continue
		}
		// /* block comment */
		if c == '/' && i+1 < n && src[i+1] == '*' {
			i += 2
			for i+1 < n && !(src[i] == '*' && src[i+1] == '/') {
				i++
			}
			i++ // skip closing /
			continue
		}
		// trailing comma before } or ]
		if c == ',' {
			// peek next non-space non-comment
			j := i + 1
			for j < n {
				// skip whitespace
				if src[j] == ' ' || src[j] == '\t' || src[j] == '\n' || src[j] == '\r' {
					j++
					continue
				}
				// skip // comment
				if src[j] == '/' && j+1 < n && src[j+1] == '/' {
					j += 2
					for j < n && src[j] != '\n' && src[j] != '\r' {
						j++
					}
					continue
				}
				// skip /* comment */
				if src[j] == '/' && j+1 < n && src[j+1] == '*' {
					j += 2
					for j+1 < n && !(src[j] == '*' && src[j+1] == '/') {
						j++
					}
					j += 2
					continue
				}
				break
			}
			if j < n && (src[j] == '}' || src[j] == ']') {
				// drop trailing comma
				continue
			}
		}
		out.WriteByte(c)
	}
	return out.Bytes()
}

// UnmarshalJSONC parses JSONC into v.
func UnmarshalJSONC(data []byte, v interface{}) error {
	clean := StripJSONC(data)
	if err := json.Unmarshal(clean, v); err != nil {
		return fmt.Errorf("invalid tasks.json: parse error: %w", err)
	}
	return nil
}

// dependsOn field may be a string or []string.
func parseDependsOn(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	s := strings.TrimSpace(string(raw))
	if strings.HasPrefix(s, "[") {
		var list []string
		if err := json.Unmarshal(raw, &list); err != nil {
			return nil
		}
		return list
	}
	var one string
	if err := json.Unmarshal(raw, &one); err != nil {
		return nil
	}
	if one == "" {
		return nil
	}
	return []string{one}
}
