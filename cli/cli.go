package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/larisgo/laravel-echo-server/echo"
	"github.com/larisgo/laravel-echo-server/errors"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/tcnksm/go-input"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Cli struct {
	/**
	 * Default configuration options.
	 */
	defaultOptions options.Config

	/**
	 * echo Server.
	 */
	echo *echo.EchoServer
}

/**
 * Create new CLI instance.
 */
func NewCli() *Cli {
	this := &Cli{}
	this.echo = echo.NewEchoServer()
	this.defaultOptions = this.echo.DefaultOptions

	return this
}

/**
 * Create a configuration file.
 */
func (this *Cli) Configure(args *Args) {
	config, file := this.setupConfig(args.Config)
	if err := this.saveConfig(config, file); err != nil {
		log.Fatal(err)
	} else {
		log.Success(fmt.Sprintf("Configuration file saved. Run [%s start%s] to run server.", Filename(), (func() string {
			if file != "laravel-echo-server.json" {
				return fmt.Sprintf(" --config=%s", file)
			}
			return ""
		})()))
	}
}

func (this *Cli) EnvIsNull(v string) bool {
	switch strings.ToLower(v) {
	case "null", "(null)":
		return true
	}
	return false
}

func (this *Cli) EnvIsEmpty(v string) bool {
	switch strings.ToLower(v) {
	case "", "empty", "(empty)":
		return true
	}
	return false
}

func (this *Cli) EnvToEmpty(v string) string {
	if this.EnvIsEmpty(v) {
		return ""
	}
	return v
}

func (this *Cli) EnvToBool(v string) bool {
	switch strings.ToLower(v) {
	case "true", "(true)":
		return true
	}
	return false
}

/**
 * Inject the .env vars into options if they exist.
 */
func (this *Cli) resolveEnvFileOptions(config options.Config, args *Args) options.Config {
	if err := godotenv.Load(this.getConfigFile(".env", args.Dir)); err != nil {
		// log.Fatal("Error loading .env file")
		return config
	}

	if SERVER_AUTH_HOST, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_AUTH_HOST"); ok && !this.EnvIsNull(SERVER_AUTH_HOST) {
		config.AuthHost = this.EnvToEmpty(SERVER_AUTH_HOST)
	}

	if SERVER_HOST, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_HOST"); ok && !this.EnvIsNull(SERVER_HOST) {
		config.Host = this.EnvToEmpty(SERVER_HOST)
	}

	if SERVER_PORT, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_PORT"); ok && !this.EnvIsNull(SERVER_PORT) {
		config.Port = this.EnvToEmpty(SERVER_PORT)
	}

	if SERVER_DEBUG, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_DEBUG"); ok && !this.EnvIsNull(SERVER_DEBUG) {
		config.DevMode = this.EnvToBool(SERVER_DEBUG)
	}

	if SERVER_REDIS_HOST, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_HOST"); ok && !this.EnvIsNull(SERVER_REDIS_HOST) {
		config.DatabaseConfig.Redis.Host = this.EnvToEmpty(SERVER_REDIS_HOST)
	}

	if SERVER_REDIS_PORT, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_PORT"); ok && !this.EnvIsNull(SERVER_REDIS_PORT) {
		config.DatabaseConfig.Redis.Port = this.EnvToEmpty(SERVER_REDIS_PORT)
	}

	if SERVER_REDIS_PASSWORD, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_REDIS_PASSWORD"); ok && !this.EnvIsNull(SERVER_REDIS_PASSWORD) {
		config.DatabaseConfig.Redis.Password = this.EnvToEmpty(SERVER_REDIS_PASSWORD)
	}

	if SERVER_PREFIX, ok := os.LookupEnv("LARAVEL_ECHO_SERVER_PREFIX"); ok && !this.EnvIsNull(SERVER_PREFIX) {
		config.DatabaseConfig.Prefix = this.EnvToEmpty(SERVER_PREFIX)
	}

	return config
}

/**
 * Setup configuration with questions.
 */
