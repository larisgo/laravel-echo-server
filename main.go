package main

import (
	"github.com/larisgo/laravel-echo-server/cli"
	"os"
)

func main() {
	args, err := cli.ParseArgs()
	if err != nil {
		// log.Error(err)
		cli.Usage()
		os.Exit(0)
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
	case "default":
		fallthrough
	default:
		cli.Usage()
		os.Exit(0)
	}
}
