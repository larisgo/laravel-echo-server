package api

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/larisgo/laravel-echo-server/channels"
	"github.com/larisgo/laravel-echo-server/express"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/pschlump/socketio"
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
	/**
	 * Configurable server options.
	 */
	options options.Config

	/**
	 * The server.
	 *
	 * @type {*httprouter.Router}
	 */
	express *express.Express

	/**
	 * Channel instance.
	 */
	channel *channels.Channel

	/**
	 * Socket.io client.
	 *
	 * @type {*socketio.Server}
	 */
	io *socketio.Server
}

/**
 * Create new instance of http subscriber.
 *
 * @param  {any} io
 * @param  {any} channel
 * @param  {any} express
 * @param  {object} options object
 */
func NewHttpApi(io *socketio.Server, channel *channels.Channel, express *express.Express, Options options.Config) *HttpApi {
	this := &HttpApi{}
	this.io = io
	this.channel = channel
	this.express = express
	this.options = Options
	return this
}

/**
 * Initialize the API.
 */
func (this *HttpApi) Init() {
	this.corsMiddleware()

	this.express.Route().GET("/", this.GetRoot)

	this.express.Route().GET("/apps/:appId/status", this.express.AuthorizeRequests(this.GetStatus))

	this.express.Route().GET("/apps/:appId/channels", this.express.AuthorizeRequests(this.GetChannels))

	this.express.Route().GET("/apps/:appId/channels/:channelName", this.express.AuthorizeRequests(this.GetChannel))

	this.express.Route().GET("/apps/:appId/channels/:channelName/users", this.express.AuthorizeRequests(this.GetChannelUsers))

}

/**
 * Add CORS middleware if applicable.
 */
func (this *HttpApi) corsMiddleware() {
	if this.options.ApiOriginAllow.AllowCors {
		this.express.Use(func(w http.ResponseWriter, r *http.Request, next func()) {
			w.Header().Set("Access-Control-Allow-Origin", this.options.ApiOriginAllow.AllowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", this.options.ApiOriginAllow.AllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", this.options.ApiOriginAllow.AllowHeaders)
			next()
		})
	}
}

/**
 * Outputs a simple message to show that the server is running.
 *
 * @param {any} req
 * @param {any} res
 */
func (this *HttpApi) GetRoot(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(200)
	fmt.Fprint(w, `OK`)
}

/**
 * Get the status of the server.
 *
 * @param {any} req
 * @param {any} res
 */
func (this *HttpApi) GetStatus(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	data, err := json.Marshal(map[string]interface{}{
		"subscription_count": this.io.GetConnectionLen(),
		"uptime":             time.Since(startTime),
		"memory_usage":       m.TotalAlloc,
	})
	if err != nil && this.options.DevMode {
		log.Error(err)
	}
	w.Write(data)
}

/**
 * Get a list of the open channels on the server.
 *
 * @param {any} req
 * @param {any} res
 */
func (this *HttpApi) GetChannels(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	prefix := r.URL.Query().Get("filter_by_prefix")
	rooms := this.io.GetRoomSet()
	channels := map[string]map[string]interface{}{}
	for channelName, sockets := range rooms {
		if prefix != "" && strings.Index(channelName, prefix) != 0 {
			break
		}
		channels[channelName] = map[string]interface{}{
			"subscription_count": len(sockets),
			"occupied":           true,
		}
	}

	data, err := json.Marshal(map[string]interface{}{
		"channels": channels,
	})
	if err != nil && this.options.DevMode {
		log.Error(err)
	}
	w.Write(data)
}

func (this *HttpApi) uniq(members []*types.Member) []*types.Member {
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
 * Get a information about a channel.
 *
 * @param  {any} req
 * @param  {any} res
 */
func (this *HttpApi) GetChannel(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	channelName := router.ByName("channelName")
	room, has := this.io.GetRoomSet()[channelName]
	subscriptionCount := 0
	if has {
		subscriptionCount = len(room)
	}
	result := map[string]interface{}{
		"subscription_count": subscriptionCount,
		"occupied":           subscriptionCount > 0,
	}
	if this.channel.IsPresence(channelName) {
		members, err := this.channel.Presence.GetMembers(channelName)
		if err != nil && this.options.DevMode {
			log.Error(err)
		}
		result["user_count"] = len(this.uniq(members))
	}

	data, err := json.Marshal(result)
	if err != nil && this.options.DevMode {
		log.Error(err)
	}
	w.Write(data)
}

/**
 * Get the users of a channel.
 *
 * @param  {any} req
 * @param  {any} res
 * @return {boolean}
 */
func (this *HttpApi) GetChannelUsers(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
	channelName := router.ByName("channelName")

	if this.channel.IsPresence(channelName) {
		this.badResponse(w, r, `User list is only possible for Presence Channels`)
		return
	}

	members, err := this.channel.Presence.GetMembers(channelName)
	if err != nil && this.options.DevMode {
		log.Error(err)
	}

	users := []int64{}
	for _, member := range this.uniq(members) {
		users = append(users, member.UserId)
	}

	data, err := json.Marshal(map[string]interface{}{
		"users": users,
	})
	if err != nil && this.options.DevMode {
		log.Error(err)
	}
	w.Write(data)
}

/**
 * Handle bad Request.
 *
 * @param  {any} req
 * @param  {any} res
 * @param  {string} message
 * @return {boolean}
 */
func (this *HttpApi) badResponse(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(400)
	fmt.Fprint(w, fmt.Sprintf(`{"error": "%s"}`, message))
}
