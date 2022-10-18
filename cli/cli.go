package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/larisgo/laravel-echo-server/echo"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/std"
	"github.com/larisgo/laravel-echo-server/types"
	_utils "github.com/larisgo/laravel-echo-server/utils"
	"github.com/zishang520/engine.io/utils"
)

type Cli struct {

	//Default configuration options.
	defaultOptions *options.Config

	//echo Server.
	echo *echo.EchoServer
}

// Create new CLI instance.
func NewCli() *Cli {
	c := &Cli{}
	c.echo = echo.NewEchoServer()
	c.defaultOptions = c.echo.DefaultOptions

	return c
}

// Create a configuration file.
func (c *Cli) Configure(args *Args) {
	defer func() {
		if err := recover(); err != nil {
			utils.Log().Fatal("%v", err)
		}
	}()

	if len(args.Dir) > 0 {
		if err := os.Chdir(args.Dir); err != nil {
			panic(err)
			return
		}
	}

	config, file, err := c.setupConfig(args.Config)
	if err != nil {
		panic(err)
		return
	}
	if err := c.saveConfig(config, file); err != nil {
		panic(err)
		return
	} else {
		utils.Log().Success("Configuration file saved. Run [%s start%s] to run server.", Filename(), (func() string {
			if file != "laravel-echo-server.json" {
				return " --config=" + file
			}
			return ""
		})())
	}
}

func (c *Cli) EnvIsNull(v string) bool {
	switch strings.ToLower(v) {
	case "null", "(null)":
		return true
	}
	return false
}

func (c *Cli) EnvIsEmpty(v string) bool {
	switch strings.ToLower(v) {
	case "", "empty", "(empty)":
		return true
	}
	return false
}

func (c *Cli) EnvToEmpty(v string) string {
	if c.EnvIsEmpty(v) {
		return ""
	}
	return v
}

func (c *Cli) EnvToBool(v string) bool {
	switch strings.ToLower(v) {
	case "true", "(true)":
		return true
	}
	return false
}

// Inject the .env vars into options if they exist.
func (c *Cli) resolveEnvFileOptions(config *options.Config) *options.Config {
	if err := godotenv.Load(); err != nil {
		return config
	}

	if SERVER_AUTH_HOST, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_AUTH_HOST"); ok && !c.EnvIsNull(SERVER_AUTH_HOST) {
		config.AuthHost = c.EnvToEmpty(SERVER_AUTH_HOST)
	}

	if SERVER_HOST, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_HOST"); ok && !c.EnvIsNull(SERVER_HOST) {
		config.Host = c.EnvToEmpty(SERVER_HOST)
	}

	if SERVER_PORT, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_PORT"); ok && !c.EnvIsNull(SERVER_PORT) {
		config.Port = c.EnvToEmpty(SERVER_PORT)
	}

	if SERVER_DEBUG, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_DEBUG"); ok && !c.EnvIsNull(SERVER_DEBUG) {
		config.DevMode = c.EnvToBool(SERVER_DEBUG)
	}

	if SERVER_REDIS_HOST, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_HOST"); ok && !c.EnvIsNull(SERVER_REDIS_HOST) {
		config.DatabaseConfig.Redis.Host = c.EnvToEmpty(SERVER_REDIS_HOST)
	}

	if SERVER_REDIS_PORT, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_PORT"); ok && !c.EnvIsNull(SERVER_REDIS_PORT) {
		config.DatabaseConfig.Redis.Port = c.EnvToEmpty(SERVER_REDIS_PORT)
	}

	if SERVER_REDIS_USERNAME, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_USERNAME"); ok && !c.EnvIsNull(SERVER_REDIS_USERNAME) {
		config.DatabaseConfig.Redis.Username = c.EnvToEmpty(SERVER_REDIS_USERNAME)
	}

	if SERVER_REDIS_PASSWORD, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_PASSWORD"); ok && !c.EnvIsNull(SERVER_REDIS_PASSWORD) {
		config.DatabaseConfig.Redis.Password = c.EnvToEmpty(SERVER_REDIS_PASSWORD)
	}

	if SERVER_REDIS_KEYPREFIX, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_KEYPREFIX"); ok && !c.EnvIsNull(SERVER_REDIS_KEYPREFIX) {
		config.DatabaseConfig.Redis.KeyPrefix = c.EnvToEmpty(SERVER_REDIS_KEYPREFIX)
	}

	if SERVER_PROTO, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_PROTO"); ok && !c.EnvIsNull(SERVER_PROTO) {
		config.Protocol = c.EnvToEmpty(SERVER_PROTO)
	}

	if SERVER_SSL_CERT, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_SSL_CERT"); ok && !c.EnvIsNull(SERVER_SSL_CERT) {
		config.SslCertPath = c.EnvToEmpty(SERVER_SSL_CERT)
	}

	if SERVER_SSL_KEY, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_SSL_KEY"); ok && !c.EnvIsNull(SERVER_SSL_KEY) {
		config.SslKeyPath = c.EnvToEmpty(SERVER_SSL_KEY)
	}

	return config
}

