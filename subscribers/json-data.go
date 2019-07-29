package subscribers

type HttpSubscriberData struct {
	Channels []string `json:"channels"`
	Channel  string   `json:"channel"`
	Name     string   `json:"name"`
	Data     string   `json:"data"`
	SocketId string   `json:"socket_id"`
}
