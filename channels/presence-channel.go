package channels

import (
	"encoding/json"
	"fmt"
	"github.com/larisgo/laravel-echo-server/database"
	"github.com/larisgo/laravel-echo-server/errors"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/pschlump/socketio"
	"sync"
)

type PresenceChannel struct {
	/**
	 * Database instance.
	 */
	db database.DatabaseDriver

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

	// mu
	mu sync.RWMutex
}

/**
 * Create a NewPresence channel instance.
 */
func NewPresenceChannel(io *socketio.Server, Options options.Config) *PresenceChannel {
	this := &PresenceChannel{}
	this.io = io
	this.options = Options
	this.db = database.NewDatabase(Options)
	return this
}

/**
 * Get the members of a presence channel.
 */
func (this *PresenceChannel) GetMembers(channel string) ([]*types.Member, error) {
	members_byte, err := this.db.Get(fmt.Sprintf(`%s:members`, channel))
	if err != nil {
		return nil, err
	}
	if len(members_byte) == 0 {
		return nil, nil
	}
	var members []*types.Member
	if err := json.Unmarshal(members_byte, &members); err != nil {
		return nil, err
	}
	return members, nil
}

/**
 * Check if a user is on a presence channel.
 */
func (this *PresenceChannel) IsMember(channel string, member *types.Member) (bool, error) {
	members, err := this.GetMembers(channel)
	if err != nil {
		return false, err
	}
	members, err = this.RemoveInactive(channel, members)
	if err != nil {
		return false, err
	}
	for _, m := range members {
		if m.UserId == member.UserId {
			return true, nil
		}
	}
	return false, err
}

/**
 * Remove inactive channel members from the presence channel.
 */
func (this *PresenceChannel) RemoveInactive(channel string, members []*types.Member) ([]*types.Member, error) {
	clients := this.io.Clients(channel)
	tmp_members := []*types.Member{}
	for _, member := range members {
		if _, ok := clients[member.SocketId]; ok {
			tmp_members = append(tmp_members, member)
		}
	}
	this.db.Set(fmt.Sprintf(`%s:members`, channel), tmp_members)

	return tmp_members, nil
}

func (this *PresenceChannel) uniq(members []*types.Member) []*types.Member {
	result := []*types.Member{}
	tempMap := map[int64]bool{} // 存放不重复主键
	for _, e := range members {
		l := len(tempMap)
		tempMap[e.UserId] = true
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

/**
 * Join a presence channel and emit that they have joined only if it is the
 * first instance of their presence.
 */
func (this *PresenceChannel) Join(socket socketio.Socket, channel string, member *types.Member) (*types.Member, error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if member == nil {
		if this.options.DevMode {
			log.Error(`Unable to join channel. Member data for presence channel missing`)
		}
		return nil, errors.NewError(`Unable to join channel. Member data for presence channel missing`)
	}
	is_member, err := this.IsMember(channel, member)
	if err != nil {
		if this.options.DevMode {
			log.Error(err)
		}
		return nil, err
	}
	members, err := this.GetMembers(channel)
	if err != nil {
		return nil, err
	}
	member.SocketId = socket.Id()
	members = append(members, member)

	this.db.Set(fmt.Sprintf(`%s:members`, channel), members)

	members = this.uniq(members)

	this.OnSubscribed(socket, channel, members)

	if !is_member {
		this.OnJoin(socket, channel, member)
	}
	return member, nil
}

/**
 * Remove a member from a presenece channel and broadcast they have left
 * only if not other presence channel instances exist.
 */
func (this *PresenceChannel) Leave(socket socketio.Socket, channel string) (*types.Member, error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	members, err := this.GetMembers(channel)
	if err != nil {
		if this.options.DevMode {
			log.Error(err)
		}
		return nil, err
	}

	member := &types.Member{}
	tmp_members := []*types.Member{}
	for _, v := range members {
		if v.SocketId == socket.Id() {
			member = v
		} else {
			tmp_members = append(tmp_members, member)
		}
	}
	this.db.Set(fmt.Sprintf(`%s:members`, channel), tmp_members)
	is_member, err := this.IsMember(channel, member)
	if err != nil {
		if this.options.DevMode {
			// Error retrieving pressence channel members.
			log.Error(err)
		}
		return nil, err
	}
	if !is_member {
		member.SocketId = "" // 清除socketid
		this.OnLeave(channel, member)
	}
	return member, nil
}

/**
 * On join event handler.
 */
func (this *PresenceChannel) OnJoin(socket socketio.Socket, channel string, member *types.Member) {
	socket.BroadcastTo(channel, `presence:joining`, channel, member)
}

/**
 * On Leave emitter.
 */
func (this *PresenceChannel) OnLeave(channel string, member *types.Member) {
	this.io.BroadcastTo(channel, `presence:leaving`, channel, member)
}

/**
 * On subscribed event emitter.
 */
func (this *PresenceChannel) OnSubscribed(socket socketio.Socket, channel string, members []*types.Member) {
	socket.Emit(`presence:subscribed`, channel, members)
}