// Setup configuration with questions.
func (c *Cli) setupConfig(defaultFile string) (*options.Config, string, error) {
	config := c.defaultOptions

	input := std.NewDefaultInput(os.Stdin, std.NewDefaultOutput(os.Stdout))

	if config.DevMode = input.Confirm("Do you want to run c server in development mode?", false); config.DevMode {
		utils.Log().Success("Yes")
	} else {
		utils.Log().Success("No")
	}

	config.Port = input.Ask("Which port would you like to serve from:", func(v string) error {
		if (len(v) == 0) || regexp.MustCompile(`^([1-9]|[1-9]\d{1,3}|[1-6][0-5][0-5][0-3][0-5])$`).MatchString(v) {
			return nil
		} else {
			return errors.New("Port numbers range from 1 to 65535")
		}
	}, "6001")
	utils.Log().Success(config.Port)

	config.Database = input.Choose("Which database would you like to use to store presence channel members?", map[string]string{
		"redis":  "Use redis to store.",
		"sqlite": "Use sqlite to store.",
	}, "sqlite")
	utils.Log().Success(config.Database)

	config.AuthHost = input.Ask("Enter the host of your Laravel authentication server:", func(_ string) error {
		return nil
	}, "http://localhost")
	utils.Log().Success("%v", config.AuthHost)

	config.Protocol = input.Choose("Will you be serving on http or https?", map[string]string{
		"http":  "Run the service using http.",
		"https": "Run the service using https.",
	}, "http")
	utils.Log().Success(config.Protocol)

	if config.Protocol == "https" {
		config.SslCertPath = input.Ask("Enter the path to your SSL cert file:", func(v string) error {
			if len(v) > 0 {
				return nil
			} else {
				return errors.New("Please enter ssl Cert Path.")
			}
		})
		utils.Log().Success(config.SslCertPath)

		config.SslKeyPath = input.Ask("Enter the path to your SSL key file:", func(v string) error {
			if len(v) > 0 {
				return nil
			} else {
				return errors.New("Please enter ssl Key Path.")
			}
		})
		utils.Log().Success(config.SslKeyPath)
	}

	if input.Confirm("Do you want to generate a client ID/Key for HTTP API?", false) {
		utils.Log().Success("Yes")
		client := options.Client{}
		if aid, err := c.createAppId(); err != nil {
			return nil, "", err
		} else {
			client.AppId = aid
		}
		if k, err := c.createApiKey(); err != nil {
			return nil, "", err
		} else {
			client.Key = k
		}
		if config.Clients == nil {
			config.Clients = []options.Client{}
		}
		config.Clients = append(config.Clients, client)

		utils.Log().Info("appId: " + client.AppId)
		utils.Log().Info("key: " + client.Key)
	} else {
		utils.Log().Success("No")
	}

	if config.ApiOriginAllow.AllowCors = input.Confirm("Do you want to setup cross domain access to the API?", false); config.ApiOriginAllow.AllowCors {
		utils.Log().Success("Yes")
	} else {
		utils.Log().Success("No")
	}

	if config.ApiOriginAllow.AllowCors {
		config.ApiOriginAllow.AllowOrigin = input.Ask("Specify the URI that may access the API:", func(_ string) error {
			return nil
		}, "http://localhost:80")
		utils.Log().Success(config.ApiOriginAllow.AllowOrigin)

		config.ApiOriginAllow.AllowMethods = input.Ask("Enter the HTTP methods that are allowed for CORS:", func(_ string) error {
			return nil
		}, "GET, POST")
		utils.Log().Success(config.ApiOriginAllow.AllowMethods)

		config.ApiOriginAllow.AllowHeaders = input.Ask("Enter the HTTP headers that are allowed for CORS:", func(_ string) error {
			return nil
		}, "Origin, Content-Type, X-Auth-Token, X-Requested-With, Accept, Authorization, X-CSRF-TOKEN, X-Socket-Id")
		utils.Log().Success(config.ApiOriginAllow.AllowHeaders)
	}

	file := input.Ask("What do you want c config to be saved as:", func(_ string) error {
		return nil
	}, defaultFile)
	utils.Log().Success(file)

	return config, file, nil
}