func (this *Cli) setupConfig(defaultFile string) (options.Config, string) {
	config := this.defaultOptions

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	if devMode, err := ui.Select("Do you want to run this server in development mode?", []string{"true", "false"}, &input.Options{
		Default: "false",
		Loop:    true,
	}); err != nil {
		log.Fatal(err)
	} else {
		config.DevMode = map[string]bool{
			"true":  true,
			"false": false,
		}[devMode]
	}
	if port, err := ui.Ask("Which port would you like to serve from?", &input.Options{
		Default: "6001",
		Loop:    true,
	}); err != nil {
		log.Fatal(err)
	} else {
		config.Port = port
	}

	if database, err := ui.Select("Which database would you like to use to store presence channel members?", []string{"redis", "sqlite"}, &input.Options{
		Default: "sqlite",
		Loop:    true,
	}); err != nil {
		log.Fatal(err)
	} else {
		config.Database = database
	}

	if authHost, err := ui.Ask("Enter the host of your Laravel authentication server.", &input.Options{
		Default: "http://localhost",
		Loop:    true,
	}); err != nil {
		log.Fatal(err)
	} else {
		config.AuthHost = authHost
	}

	protocol, err := ui.Select("Will you be serving on http or https?", []string{"http", "https"}, &input.Options{
		Default: "http",
		Loop:    true,
	})
	if err != nil {
		log.Fatal(err)
	} else {
		config.Protocol = protocol
	}

	if protocol == "https" {
		if sslCertPath, err := ui.Ask("Enter the path to your SSL cert file.", &input.Options{
			Required: true,
			Loop:     true,
		}); err != nil {
			log.Fatal(err)
		} else {
			config.SslCertPath = sslCertPath
		}

		if sslKeyPath, err := ui.Ask("Enter the path to your SSL key file.", &input.Options{
			Required: true,
			Loop:     true,
		}); err != nil {
			log.Fatal(err)
		} else {
			config.SslKeyPath = sslKeyPath
		}
	}

	if addClient, err := ui.Select("Do you want to generate a client ID/Key for HTTP API?", []string{"true", "false"}, &input.Options{
		Default: "false",
		Loop:    true,
	}); err != nil {
		log.Fatal(err)
	} else {
		if map[string]bool{"true": true, "false": false}[addClient] {
			client := options.Client{
				AppId: this.createAppId(),
				Key:   this.createApiKey(),
			}
			if len(config.Clients) == 0 {
				config.Clients = []options.Client{}
			}
			config.Clients = append(config.Clients, client)
			// log.Info(fmt.Sprintf("appId: %s", client.AppId))
			// log.Info(fmt.Sprintf("key: %s", client.Key))
		}
	}

	if corsAllow, err := ui.Select("Do you want to setup cross domain access to the API?", []string{"true", "false"}, &input.Options{
		Default: "false",
		Loop:    true,
	}); err != nil {
		log.Fatal(err)
	} else {
		config.ApiOriginAllow.AllowCors = map[string]bool{
			"true":  true,
			"false": false,
		}[corsAllow]
	}

	if config.ApiOriginAllow.AllowCors {
		if allowOrigin, err := ui.Ask("Specify the URI that may access the API:", &input.Options{
			Default: "http://localhost:80",
			Loop:    true,
		}); err != nil {
			log.Fatal(err)
		} else {
			config.ApiOriginAllow.AllowOrigin = allowOrigin
		}
		if allowMethods, err := ui.Ask("Enter the HTTP methods that are allowed for CORS:", &input.Options{
			Default: "GET, POST",
			Loop:    true,
		}); err != nil {
			log.Fatal(err)
		} else {
			config.ApiOriginAllow.AllowMethods = allowMethods
		}
		if allowHeaders, err := ui.Ask("Enter the HTTP headers that are allowed for CORS:", &input.Options{
			Default: "Origin, Content-Type, X-Auth-Token, X-Requested-With, Accept, Authorization, X-CSRF-TOKEN, X-Socket-Id",
			Loop:    true,
		}); err != nil {
			log.Fatal(err)
		} else {
			config.ApiOriginAllow.AllowHeaders = allowHeaders
		}
	}
	file, err := ui.Ask("What do you want this config to be saved as?", &input.Options{
		Default: defaultFile,
		Loop:    true,
	})
	if err != nil {
		log.Fatal(err)
	}
	return config, file
}

/**
 * Save configuration file.
 */
func (this *Cli) saveConfig(config options.Config, file string) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Error(err)
			// return err
		}
	}()

	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func (this *Cli) FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (this *Cli) Exit(lockFile string) {
	os.Remove(lockFile)
	os.Exit(0)
}

/**
 * Start the Laravel Echo server.
 */
func (this *Cli) Start(args *Args) {
	configFile := this.getConfigFile(args.Config, args.Dir)
	if !this.FileExists(configFile) {
		log.Fatal(fmt.Sprintf(`Error: The config file [%s] cound not be found.`, args.Config))
	}
	config := this.readConfigFile(configFile, args)
	if args.Dev {
		config.DevMode = true
	}

	lockFile := filepath.Clean(path.Join(filepath.Dir(configFile), fmt.Sprintf("%s%s", strings.TrimSuffix(args.Config, ".json"), ".lock")))
	if this.FileExists(lockFile) {
		lockProcess, err := ioutil.ReadFile(lockFile)
		if err != nil {
			log.Fatal(err)
		}
		processInfo := types.PocessLockData{}
		if err := json.Unmarshal(lockProcess, &processInfo); err == nil {
			// kill proccess
			if process, err := os.FindProcess(processInfo.Process); err == nil {
				if args.Force {
					if err := process.Signal(syscall.SIGTERM); err != nil {
						log.Error(err)
					} else {
						log.Warning(fmt.Sprintf(`Warning: Closing process %d because you used the "--force" option.`, processInfo.Process))
					}
				} else {
					log.Fatal(`Error: There is already a server running! Use the option "--force" to stop it and start another one.`)
				}
			}
		}
	}

	pid_t, err := json.MarshalIndent(types.PocessLockData{Process: os.Getpid()}, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(lockFile, pid_t, 0644); err != nil {
		log.Fatal(err)
	}

	SignalC := make(chan os.Signal)

	signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range SignalC {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				this.Exit(lockFile)
			}
		}
	}()

	this.echo.Run(config)

	for {
		time.Sleep(time.Second)
	}
}

