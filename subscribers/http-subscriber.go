package subscribers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/larisgo/laravel-echo-server/express"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/zishang520/engine.io/utils"
)

type HttpSubscriber struct {

	// The server.
	express *express.Express

	// Configurable server options.
	options *options.Config

	_close bool
	mu     sync.RWMutex
}

// Create new instance of http subscriber.
func NewHttpSubscriber(express *express.Express, _options *options.Config) Subscriber {
	sub := &HttpSubscriber{}
	sub.express = express
	sub.options = _options
	sub._close = false
	return sub
}

// Subscribe to events to broadcast.
func (sub *HttpSubscriber) Subscribe(callback Broadcast) {
	// Broadcast a message to a channel
	sub.express.Route().POST("/apps/:appId/events", sub.express.AuthorizeRequests(func(w http.ResponseWriter, r *http.Request, router httprouter.Params) {

		if sub.unSubscribed() {
			w.WriteHeader(http.StatusNotFound)
			w.Write(nil)
		} else {
			sub.handleData(w, r, router, callback)
		}
	}))

	utils.Log().Success("Listening for http events...")
}

// Unsubscribe from events to broadcast.
func (sub *HttpSubscriber) UnSubscribe() {
	sub.mu.Lock()
	defer sub.mu.Unlock()

	sub._close = true
}

func (sub *HttpSubscriber) unSubscribed() bool {
	sub.mu.RLock()
	defer sub.mu.RUnlock()

	return sub._close
}

// Handle incoming event data.
func (sub *HttpSubscriber) handleData(w http.ResponseWriter, r *http.Request, router httprouter.Params, broadcast Broadcast) {
	data := bytes.NewBuffer(nil)

	if bd, ok := r.Body.(io.ReadCloser); ok && bd != nil {
		data.ReadFrom(bd)
		bd.Close()
	} else {
		sub.badResponse(w, r, `Event must include channel, event name and data`)
		return
	}

	var body HttpSubscriberData
	if err := json.NewDecoder(data).Decode(&body); err != nil {
		sub.badResponse(w, r, err.Error())
		return
	}
	if (len(body.Channels) > 0 || body.Channel != "") && body.Name != "" && body.Data != "" {
		var data any
		if err := json.Unmarshal([]byte(body.Data), &data); err != nil {
			sub.badResponse(w, r, err.Error())
			return
		}

		message := &types.Data{
			Event:  body.Name,
			Data:   data,
			Socket: body.SocketId,
		}
		channels := []string{}
		if len(body.Channels) > 0 {
			channels = body.Channels
		} else {
			channels = []string{body.Channel}
		}

		if sub.options.DevMode {
			utils.Log().Info("Channel: " + sub.join(channels, ", "))
			utils.Log().Info("Event: " + message.Event)
		}
		for _, channel := range channels {
			// sync
			broadcast(channel, message)
		}
	} else {
		sub.badResponse(w, r, `Event must include channel, event name and data`)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, `{"message":"ok"}`)
}

// Handle bad Request.
func (sub *HttpSubscriber) badResponse(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	data, _ := json.Marshal(map[string]any{
		"error": message,
	})
	w.Write(data)
}

// join
func (sub *HttpSubscriber) join(v []string, splite string) string {
	sb := new(strings.Builder)
	for _, v := range v {
		if sb.Len() > 0 {
			sb.WriteString(splite)
		}
		sb.WriteString(v)
	}
	return sb.String()
}
