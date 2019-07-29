package echo

import (
	"github.com/larisgo/laravel-echo-server/api"
	"github.com/larisgo/laravel-echo-server/channels"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/server"
	"github.com/larisgo/laravel-echo-server/subscribers"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/larisgo/laravel-echo-server/version"
	"github.com/pschlump/socketio"
)

type EchoServer struct {
	/**
	 * Default server options.
	 */
	DefaultOptions options.Config

	/**
	 * Configurable server options.
	 */
	options options.Config

	/**
	 * Socket.io server instance.
	 */
	server *server.Server

	/**
	 * Channel instance.
	 */
	channel *channels.Channel

	/**
	 * Subscribers
	 */
	subscribers []subscribers.Subscriber

	/**
	 * Http api instance.
	 */
	httpApi *api.HttpApi
}

/**
 * Create a new instance.
 */
func NewEchoServer() *EchoServer {
	this := &EchoServer{}

	this.DefaultOptions = options.Config{
		AuthHost:     "http://localhost",
		AuthProtocol: "http",
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
		Socketio:    options.Socketio{},
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

	return this
}

/**
 * Start the Echo Server.
 */
func (this *EchoServer) Run(Options options.Config) *EchoServer {
	// _options, err := options.Assign(this.DefaultOptions, Options)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	this.options = Options
	this.Startup()

	this.server = server.NewServer(this.options)
	io, err := this.server.Init()
	if err != nil {
		log.Fatal(err)
	}
	this.Init(io)

	log.Info("Server ready!")
	return this
}

/**
 * Initialize the class
 */
func (this *EchoServer) Init(io *socketio.Server) {
	this.channel = channels.NewChannel(io, this.options)
	this.subscribers = []subscribers.Subscriber{}
	if this.options.Subscribers.Http {
		this.subscribers = append(this.subscribers, subscribers.NewHttpSubscriber(this.server.Express, this.options))
	}
	if this.options.Subscribers.Redis {
		this.subscribers = append(this.subscribers, subscribers.NewRedisSubscriber(this.options))
	}

	this.httpApi = api.NewHttpApi(io, this.channel, this.server.Express, this.options)
	this.httpApi.Init()
	this.OnConnect()
	this.Listen()
}

/**
 * Text shown at Startup.
 */
func (this *EchoServer) Startup() {
	log.Line(`
                         __             __
|   _  __ _     _  |    |_  _ |_  _    (_  _  __    _  __
|__(_| | (_|\_/(/_ |    |__(_ | |(_)   __)(/_ | \_/(/_ |
	`)
	log.Line(version.VERSION)

	if this.options.DevMode {
		log.Warning("Starting server in DEV mode...")
	} else {
		log.Info("Starting server...")
	}
}

/**
 * Listen for incoming event from subscibers.
 */
func (this *EchoServer) Listen() {
	for _, subscriber := range this.subscribers {
		subscriber.Subscribe(func(channel string, message types.Data) {
			this.Broadcast(channel, message)
		})
	}
}

/**
 * Return a channel by its socket id.
 */
func (this *EchoServer) Find(socket_id string) socketio.Socket {
	return this.server.Io.GetSocket(socket_id)
}

/**
 * Broadcast events to channels from subscribers.
 */
func (this *EchoServer) Broadcast(channel string, message types.Data) bool {
	if socket := this.Find(message.Socket); message.Socket != "" && socket != nil {
		return this.ToOthers(socket, channel, message)
	} else {
		return this.ToAll(channel, message)
	}
}

/**
 * Broadcast to others on channel.
 */
func (this *EchoServer) ToOthers(socket socketio.Socket, channel string, message types.Data) bool {
	socket.BroadcastTo(channel, message.Event, channel, message.Data)
	return true
}

/**
 * Broadcast to all members on channel.
 */
func (this *EchoServer) ToAll(channel string, message types.Data) bool {
	this.server.Io.BroadcastTo(channel, message.Event, channel, message.Data)
	return true
}

/**
 * On server connection.
 */
func (this *EchoServer) OnConnect() {
	this.server.Io.On("connection", func(socket socketio.Socket) error {
		this.OnSubscribe(socket)
		this.OnUnsubscribe(socket)
		this.OnDisconnecting(socket)
		this.OnClientEvent(socket)
		return nil
	})
}

/**
 * On subscribe to a channel.
 */
func (this *EchoServer) OnSubscribe(socket socketio.Socket) {
	socket.On("subscribe", func(coon socketio.Socket, data types.Data) error {
		this.channel.Join(socket, data)
		return nil
	})
}

/**
 * On unsubscribe from a channel.
 */
func (this *EchoServer) OnUnsubscribe(socket socketio.Socket) {
	socket.On("unsubscribe", func(coon socketio.Socket, data types.Data) error {
		this.channel.Leave(socket, data.Channel, "unsubscribed")
		return nil
	})
}

/**
 * On socket disconnecting.
 */
func (this *EchoServer) OnDisconnecting(socket socketio.Socket) {
	socket.On("disconnect", func(socket socketio.Socket) error {
		for _, room := range socket.Rooms() {
			this.channel.Leave(socket, room, "disconnect")
		}
		return nil
	})
}

/**
 * On client events.
 */
func (this *EchoServer) OnClientEvent(socket socketio.Socket) {
	socket.On("client event", func(coon socketio.Socket, data types.Data) error {
		this.channel.ClientEvent(socket, data)
		return nil
	})
}
