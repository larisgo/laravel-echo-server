package main

import (
	"os"

	"github.com/larisgo/laravel-echo-server/cli"
	"github.com/zishang520/engine.io/utils"
)

func main() {
	args, err := cli.ParseArgs()
	if err != nil {
		utils.Log().Fatal("%v", err)
		// os.Exit(0)
	}
	cmd := cli.NewCli()
	switch args.Command {
	case "start":
		cmd.Start(args)
	case "stop":
		cmd.Stop(args)
	case "init", "configure":
		cmd.Configure(args)
	case "client:add":
		cmd.ClientAdd(args)
	case "client:remove":
		cmd.ClientRemove(args)
	default:
		cli.Usage()
		os.Exit(0)
	}
}
