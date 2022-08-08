package channels

import (
	"encoding/json"
	"github.com/larisgo/laravel-echo-server/database"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/zishang520/engine.io/utils"
	"github.com/zishang520/socket.io/socket"
)

type PresenceChannel struct {

	// Database instance.
	db database.DatabaseDriver

	// Configurable server options.
	options *options.Config

	// Socket.io client.
	io *socket.Server
}

// Create a NewPresence channel instance.
func NewPresenceChannel(io *socket.Server, _options *options.Config) (pch *PresenceChannel, err error) {
	pch = &PresenceChannel{}
	pch.io = io
	pch.options = _options
	pch.db, err = database.NewDatabase(_options)
	if err != nil {
		return nil, err
	}
	return pch, nil
}

func (pch *PresenceChannel) Close() error {
	return pch.db.Close()
}

// Get the members of a presence channel.
func (pch *PresenceChannel) GetMembers(channel string) (members types.Members, _ error) {
	data, err := pch.db.Get(channel + ":members")
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return members, nil
	}
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// Check if a user is on a presence channel.
func (pch *PresenceChannel) IsMember(channel string, member *types.Member) (bool, error) {
	members, err := pch.GetMembers(channel)
	if err != nil {
		return false, err
	}
	members, err = pch.RemoveInactive(channel, members, member)
	if err != nil {
		return false, err
	}
	for _, m := range members {
		if m.UserId == member.UserId {
			return true, nil
		}
	}
	return false, nil
}

// Remove inactive channel members from the presence channel.
func (pch *PresenceChannel) RemoveInactive(channel string, members types.Members, member *types.Member) (_members types.Members, _ error) {
	clients, err := pch.io.Sockets().In(socket.Room(channel)).AllSockets()
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if clients.Has(member.SocketId) {
			_members = append(_members, member)
		}
	}

	pch.db.Set(channel+":members", _members)

	return _members, nil
}

// Join a presence channel and emit that they have joined only if it is the
// first instance of their presence.
func (pch *PresenceChannel) Join(socket *socket.Socket, channel string, member *types.Member) error {
	if member == nil {
		if pch.options.DevMode {
			utils.Log().Error(`Unable to join channel. Member data for presence channel missing`)
		}
		return nil
	}
	is_member, err := pch.IsMember(channel, member)
	if err != nil {
		if pch.options.DevMode {
			utils.Log().Error("%v", err)
		}
		return err
	}
	members, err := pch.GetMembers(channel)
	if err != nil {
		if pch.options.DevMode {
			utils.Log().Error("%v", err)
		}
		return err
	}
	member.SocketId = socket.Id()
	members = append(members, member)

	pch.db.Set(channel+":members", members)

	pch.OnSubscribed(socket, channel, members.Unique(true))

	if !is_member {
		pch.OnJoin(socket, channel, member)
	}
	return nil
}

// Remove a member from a presenece channel and broadcast they have left
// only if not other presence channel instances exist.
func (pch *PresenceChannel) Leave(socket *socket.Socket, channel string) error {
	members, err := pch.GetMembers(channel)
	if err != nil {
		if pch.options.DevMode {
			utils.Log().Error("%v", err)
		}
		return err
	}

	member := &types.Member{}

	_members := types.Members{}
	for _, v := range members {
		if v.SocketId == socket.Id() {
			member = v
		} else {
			_members = append(_members, v)
		}
	}

	pch.db.Set(channel+":members", _members)

	is_member, err := pch.IsMember(channel, member)
	if err != nil {
		if pch.options.DevMode {
			// Error retrieving pressence channel members.
			utils.Log().Error("%v", err)
		}
		return err
	}
	if !is_member {
		member.SocketId = ""
		pch.OnLeave(channel, member)
	}
	return nil
}

// On join event handler.
func (pch *PresenceChannel) OnJoin(_socket *socket.Socket, channel string, member *types.Member) {
	// ch.io.Sockets().Sockets().Load(_socket.Id())
	_socket.Broadcast().To(socket.Room(channel)).Emit("presence:joining", channel, member)
}

// On Leave emitter.
func (pch *PresenceChannel) OnLeave(channel string, member *types.Member) {
	pch.io.To(socket.Room(channel)).Emit("presence:leaving", channel, member)
}

// On subscribed event emitter.
func (pch *PresenceChannel) OnSubscribed(_socket *socket.Socket, channel string, members types.Members) {
	pch.io.To(socket.Room(_socket.Id())).Emit("presence:subscribed", channel, members)
}