/**
 * Stop the Laravel Echo server.
 */
func (this *Cli) Stop(args *Args) {
	configFile := this.getConfigFile(args.Config, args.Dir)
	lockFile := filepath.Clean(path.Join(filepath.Dir(configFile), fmt.Sprintf("%s%s", strings.TrimSuffix(args.Config, ".json"), ".lock")))

	if this.FileExists(lockFile) {
		lockProcess, err := ioutil.ReadFile(lockFile)
		if err != nil {
			log.Fatal(err)
		}
		processInfo := types.PocessLockData{}
		if err := json.Unmarshal(lockProcess, &processInfo); err != nil {
			log.Fatal(err)
		}
		// kill proccess
		if process, err := os.FindProcess(processInfo.Process); err != nil {
			log.Error(`No running servers to close.`)
		} else {
			os.Remove(lockFile)
			if err := process.Signal(syscall.SIGTERM); err != nil {
				log.Fatal(err)
			}
			log.Success(`Closed the running server.`)
		}
	} else {
		log.Error(`Error: Could not find any lock file.`)
	}
}

/**
 * Create an app key for server.
 */
func (this *Cli) getRandomString(bytes int) string {
	data := make([]byte, bytes)
	n, err := rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	if n != bytes {
		log.Fatal(errors.NewError(`RandomString Length error`))
	}
	return hex.EncodeToString(data)
}

/**
 * Create an api key for the HTTP API.
 */
func (this *Cli) createApiKey() string {
	return this.getRandomString(16)
}

/**
 * Create an api key for the HTTP API.
 */
func (this *Cli) createAppId() string {
	return this.getRandomString(8)
}

/**
 * Add a registered referrer.
 */
func (this *Cli) ClientAdd(args *Args) {
	configFile := this.getConfigFile(args.Config, args.Dir)
	if !this.FileExists(configFile) {
		log.Fatal(fmt.Sprintf(`Error: The config file [%s] cound not be found.`, args.Config))
	}
	config := this.readConfigFile(configFile, args)
	appId := ""
	if len(args.Args) > 0 {
		appId = args.Args[0]
	} else {
		appId = this.createAppId()
	}
	if appId == "" {
		log.Fatal("appId is empty.")
	}
	if len(config.Clients) == 0 {
		config.Clients = []options.Client{}
	}
	var client options.Client
	var index int
	has_client := false
	for index, client = range config.Clients {
		if client.AppId == appId {
			has_client = true
			client.Key = this.createApiKey()
			config.Clients[index] = client
			log.Info("API Client updated!")
			break
		}
	}

	if !has_client {
		client = options.Client{
			AppId: appId,
			Key:   this.createApiKey(),
		}
		config.Clients = append(config.Clients, client)
		log.Info("API Client added!")
	}
	log.Info(fmt.Sprintf("appId: %s", client.AppId))
	log.Info(fmt.Sprintf("key: %s", client.Key))

	if err := this.saveConfig(config, args.Config); err != nil {
		log.Fatal(err)
	}
}

/**
 * Remove a registered referrer.
 */
func (this *Cli) ClientRemove(args *Args) {
	configFile := this.getConfigFile(args.Config, args.Dir)
	if !this.FileExists(configFile) {
		log.Fatal(fmt.Sprintf(`Error: The config file [%s] cound not be found.`, args.Config))
	}
	config := this.readConfigFile(configFile, args)
	appId := ""
	if len(args.Args) > 0 {
		appId = args.Args[0]
	} else {
		appId = this.createAppId()
	}
	if appId == "" {
		log.Fatal("appId is empty.")
	}
	clients_length := len(config.Clients)
	if clients_length == 0 {
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
	if (_index + 1) < clients_length {
		config.Clients = append(config.Clients[:_index], config.Clients[_index+1:]...)
	} else {
		config.Clients = config.Clients[:_index]
	}

	log.Info(fmt.Sprintf("Client removed: %s", appId))

	if err := this.saveConfig(config, args.Config); err != nil {
		log.Fatal(err)
	}
}

/**
 * Gets the config file with the provided args
 */
func (this *Cli) getConfigFile(file string, dir string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	filePath := filepath.Clean(path.Join(dir, file))
	if filepath.IsAbs(filePath) {
		return filePath
	}

	return filepath.Clean(path.Join(cwd, filePath))
}

/**
 * Tries to read a config file
 */
func (this *Cli) readConfigFile(file string, args *Args) options.Config {
	var data options.Config

	bytes_data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(bytes_data, &data); err != nil {
		log.Fatal(err)
	}

	return this.resolveEnvFileOptions(data, args)
}
