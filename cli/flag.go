package cli

import (
	"flag"
	"fmt"
	"github.com/larisgo/laravel-echo-server/utils"
	"os"
	"os/exec"
	"path/filepath"
)

// const usage1 string = `Usage: laravel-echo-server [OPTIONS] <local port or address>
// Options:
// `

const usage2 string = `Advanced usage: %s [OPTIONS] <command> [command args] [...]
Commands:
  start [-config=laravel-echo-server.json] [-dir] [-force] [-dev]
        Starts the server.
  stop [-config=laravel-echo-server.json] [-dir]
        Stops the server.
  configure|init [-config=laravel-echo-server.json] [-dir]
        Creates a custom config file.
  client:add [-config=laravel-echo-server.json] [-dir] [id]
        Register a client that can make api requests.
  client:remove [-config=laravel-echo-server.json] [-dir] [id]
        Remove a registered client.
  help|h
        Print help
  version
        Print version
`

type Args struct {
	Config  string
	Dir     string
	Force   bool
	Dev     bool
	Command string
	Args    []string
}

var (
	StartFlag        = flag.NewFlagSet("start", flag.ExitOnError)
	StopFlag         = flag.NewFlagSet("stop", flag.ExitOnError)
	ConfigureFlag    = flag.NewFlagSet("configure", flag.ExitOnError)
	InitFlag         = flag.NewFlagSet("init", flag.ExitOnError)
	ClientAddFlag    = flag.NewFlagSet("client:add", flag.ExitOnError)
	ClientRemoveFlag = flag.NewFlagSet("client:remove", flag.ExitOnError)
)

func Filename() string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "laravel-echo-server"
	}
	return filepath.Base(file)
}

func Usage() {
	fmt.Fprintf(os.Stderr, fmt.Sprintf(usage2, Filename()))
}

func ParseArgs() (opts *Args, err error) {
	flag.Usage = Usage

	flag.Parse()

	var (
		StartFlagconfig = StartFlag.String("config", "laravel-echo-server.json", "The config file to use.")
		StartFlagdir    = StartFlag.String("dir", "", "The working directory to use.")
		StartFlagforce  = StartFlag.Bool("force", false, "If a server is already running, stop it.")
		StartFlagdev    = StartFlag.Bool("dev", false, "Run in dev mode.")

		StopFlagconfig = StopFlag.String("config", "laravel-echo-server.json", "The config file to use.")
		StopFlagdir    = StopFlag.String("dir", "", "The working directory to use.")

		InitFlagconfig = InitFlag.String("config", "laravel-echo-server.json", "The config file to use.")
		InitFlagdir    = InitFlag.String("dir", "", "The working directory to use.")

		ConfigureFlagconfig = ConfigureFlag.String("config", "laravel-echo-server.json", "The config file to use.")
		ConfigureFlagdir    = ConfigureFlag.String("dir", "", "The working directory to use.")

		ClientAddFlagconfig = ClientAddFlag.String("config", "laravel-echo-server.json", "The config file to use.")
		ClientAddFlagdir    = ClientAddFlag.String("dir", "", "The working directory to use.")

		ClientRemoveFlagconfig = ClientRemoveFlag.String("config", "laravel-echo-server.json", "The config file to use.")
		ClientRemoveFlagdir    = ClientRemoveFlag.String("dir", "", "The working directory to use.")
	)

	switch flag.Arg(0) {
	case "start":
		StartFlag.Parse(flag.Args()[1:])
		opts = &Args{
			Config: *StartFlagconfig,
			Dir:    *StartFlagdir,
			Force:  *StartFlagforce,
			Dev:    *StartFlagdev,
		}
		opts.Command = "start"
		opts.Args = StartFlag.Args()
	case "stop":
		StopFlag.Parse(flag.Args()[1:])
		opts = &Args{
			Config: *StopFlagconfig,
			Dir:    *StopFlagdir,
		}
		opts.Command = "stop"
		opts.Args = StopFlag.Args()
	case "init":
		InitFlag.Parse(flag.Args()[1:])
		opts = &Args{
			Config: *InitFlagconfig,
			Dir:    *InitFlagdir,
		}
		opts.Command = "init"
		opts.Args = InitFlag.Args()
	case "configure":
		ConfigureFlag.Parse(flag.Args()[1:])
		opts = &Args{
			Config: *ConfigureFlagconfig,
			Dir:    *ConfigureFlagdir,
		}
		opts.Command = "configure"
		opts.Args = ConfigureFlag.Args()
	case "client:add":
		ClientAddFlag.Parse(flag.Args()[1:])
		opts = &Args{
			Config: *ClientAddFlagconfig,
			Dir:    *ClientAddFlagdir,
		}
		opts.Command = "client:add"
		opts.Args = ClientAddFlag.Args()
	case "client:remove":
		ClientRemoveFlag.Parse(flag.Args()[1:])
		opts = &Args{
			Config: *ClientRemoveFlagconfig,
			Dir:    *ClientRemoveFlagdir,
		}
		opts.Command = "client:remove"
		opts.Args = ClientRemoveFlag.Args()
	case "version":
		fmt.Println(utils.VERSION)
		os.Exit(0)
	case "h", "help":
		flag.Usage()
		os.Exit(0)
	case "":
		err = fmt.Errorf("Please provide a valid command.")
		return
	default:
		if len(flag.Args()) > 1 {
			err = fmt.Errorf("Please provide a valid command, got %d: %v",
				len(flag.Args()),
				flag.Args())
			return
		}
		opts = &Args{}
		opts.Command = "default"
		opts.Args = flag.Args()
	}
	return
}
