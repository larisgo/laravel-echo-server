package types

type Auth struct {
	Headers map[string]string `json:"headers"`
}

type Data struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Data    interface{} `json:"data"`
	Auth    Auth        `json:"auth"`
	Socket  string      `json:"socket"`
}

type Member struct {
	SocketId string      `json:"socket_id"`
	UserId   int64       `json:"user_id"`
	UserInfo interface{} `json:"user_info"`
}

type AuthenticateData struct {
	ChannelData Member `json:"channel_data"`
}