// Save configuration file.
func (c *Cli) saveConfig(config *options.Config, file string) (err error) {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
	}()

	if err := f.Truncate(0); err != nil {
		return err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	return
}

// Start the Laravel Echo server.
func (c *Cli) Start(args *Args) {
	var lockFile string

	defer func() {
		if err := recover(); err != nil {
			if lockFile != "" && _utils.Exists(lockFile) {
				os.Remove(lockFile)
			}
			utils.Log().Fatal("%v", err)
		}
	}()

	if len(args.Dir) > 0 {
		if err := os.Chdir(args.Dir); err != nil {
			panic(err)
			return
		}
	}

	configFile, err := c.getConfigFile(args.Config, args.Dir)
	if err != nil {
		panic(err)
		return
	}
	if !_utils.Exists(configFile) {
		panic(errors.New(`Error: The config file [` + args.Config + `] cound not be found.`))
		return
	}
	config, err := c.readConfigFile(configFile)
	if err != nil {
		panic(err)
		return
	}
	if args.Dev {
		config.DevMode = true
	}

	lockFile = filepath.Clean(path.Join(filepath.Dir(configFile), strings.TrimSuffix(args.Config, ".json")+".lock"))

	if _utils.Exists(lockFile) {
		lockProcess, err := ioutil.ReadFile(lockFile)
		if err != nil {
			panic(err)
			return
		}
		var processInfo *types.PocessLockData
		if err := json.Unmarshal(lockProcess, &processInfo); err == nil && processInfo != nil {
			// kill proccess
			if process, err := os.FindProcess(processInfo.Process); err == nil {
				if args.Force {
					if err := process.Signal(os.Kill); err != nil {
						utils.Log().Error("%v", err)
						return
					} else {
						utils.Log().Warning(`Warning: Closing process %d because you used the "--force" option.`, processInfo.Process)
					}
				} else {
					utils.Log().Error(`Error: There is already a server running! Use the option "--force" to stop it and start another one.)`)
					return
				}
			}
		}
	}

	exit := make(chan struct{}, 1)
	SignalC := make(chan os.Signal, 1)

	signal.Notify(SignalC, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range SignalC {
			switch s {
			case os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				close(exit)
				return
			}
		}
	}()

	if err := c.echo.Run(config); err != nil {
		panic(err)
		return
	}

	p_id, err := json.MarshalIndent(&types.PocessLockData{Process: os.Getpid()}, "", "    ")
	if err != nil {
		panic(err)
		return
	}
	if err := ioutil.WriteFile(lockFile, p_id, 0644); err != nil {
		panic(err)
		return
	}

	<-exit

	c.echo.Stop()
	os.Remove(lockFile)
	os.Exit(0)
}

// Stop the Laravel Echo server.
func (c *Cli) Stop(args *Args) {
	defer func() {
		if err := recover(); err != nil {
			utils.Log().Fatal("%v", err)
		}
	}()

	if len(args.Dir) > 0 {
		if err := os.Chdir(args.Dir); err != nil {
			panic(err)
			return
		}
	}

	configFile, err := c.getConfigFile(args.Config, args.Dir)
	if err != nil {
		panic(err)
		return
	}
	lockFile := filepath.Clean(path.Join(filepath.Dir(configFile), strings.TrimSuffix(args.Config, ".json")+".lock"))

	if !_utils.Exists(lockFile) {
		panic(errors.New(`Could not find any lock file.`))
		return
	}

	lockProcess, err := ioutil.ReadFile(lockFile)
	if err != nil {
		panic(err)
		return
	}
	var processInfo *types.PocessLockData
	if err := json.Unmarshal(lockProcess, &processInfo); err != nil {
		panic(err)
		return
	}
	process, err := os.FindProcess(processInfo.Process)
	if err != nil {
		panic(err)
		return
	}
	if err := process.Signal(quitSignal()); err != nil {
		panic(err)
		return
	}
	utils.Log().Success(`Closed the running server.`)
}

// Create an app key for server.
func (c *Cli) getRandomString(length int) (string, error) {
	data := make([]byte, length)
	n, err := rand.Read(data)
	if err != nil {
		return "", err
	}
	if n != length {
		return "", errors.New(`RandomString Length error`)
	}
	return hex.EncodeToString(data), nil
}

