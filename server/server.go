package server

import (
	"fmt"
	"github.com/larisgo/laravel-echo-server/errors"
	"github.com/larisgo/laravel-echo-server/express"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/pschlump/socketio"
	"net/http"
	"time"
)

type Server struct {
	/**
	 * The http server.
	 *
	 * @type {*express.Express}
	 */
	Express *express.Express

	/**
	 * Socket.io client.
	 *
	 * @type {*socketio.Server}
	 */
	Io *socketio.Server

	/**
	 * Configurable server options.
	 */
	options options.Config
}

/**
 * Create a new server instance.
 */
func NewServer(Options options.Config) *Server {
	this := &Server{}
	this.options = Options
	return this
}

/**
 * Start the Socket.io server.
 *
 * @return {*socketio.Server}
 */
func (this *Server) Init() (*socketio.Server, error) {
	if err := this.ServerProtocol(); err != nil {
		return nil, err
	}
	host := this.options.Host
	if host == "" {
		host = "localhost"
	}
	log.Success(fmt.Sprintf(`Running at %s on port %s`, host, this.GetPort()))
	return this.Io, nil
}

/**
 * Sanitize the port string from any extra characters
 *
 * @return {string}
 */
func (this *Server) GetPort() string {
	return this.options.Port
}

/**
 * Select the http protocol to run on.
 *
 * @return {error} err
 */
func (this *Server) ServerProtocol() error {
	if this.options.Protocol == "https" {
		if err := this.Secure(); err != nil {
			return err
		}
		return this.httpServer(true)
	}
	return this.httpServer(false)
}

/**
 * Load SSL 'key' & 'cert' files if https is enabled.
 *
 * @return {error} err
 */
func (this *Server) Secure() error {
	if this.options.SslCertPath == "" || this.options.SslKeyPath == "" {
		return errors.NewError(`SSL paths are missing in server config.`)
	}
	return nil
}

/**
 * Create a socket.io server.
 *
 * @return {error} err
 */
func (this *Server) httpServer(secure bool) (err error) {
	this.Express = express.NewExpress(this.options)

	this.Express.Use(func(w http.ResponseWriter, r *http.Request, next func()) {
		for key, value := range this.options.Headers {
			w.Header().Set(key, value)
		}
		next()
	})

	this.Io, err = socketio.NewServer(nil)
	if err != nil {
		return err
	}
	this.Io.SetPingTimeout(time.Duration((func() int {
		if this.options.Socketio.PingTimeout <= 0 {
			return 5000
		}
		return this.options.Socketio.PingTimeout
	})()) * time.Millisecond)
	this.Io.SetPingInterval(time.Duration((func() int {
		if this.options.Socketio.PingInterval <= 0 {
			return 25000
		}
		return this.options.Socketio.PingInterval
	})()) * time.Millisecond)

	this.Io.SetMaxConnection(this.options.Socketio.MaxConnection)

	this.Io.SetAllowRequest(func(w http.ResponseWriter, r *http.Request) error {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		return nil
	})

	http.Handle("/socket.io/", this.Io)
	http.Handle("/", this.Express)

	switch hosts := this.options.Host.(type) {
	case string:
		go this.listen(hosts, secure)
	case options.Hosts:
		for _, host := range hosts {
			go this.listen(host, secure)
		}
	default:
		return errors.NewError(`Host type error, can only be a string or string slice.`)
	}
	return nil
}

/**
 * listen
 */
func (this *Server) listen(host string, secure bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	var err error
	if secure {
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%s", host, this.GetPort()), this.options.SslCertPath, this.options.SslKeyPath, nil)
	} else {
		err = http.ListenAndServe(fmt.Sprintf("%s:%s", host, this.GetPort()), nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
