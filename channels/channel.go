package channels

import (
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/zishang520/engine.io/utils"
	"github.com/zishang520/socket.io/socket"
	"regexp"
	"strings"
)

type Channel struct {

	// Channels and patters for private channels.
	privateChannels []*regexp.Regexp

	// Allowed client events
	clientEvents []*regexp.Regexp

	// Private channel instance.
	Private *PrivateChannel

	// Presence channel instance.
	Presence *PresenceChannel

	// Configurable server options.
	options *options.Config

	// Socket.io client.
	io *socket.Server
}

// Create a new channel instance.
func NewChannel(io *socket.Server, _options *options.Config) (ch *Channel, err error) {
	ch = &Channel{}

	ch.io = io
	ch.options = _options

	ch.privateChannels = []*regexp.Regexp{
		regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(`private-*`), `\*`, `.*`)),
		regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(`presence-*`), `\*`, `.*`)),
	}

	ch.clientEvents = []*regexp.Regexp{
		regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(`client-*`), `\*`, `.*`)),
	}

	ch.Private = NewPrivateChannel(ch.options)
	ch.Presence, err = NewPresenceChannel(io, ch.options)
	if err != nil {
		return nil, err
	}

	if ch.options.DevMode {
		utils.Log().Success(`Channels are ready.`)
	}

	return ch, nil
}

// Join a channel.
func (ch *Channel) Join(_socket *socket.Socket, data *types.Data) {
	if data.Channel != "" {
		if ch.IsPrivate(data.Channel) {
			ch.JoinPrivate(_socket, data)
		} else {
			_socket.Join(socket.Room(data.Channel))
			ch.OnJoin(_socket, data.Channel)
		}
	}
}

// Trigger a client message
func (ch *Channel) ClientEvent(_socket *socket.Socket, data *types.Data) {
	if data.Event != "" && data.Channel != "" {
		if ch.IsClientEvent(data.Event) &&
			ch.IsPrivate(data.Channel) &&
			ch.IsInChannel(_socket, data.Channel) {
			// ch.io.Sockets().Sockets().Load(_socket.Id())
			_socket.Broadcast().To(socket.Room(data.Channel)).Emit(data.Event, data.Channel, data.Data)
		}
	}
}

// Leave a channel.
func (ch *Channel) Leave(_socket *socket.Socket, channel string, reason string) {
	if channel != "" {
		if ch.IsPresence(channel) {
			ch.Presence.Leave(_socket, channel)
		}

		_socket.Leave(socket.Room(channel))

		if ch.options.DevMode {
			utils.Log().Info(`%s left channel: %s (%s)`, _socket.Id(), channel, reason)
		}
	}
}

// Check if the incoming socket connection is a private channel.
func (ch *Channel) IsPrivate(channel string) bool {
	for _, privateChannel := range ch.privateChannels {
		if privateChannel.MatchString(channel) {
			return true
		}
	}
	return false
}

// Join private channel, emit data to presence channels.
func (ch *Channel) JoinPrivate(_socket *socket.Socket, data *types.Data) {
	res, status, err := ch.Private.Authenticate(_socket, data)
	if err != nil {
		if ch.options.DevMode {
			utils.Log().Error("%v", err)
		}
		_socket.Emit("subscription_error", data.Channel, status)
	} else {
		_socket.Join(socket.Room(data.Channel))
		if ch.IsPresence(data.Channel) {
			if channel_data, is_auth := res.(*types.AuthenticateData); is_auth {
				ch.Presence.Join(_socket, data.Channel, &channel_data.ChannelData)
			}
		}
		ch.OnJoin(_socket, data.Channel)
	}
}

// Check if a channel is a presence channel.
func (ch *Channel) IsPresence(channel string) bool {
	return strings.LastIndex(channel, `presence-`) == 0
}

// On join a channel log success.
func (ch *Channel) OnJoin(_socket *socket.Socket, channel string) {
	if ch.options.DevMode {
		utils.Log().Info(`%s joined channel: %s`, _socket.Id(), channel)
	}
}

// Check if client is a client event
func (ch *Channel) IsClientEvent(event string) bool {
	for _, clientEvent := range ch.clientEvents {
		if clientEvent.MatchString(event) {
			return true
		}
	}
	return false
}

// Check if a socket has joined a channel.
func (ch *Channel) IsInChannel(_socket *socket.Socket, channel string) bool {
	return _socket.Rooms().Has(socket.Room(channel))
}
