package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/larisgo/laravel-echo-server/channels"
	"github.com/larisgo/laravel-echo-server/express"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"github.com/zishang520/socket.io/socket"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

type HttpApi struct {
	// Configurable server options.
	options *options.Config

	// The server.
	express *express.Express

	// Channel instance.
	channel *channels.Channel

	// Socket.io client.
	io *socket.Server
}

// Create new instance of http subscriber.
func NewHttpApi(io *socket.Server, channel *channels.Channel, express *express.Express, _options *options.Config) *HttpApi {
	api := &HttpApi{}
	api.io = io
	api.channel = channel
	api.express = express
	api.options = _options
	return api
}

// Initialize the API.
func (api *HttpApi) Init() {
	api.corsMiddleware()

	api.express.Route().GET("/", api.GetRoot)

	api.express.Route().GET("/apps/:appId/status", api.express.AuthorizeRequests(api.GetStatus))

	api.express.Route().GET("/apps/:appId/channels", api.express.AuthorizeRequests(api.GetChannels))

	api.express.Route().GET("/apps/:appId/channels/:channelName", api.express.AuthorizeRequests(api.GetChannel))

	api.express.Route().GET("/apps/:appId/channels/:channelName/users", api.express.AuthorizeRequests(api.GetChannelUsers))

}

// Add CORS middleware if applicable.
func (api *HttpApi) corsMiddleware() {
	if api.options.ApiOriginAllow.AllowCors {
		api.express.Use(func(w http.ResponseWriter, r *http.Request, next func()) {
			w.Header().Set("Access-Control-Allow-Origin", api.options.ApiOriginAllow.AllowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", api.options.ApiOriginAllow.AllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", api.options.ApiOriginAllow.AllowHeaders)
			next()
		})
	}
}

// Outputs a simple message to show that the server is running.
func (api *HttpApi) GetRoot(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `OK`)
}

// Get the status of the server.
func (api *HttpApi) GetStatus(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	data, err := json.Marshal(map[string]interface{}{
		"subscription_count": api.io.Engine().ClientsCount(),
		"uptime":             time.Since(startTime),
		"memory_usage":       m.TotalAlloc,
	})
	if err != nil {
		if api.options.DevMode {
			utils.Log().Error("%v", err)
		}
		api.badResponse(w, r, err.Error())
		return
	} else {
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// Get a list of the open channels on the server.
func (api *HttpApi) GetChannels(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	prefix := r.URL.Query().Get("filter_by_prefix")
	rooms := api.io.Sockets().Adapter().Rooms()
	channels := map[socket.Room]map[string]interface{}{}
	rooms.Range(func(channelName, sockets interface{}) bool {
		cn := channelName.(socket.Room)
		ss := sockets.(*types.Set[socket.SocketId])
		if ss.Has(socket.SocketId(cn)) {
			return true
		}
		if prefix != "" && strings.Index(string(cn), prefix) != 0 {
			return true
		}
		channels[cn] = map[string]interface{}{
			"subscription_count": ss.Len(),
			"occupied":           true,
		}
		return true
	})

	data, err := json.Marshal(map[string]interface{}{
		"channels": channels,
	})
	if err != nil {
		if api.options.DevMode {
			utils.Log().Error("%v", err)
		}
		api.badResponse(w, r, err.Error())
		return
	} else {
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// Get a information about a channel.
func (api *HttpApi) GetChannel(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	channelName := router.ByName("channelName")
	subscriptionCount := 0
	if sockets, ok := api.io.Sockets().Adapter().Rooms().Load(channelName); ok {
		subscriptionCount = sockets.(*types.Set[socket.SocketId]).Len()
	}
	result := map[string]interface{}{
		"subscription_count": subscriptionCount,
		"occupied":           subscriptionCount > 0,
	}
	if api.channel.IsPresence(channelName) {
		members, err := api.channel.Presence.GetMembers(channelName)
		if err != nil {
			if api.options.DevMode {
				utils.Log().Error("%v", err)
			}
			api.badResponse(w, r, err.Error())
			return
		} else {
			result["user_count"] = len(members.Unique(false))
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		if api.options.DevMode {
			utils.Log().Error("%v", err)
		}
		api.badResponse(w, r, err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Get the users of a channel.
func (api *HttpApi) GetChannelUsers(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	channelName := router.ByName("channelName")

	if !api.channel.IsPresence(channelName) {
		api.badResponse(w, r, "User list is only possible for Presence Channels")
		return
	}

	members, err := api.channel.Presence.GetMembers(channelName)
	if err != nil {
		if api.options.DevMode {
			utils.Log().Error("%v", err)
		}
		api.badResponse(w, r, err.Error())
		return
	}

	users := []uint64{}
	for _, member := range members.Unique(false) {
		users = append(users, member.UserId)
	}

	data, err := json.Marshal(map[string]interface{}{
		"users": users,
	})
	if err != nil {
		if api.options.DevMode {
			utils.Log().Error("%v", err)
		}
		api.badResponse(w, r, err.Error())
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Handle bad Request.
func (api *HttpApi) badResponse(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	data, _ := json.Marshal(map[string]interface{}{
		"error": message,
	})
	w.Write(data)
}
