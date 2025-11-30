package bash

import (
	"fmt"

	"github.com/xhd2015/kool/tools/bash/client"
	"github.com/xhd2015/kool/tools/bash/history"
	"github.com/xhd2015/kool/tools/bash/server"
	"github.com/xhd2015/kool/tools/bash/web"
)

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command: history, server, server-exec")
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "history":
		return history.Handle(args)
	case "server":
		return server.Handle(args)
	case "server-exec":
		return client.Handle(args)
	case "web":
		return web.Handle(args)
	}

	return fmt.Errorf("unknown command: %s", cmd)
}
