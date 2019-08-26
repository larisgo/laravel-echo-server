package channels

import (
	"encoding/json"
	"fmt"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/pschlump/socketio"
	"regexp"
	"strings"
)

type Channel struct {
	/**
	 * Channels and patters for private channels.
	 */
	privateChannels []string

	/**
	 * Allowed client events
	 */
	clientEvents []string

	/**
	 * Private channel instance.
	 */
	Private *PrivateChannel

	/**
	 * Presence channel instance.
	 */
	Presence *PresenceChannel

	/**
	 * Configurable server options.
	 */
	options options.Config

	/**
	 * Socket.io client.
	 *
	 * @type {*socketio.Server}
	 */
	io *socketio.Server
}

/**
 * Create a new channel instance.
 */
func NewChannel(io *socketio.Server, Options options.Config) *Channel {
	this := &Channel{}

	this.io = io
	this.options = Options

	this.privateChannels = []string{
		`private-*`,
		`presence-*`,
	}

	this.clientEvents = []string{
		`client-*`,
	}

	this.Private = NewPrivateChannel(this.options)
	this.Presence = NewPresenceChannel(io, this.options)

	if this.options.DevMode {
		log.Success(`Channels are ready.`)
	}

	return this
}

/**
 * Join a channel.
 */
func (this *Channel) Join(socket socketio.Socket, data types.Data) {
	if data.Channel != "" {
		if this.IsPrivate(data.Channel) {
			this.JoinPrivate(socket, data)
		} else {
			socket.Join(data.Channel)
			this.OnJoin(socket, data.Channel)
		}
	}
}

/**
 * Trigger a client message
 */
func (this *Channel) ClientEvent(socket socketio.Socket, data types.Data) {
	if data.Event != "" && data.Channel != "" {
		if this.IsClientEvent(data.Event) &&
			this.IsPrivate(data.Channel) &&
			this.IsInChannel(socket, data.Channel) {
			socket.BroadcastTo(data.Channel, data.Event, data.Channel, data.Data)
		}
	}
}

/**
 * Leave a channel.
 */
func (this *Channel) Leave(socket socketio.Socket, channel string, reason string) {
	if channel != "" {
		if this.IsPresence(channel) {
			this.Presence.Leave(socket, channel)
		}

		socket.Leave(channel)

		if this.options.DevMode {
			log.Info(fmt.Sprintf(`%S left channel: %s (%s)`, socket.Id(), channel, reason))
		}
	}
}

/**
 * Check if the incoming socket connection is a private channel.
 */
func (this *Channel) IsPrivate(channel string) bool {
	for _, privateChannel := range this.privateChannels {
		if regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(privateChannel), `\*`, `.*`)).MatchString(channel) {
			return true
		}
	}
	return false
}

/**
 * Join private channel, emit data to presence channels.
 */
func (this *Channel) JoinPrivate(socket socketio.Socket, data types.Data) {
	res, code, err := this.Private.Authenticate(socket, data)
	if err != nil {
		if this.options.DevMode {
			log.Error(err)
		}
		socket.Emit("subscription_error", data.Channel, code)
	} else {
		socket.Join(data.Channel)
		if this.IsPresence(data.Channel) {
			var res_channel_data types.AuthenticateData
			if err := json.Unmarshal(res, &res_channel_data); err == nil {
				if _, err := this.Presence.Join(socket, data.Channel, &res_channel_data.ChannelData); err != nil {
					if this.options.DevMode {
						log.Error(err)
					}
				}
			} else {
				if this.options.DevMode {
					log.Error(err)
				}
			}
		}
		this.OnJoin(socket, data.Channel)
	}

}

/**
 * Check if a channel is a presence channel.
 */
func (this *Channel) IsPresence(channel string) bool {
	return strings.LastIndex(channel, `presence-`) == 0
}

/**
 * On join a channel log success.
 */
func (this *Channel) OnJoin(socket socketio.Socket, channel string) {
	if this.options.DevMode {
		log.Info(fmt.Sprintf(`%s joined channel: %s`, socket.Id(), channel))
	}
}

/**
 * Check if client is a client event
 */
func (this *Channel) IsClientEvent(event string) bool {
	for _, clientEvent := range this.clientEvents {
		if regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(clientEvent), `\*`, `.*`)).MatchString(event) {
			return true
		}
	}
	return false
}

/**
 * Check if a socket has joined a channel.
 */
func (this *Channel) IsInChannel(socket socketio.Socket, channel string) bool {
	return socket.HasRoom(channel)
}
