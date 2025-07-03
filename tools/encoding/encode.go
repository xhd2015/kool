package encoding

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/xhd2015/kool/pkgs/terminal"
	"github.com/xhd2015/less-gen/flags"
)

const encodeHelp = `
kool encode <algorithm> <data>

Supported algorithms:
- base64
- url
- hex
- ascii_hex      can be decoded via ` + "`echo -ne '\\x21...'`" + `
`

func HandleEncode(args []string) error {
	var verbose bool
	args, err := flags.String("--verbose", &verbose).
		Help("-h,--help", encodeHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires algorithm, usage: kool encode <algorithm> <data>, try kool encode --help")
	}

	alg := args[0]
	args = args[1:]

	if alg == "help" || alg == "--help" {
		fmt.Print(strings.TrimPrefix(encodeHelp, "\n"))
		return nil
	}

	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return fmt.Errorf("usage: kool encode <algorithm> <data>, try kool encode --help. %w", err)
	}

	switch alg {
	case "base64":
		fmt.Println(base64.StdEncoding.EncodeToString([]byte(data)))
		return nil
	case "url":
		fmt.Println(url.QueryEscape(data))
		return nil
	case "hex":
		fmt.Println(hex.EncodeToString([]byte(data)))
	case "ascii_hex":
		// i.e. !E -> \x21\x45
		bs := []byte(data)
		for i := 0; i < len(bs); i++ {
			fmt.Printf("\\x%02x", bs[i])
		}
		fmt.Println()
		return nil
	default:
		return fmt.Errorf("unknown algorithm: %s", alg)
	}
	return nil
}
