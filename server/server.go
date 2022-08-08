package server

import (
	"errors"
	"github.com/larisgo/laravel-echo-server/express"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"github.com/zishang520/socket.io/socket"
	"net/http"
)

type Server struct {
	Express *express.Express

	// Socket.io client.
	Io *socket.Server

	// Configurable server options.
	options *options.Config

	// The http server.
	server *types.HttpServer
}

// Create a new server instance.
func NewServer(_options *options.Config) *Server {
	serv := &Server{}
	serv.options = _options
	return serv
}

// Start the Socket.io server.
func (serv *Server) Init() (*socket.Server, error) {
	if err := serv.ServerProtocol(); err != nil {
		return nil, err
	}
	host := serv.options.Host
	if host == "" {
		host = "localhost"
	}
	utils.Log().Success(`Running at %s on port %s`, host, serv.GetPort())
	return serv.Io, nil
}

// Sanitize the port string from any extra characters
func (serv *Server) GetPort() string {
	return serv.options.Port
}

// Select the http protocol to run on.
func (serv *Server) ServerProtocol() error {
	if serv.options.Protocol == "https" {
		if err := serv.Secure(); err != nil {
			return err
		}
		return serv.httpServer(true)
	}
	return serv.httpServer(false)
}

// Load SSL 'key' & 'cert' files if https is enabled.
func (serv *Server) Secure() error {
	if serv.options.SslCertPath == "" || serv.options.SslKeyPath == "" {
		return errors.New(`SSL paths are missing in server config.`)
	}
	return nil
}

// Create a socket.io server.
func (serv *Server) httpServer(secure bool) (err error) {
	serv.Express = express.NewExpress(serv.options)

	serv.Express.Use(func(w http.ResponseWriter, r *http.Request, next func()) {
		for key, value := range serv.options.Headers {
			w.Header().Set(key, value)
		}
		next()
	})

	serv.server = types.CreateServer(serv.Express)

	serv.Io = socket.NewServer(serv.server, serv.options.Socketio.Config())

	switch hosts := serv.options.Host.(type) {
	case string:
		serv.listen(hosts, secure)
	case options.Hosts:
		for _, host := range hosts {
			serv.listen(host, secure)
		}
	default:
		return errors.New(`Host type error, can only be a string or string slice.`)
	}
	return nil
}

// Close
func (serv *Server) Close() error {
	return serv.server.Close(nil)
}

// listen
func (serv *Server) listen(host string, secure bool) {
	if secure {
		serv.server.ListenTLS(host+":"+serv.GetPort(), serv.options.SslCertPath, serv.options.SslKeyPath, nil)
	} else {
		serv.server.Listen(host+":"+serv.GetPort(), nil)
	}
}
