package encoding

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/xhd2015/kool/pkgs/terminal"
)

const decodeHelp = `
kool encode <algorithm> <data>

Supported algorithms:
- base64
- url
- hex
- ascii_hex

Examples:
  kool decode ascii_hex '\x41\x21'
`

func HandleDecode(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help") {
		fmt.Print(strings.TrimPrefix(decodeHelp, "\n"))
		return nil
	}

	var alg string
	if len(args) > 0 {
		switch args[0] {
		case "base64", "url", "hex", "ascii_hex":
			alg = args[0]
			args = args[1:]
		}
	}

	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return fmt.Errorf("usage: kool decode <data>, try kool decode --help. %w", err)
	}

	data = strings.TrimSuffix(data, "\n")
	data = strings.TrimSuffix(data, "\r")
	if data == "" {
		return fmt.Errorf("requires data")
	}

	switch alg {
	case "ascii_hex":
		decoded, err := decodeAsciiHex(data)
		if err != nil {
			return err
		}
		fmt.Println(decoded)
		return nil
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return err
		}
		fmt.Println(string(decoded))
		return nil
	case "url":
		decoded, err := url.QueryUnescape(data)
		if err != nil {
			return err
		}
		fmt.Println(decoded)
		return nil
	case "hex":
		decoded, err := hex.DecodeString(data)
		if err != nil {
			return err
		}
		fmt.Println(string(decoded))
		return nil
	default:
		// auto
		if decoded, err := decodeAsciiHex(data); err == nil {
			fmt.Println(decoded)
			return nil
		}

		// base64 decode
		if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
			fmt.Println(string(decoded))
			return nil
		}

		// try url first
		if unescaped, err := url.QueryUnescape(data); err == nil {
			fmt.Println(unescaped)
			return nil
		}

	}

	return fmt.Errorf("unable to decode data")
}

func decodeAsciiHex(input string) (string, error) {
	// Check if the input is well-formed (starts with \x and has valid length)
	if len(input) < 4 || input[0] != '\\' || input[1] != 'x' {
		return "", fmt.Errorf("invalid hex escape sequence")
	}

	var result strings.Builder
	// Process the string in steps of 4 characters (\xHH)
	for i := 0; i < len(input); i += 4 {
		// Ensure we have enough characters for \xHH
		if i+4 > len(input) || input[i] != '\\' || input[i+1] != 'x' {
			return "", fmt.Errorf("malformed hex escape sequence at position %d", i)
		}

		// Extract the two hex digits
		hex := input[i+2 : i+4]
		// Parse hex to integer
		value, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return "", fmt.Errorf("invalid hex value %s: %v", hex, err)
		}
		// Write the character to the result
		result.WriteRune(rune(value))
	}

	return result.String(), nil
}
