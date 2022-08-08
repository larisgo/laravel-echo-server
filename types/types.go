package types

import (
	"github.com/zishang520/socket.io/socket"
)

type Auth struct {
	Headers map[string]string `json:"headers" mapstructure:"headers"`
}

type Data struct {
	Channel string      `json:"channel" mapstructure:"channel"`
	Event   string      `json:"event" mapstructure:"event"`
	Data    interface{} `json:"data" mapstructure:"data"`
	Auth    Auth        `json:"auth" mapstructure:"auth"`
	Socket  string      `json:"socket" mapstructure:"socket"`
}

type Member struct {
	SocketId socket.SocketId `json:"socket_id"`
	UserId   uint64          `json:"user_id"`
	UserInfo interface{}     `json:"user_info"`
}

type Members []*Member

func (members Members) Unique(reverse bool) (result Members) {
	tempMap := map[uint64]struct{}{}
	if reverse {
		for i := len(members) - 1; i > 0; i = i - 1 {
			if _, ok := tempMap[members[i].UserId]; !ok {
				tempMap[members[i].UserId] = struct{}{}
				result = append(result, members[i])
			}
		}
	} else {
		for i, j := 0, len(members); i < j; i = i + 1 {
			if _, ok := tempMap[members[i].UserId]; !ok {
				tempMap[members[i].UserId] = struct{}{}
				result = append(result, members[i])
			}
		}
	}
	tempMap = nil
	return result
}

type AuthenticateData struct {
	ChannelData Member `json:"channel_data"`
}

type PocessLockData struct {
	Process int `json:"process"`
}
