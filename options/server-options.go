package options

import (
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/socket.io/socket"
	"time"
)

type ServerOptions struct {
	// how many ms without a pong packet to consider the connection closed
	PingTimeout *int64 `json:"pingTimeout,omitempty"`

	// how many ms before sending a new ping packet
	PingInterval *int64 `json:"pingInterval,omitempty"`

	// how many ms before an uncompleted transport upgrade is cancelled
	UpgradeTimeout *int64 `json:"upgradeTimeout,omitempty"`

	// how many bytes or characters a message can be, before closing the session (to avoid DoS).
	MaxHttpBufferSize *int64 `json:"maxHttpBufferSize,omitempty"`

	// whether to allow transport upgrades
	AllowUpgrades *bool `json:"allowUpgrades,omitempty"`

	// parameters of the WebSocket permessage-deflate extension (see ws module api docs). Set to false to disable.
	PerMessageDeflate *types.PerMessageDeflate `json:"perMessageDeflate,omitempty"`

	// parameters of the http compression for the polling transports (see zlib api docs). Set to false to disable.
	HttpCompression *types.HttpCompression `json:"httpCompression,omitempty"`

	// the options that will be forwarded to the cors module
	Cors *types.Cors `json:"cors,omitempty"`

	// whether to enable compatibility with Socket.IO v2 clients
	AllowEIO3 *bool `json:"allowEIO3,omitempty"`

	// name of the path to capture
	Path *string `json:"path,omitempty"`

	// destroy unhandled upgrade requests
	DestroyUpgrade *bool `json:"destroyUpgrade,omitempty"`

	//  milliseconds after which unhandled requests are ended
	DestroyUpgradeTimeout *int64 `json:"destroyUpgradeTimeout,omitempty"`

	// whether to serve the client files
	ServeClient *bool `json:"serveClient,omitempty"`

	// how many ms before a client without namespace is closed
	ConnectTimeout *int64 `json:"connectTimeout,omitempty"`
}

func (s *ServerOptions) Config() *socket.ServerOptions {
	if s == nil {
		return nil
	}

	c := socket.DefaultServerOptions()
	if s.PingTimeout != nil {
		c.SetPingTimeout(time.Duration(*s.PingTimeout) * time.Millisecond)
	}
	if s.PingInterval != nil {
		c.SetPingInterval(time.Duration(*s.PingInterval) * time.Millisecond)
	}
	if s.UpgradeTimeout != nil {
		c.SetUpgradeTimeout(time.Duration(*s.UpgradeTimeout) * time.Millisecond)
	}
	if s.MaxHttpBufferSize != nil {
		c.SetMaxHttpBufferSize(*s.MaxHttpBufferSize)
	}
	if s.AllowUpgrades != nil {
		c.SetAllowUpgrades(*s.AllowUpgrades)
	}
	if s.PerMessageDeflate != nil {
		c.SetPerMessageDeflate(s.PerMessageDeflate)
	}
	if s.HttpCompression != nil {
		c.SetHttpCompression(s.HttpCompression)
	}
	if s.Cors != nil {
		c.SetCors(s.Cors)
	}
	if s.AllowEIO3 != nil {
		c.SetAllowEIO3(*s.AllowEIO3)
	}
	if s.Path != nil {
		c.SetPath(*s.Path)
	}
	if s.DestroyUpgrade != nil {
		c.SetDestroyUpgrade(*s.DestroyUpgrade)
	}
	if s.DestroyUpgradeTimeout != nil {
		c.SetDestroyUpgradeTimeout(time.Duration(*s.DestroyUpgradeTimeout) * time.Millisecond)
	}
	if s.ServeClient != nil {
		c.SetServeClient(*s.ServeClient)
	}
	if s.ConnectTimeout != nil {
		c.SetConnectTimeout(time.Duration(*s.ConnectTimeout) * time.Millisecond)
	}

	return c
}