// Create an api key for the HTTP API.
func (c *Cli) createApiKey() (string, error) {
	return c.getRandomString(16)
}

// Create an api key for the HTTP API.
func (c *Cli) createAppId() (string, error) {
	return c.getRandomString(8)
}

// Add a registered referrer.
func (c *Cli) ClientAdd(args *Args) {
	defer func() {
		if err := recover(); err != nil {
			utils.Log().Fatal("%v", err)
		}
	}()

	if len(args.Dir) > 0 {
		if err := os.Chdir(args.Dir); err != nil {
			panic(err)
			return
		}
	}

	configFile, err := c.getConfigFile(args.Config, args.Dir)
	if err != nil {
		panic(err)
		return
	}
	if !_utils.Exists(configFile) {
		panic(errors.New(`Error: The config file [` + args.Config + `] cound not be found.`))
		return
	}
	config, err := c.readConfigFile(configFile)
	if err != nil {
		panic(err)
		return
	}
	appId := ""
	if len(args.Args) > 0 {
		appId = args.Args[0]
	} else {
		if aid, err := c.createAppId(); err != nil {
			panic(err)
		} else {
			appId = aid
		}
	}
	if appId == "" {
		panic(errors.New("appId is empty."))
		return
	}
	if config.Clients == nil {
		config.Clients = []options.Client{}
	}
	var client options.Client
	var index int
	has_client := false
	for index, client = range config.Clients {
		if client.AppId == appId {
			has_client = true
			if k, err := c.createApiKey(); err != nil {
				panic(err)
				return
			} else {
				client.Key = k
			}
			config.Clients[index] = client
			utils.Log().Info("API Client updated!")
			break
		}
	}

	if !has_client {
		client = options.Client{
			AppId: appId,
		}
		if k, err := c.createApiKey(); err != nil {
			panic(err)
			return
		} else {
			client.Key = k
		}
		config.Clients = append(config.Clients, client)
		utils.Log().Info("API Client added!")
	}
	utils.Log().Info("appId: " + client.AppId)
	utils.Log().Info("key: " + client.Key)

	if err := c.saveConfig(config, args.Config); err != nil {
		panic(err)
		return
	}
}

// Remove a registered referrer.
func (c *Cli) ClientRemove(args *Args) {
	defer func() {
		if err := recover(); err != nil {
			utils.Log().Fatal("%v", err)
		}
	}()

	if len(args.Dir) > 0 {
		if err := os.Chdir(args.Dir); err != nil {
			panic(err)
			return
		}
	}

	configFile, err := c.getConfigFile(args.Config, args.Dir)
	if err != nil {
		panic(err)
		return
	}
	if !_utils.Exists(configFile) {
		panic(errors.New(`Error: The config file [` + args.Config + `] cound not be found.`))
		return
	}
	config, err := c.readConfigFile(configFile)
	if err != nil {
		panic(err)
		return
	}
	appId := ""
	if len(args.Args) > 0 {
		appId = args.Args[0]
	} else {
		if aid, err := c.createAppId(); err != nil {
			panic(err)
		} else {
			appId = aid
		}
	}
	if appId == "" {
		panic(errors.New("appId is empty."))
		return
	}
	if config.Clients == nil {
		config.Clients = []options.Client{}
	}
	// let index ;
	_index := 0
	for index, client := range config.Clients {
		if client.AppId == appId {
			_index = index
			break
		}
	}
	if (_index + 1) < len(config.Clients) {
		config.Clients = append(config.Clients[:_index], config.Clients[_index+1:]...)
	} else {
		config.Clients = config.Clients[:_index]
	}

	utils.Log().Info("Client removed: " + appId)

	if err := c.saveConfig(config, args.Config); err != nil {
		panic(err)
		return
	}
}

// Gets the config file with the provided args
func (c *Cli) getConfigFile(file string, dir string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	filePath := filepath.Clean(path.Join(dir, file))
	if filepath.IsAbs(filePath) {
		return filePath, nil
	}

	return filepath.Clean(path.Join(cwd, filePath)), nil
}

// Tries to read a config file
func (c *Cli) readConfigFile(file string) (*options.Config, error) {
	var data *options.Config

	bytes_data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes_data, &data); err != nil {
		return nil, err
	}

	return c.resolveEnvFileOptions(data), nil
}
