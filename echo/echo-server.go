package echo

import (
	"github.com/larisgo/laravel-echo-server/api"
	"github.com/larisgo/laravel-echo-server/channels"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/server"
	"github.com/larisgo/laravel-echo-server/subscribers"
	"github.com/larisgo/laravel-echo-server/types"
	_utils "github.com/larisgo/laravel-echo-server/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/zishang520/engine.io/utils"
	"github.com/zishang520/socket.io/socket"
	"sync"
)

type EchoServer struct {

	// Default server options.
	DefaultOptions *options.Config

	// Configurable server options.
	options *options.Config

	// Socket.io server instance.
	server *server.Server

	// Channel instance.
	channel *channels.Channel

	// Subscribers
	subscribers []subscribers.Subscriber

	// Http api instance.
	httpApi *api.HttpApi

	mu sync.RWMutex
}

// Create a new instance.
func NewEchoServer() *EchoServer {
	ec := &EchoServer{}

	ec.DefaultOptions = &options.Config{
		AuthHost:     "http://localhost",
		AuthEndpoint: "/broadcasting/auth",
		Clients:      []options.Client{},
		Database:     "redis",
		DatabaseConfig: options.DatabaseConfig{
			Sqlite: options.Sqlite{
				DatabasePath: "/database/laravel-echo-server.sqlite",
			},
		},
		DevMode:     false,
		Host:        "127.0.0.1",
		Port:        "6001",
		Protocol:    "http",
		SslCertPath: "",
		SslKeyPath:  "",
		Subscribers: options.Subscribers{
			Http:  true,
			Redis: true,
		},
		ApiOriginAllow: options.ApiOriginAllow{
			AllowCors:    false,
			AllowOrigin:  "",
			AllowMethods: "",
			AllowHeaders: "",
		},
	}

	return ec
}

// Start the Echo Server.
func (ec *EchoServer) Run(_options *options.Config) error {
	ops, err := options.Assign(ec.DefaultOptions, _options)
	if err != nil {
		return err
	}
	ec.options = ops
	ec.Startup()

	ec.server = server.NewServer(ec.options)
	io, err := ec.server.Init()
	if err != nil {
		return err
	}
	if err := ec.Init(io); err != nil {
		return err
	}

	utils.Log().Info("Server ready!")
	return nil
}

// Initialize the class
func (ec *EchoServer) Init(io *socket.Server) (err error) {
	ec.channel, err = channels.NewChannel(io, ec.options)
	if err != nil {
		return err
	}

	ec.mu.Lock()
	ec.subscribers = []subscribers.Subscriber{}
	ec.mu.Unlock()
	if ec.options.Subscribers.Http {
		ec.mu.Lock()
		ec.subscribers = append(ec.subscribers, subscribers.NewHttpSubscriber(ec.server.Express, ec.options))
		ec.mu.Unlock()
	}
	if ec.options.Subscribers.Redis {
		r, err := subscribers.NewRedisSubscriber(ec.options)
		if err != nil {
			return err
		}
		ec.mu.Lock()
		ec.subscribers = append(ec.subscribers, r)
		ec.mu.Unlock()
	}

	ec.httpApi = api.NewHttpApi(io, ec.channel, ec.server.Express, ec.options)
	ec.httpApi.Init()

	ec.OnConnect()
	ec.Listen()
	return nil
}

// Text shown at Startup.
func (ec *EchoServer) Startup() {
	utils.Log().Println(_utils.TITLE)
	utils.Log().Println(_utils.VERSION)

	if ec.options.DevMode {
		utils.Log().Warning("Starting server in DEV mode...")
	} else {
		utils.Log().Info("Starting server...")
	}
}

// Stop the echo server.
func (ec *EchoServer) Stop() {
	utils.Log().Default("Stopping the LARAVEL ECHO SERVER")

	ec.mu.RLock()
	for _, subscriber := range ec.subscribers {
		subscriber.UnSubscribe()
	}
	ec.mu.RUnlock()

	ec.channel.Presence.Close()

	ec.server.Io.Close(nil)

	ec.mu.Lock()
	ec.subscribers = nil
	ec.mu.Unlock()

	utils.Log().Default("The LARAVEL ECHO SERVER server has been stopped.")
}

// Listen for incoming event from subscibers.
func (ec *EchoServer) Listen() {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	for _, subscriber := range ec.subscribers {
		subscriber.Subscribe(func(channel string, message *types.Data) {
			ec.Broadcast(channel, message)
		})
	}
}

// Return a channel by its socket id.
func (ec *EchoServer) Find(id string) *socket.Socket {
	if _socket, ok := ec.server.Io.Sockets().Sockets().Load(socket.SocketId(id)); ok {
		return _socket.(*socket.Socket)
	}
	return nil
}

// Broadcast events to channels from subscribers.
func (ec *EchoServer) Broadcast(channel string, message *types.Data) error {
	if socket := ec.Find(message.Socket); message.Socket != "" && socket != nil {
		return ec.ToOthers(socket, channel, message)
	} else {
		return ec.ToAll(channel, message)
	}
}

// Broadcast to others on channel.
func (ec *EchoServer) ToOthers(_socket *socket.Socket, channel string, message *types.Data) error {
	return _socket.Broadcast().To(socket.Room(channel)).Emit(message.Event, channel, message.Data)
}

// Broadcast to all members on channel.
func (ec *EchoServer) ToAll(channel string, message *types.Data) error {
	return ec.server.Io.To(socket.Room(channel)).Emit(message.Event, channel, message.Data)
}

// On server connection.
func (ec *EchoServer) OnConnect() {
	ec.server.Io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		ec.OnSubscribe(client)
		ec.OnUnsubscribe(client)
		ec.OnDisconnecting(client)
		ec.OnClientEvent(client)
	})
	ec.server.Io.On("error", func(errs ...any) {
		// errs = append(errs, (any)(""))
		utils.Log().Error("%v", errs[0])
	})
}

// On subscribe to a channel.
func (ec *EchoServer) OnSubscribe(_socket *socket.Socket) {
	_socket.On("subscribe", func(msgs ...any) {
		var data *types.Data
		if err := mapstructure.Decode(msgs[0], &data); err != nil {
			utils.Log().Error("OnSubscribe error: %v", err)
			return
		}
		ec.channel.Join(_socket, data)
	})
}

// On unsubscribe from a channel.
func (ec *EchoServer) OnUnsubscribe(_socket *socket.Socket) {
	_socket.On("unsubscribe", func(msgs ...any) {
		var data *types.Data
		if err := mapstructure.Decode(msgs[0], &data); err != nil {
			utils.Log().Error("OnUnsubscribe error: %v", err)
			return
		}
		ec.channel.Leave(_socket, data.Channel, "unsubscribed")
	})
}

// On socket disconnecting.
func (ec *EchoServer) OnDisconnecting(_socket *socket.Socket) {
	_socket.On("disconnect", func(reasons ...any) {
		for _, room := range _socket.Rooms().Keys() {
			ec.channel.Leave(_socket, string(room), reasons[0].(string))
		}
	})
}

// On client events.
func (ec *EchoServer) OnClientEvent(_socket *socket.Socket) {
	_socket.On("client event", func(msgs ...any) {
		var data *types.Data
		if err := mapstructure.Decode(msgs[0], &data); err != nil {
			utils.Log().Error("OnClientEvent error: %v", err)
			return
		}
		ec.channel.ClientEvent(_socket, data)
	})
}
