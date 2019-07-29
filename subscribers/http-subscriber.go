package subscribers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/larisgo/laravel-echo-server/errors"
	"github.com/larisgo/laravel-echo-server/express"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpSubscriber struct {
	/**
	 * The server.
	 *
	 * @type {*httprouter.Router}
	 */
	express *express.Express

	/**
	 * Configurable server options.
	 */
	options options.Config
}

/**
 * Create new instance of http subscriber.
 *
 * @param  {any} express
 */
func NewHttpSubscriber(express *express.Express, Options options.Config) Subscriber {
	this := &HttpSubscriber{}
	this.express = express
	this.options = Options
	return Subscriber(this)
}

/**
 * Subscribe to events to broadcast.
 *
 * @return {void}
 */
func (this *HttpSubscriber) Subscribe(callback Broadcast) {
	// Broadcast a message to a channel
	this.express.Route().POST("/apps/:appId/events", this.express.AuthorizeRequests(func(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
		this.handleData(w, r, router, callback)
	}))

	log.Success("Listening for http events...")
}

/**
 * Handle incoming event data.
 *
 * @param  {any} req
 * @param  {any} res
 * @param  {any} body
 * @param  {Function} broadcast
 * @return {boolean}
 */
func (this *HttpSubscriber) handleData(w http.ResponseWriter, r *http.Request, router httprouter.Params, broadcast Broadcast) error {
	// body = JSON.parse(Buffer.concat(body).toString());
	post_data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		this.badResponse(w, r, `Failed to get data`)
		return err
	}

	var body HttpSubscriberData
	if err := json.Unmarshal(post_data, &body); err != nil {
		this.badResponse(w, r, `JSON parsing error`)
		return err
	}
	if (len(body.Channels) > 0 || body.Channel != "") && body.Name != "" && body.Data != "" {
		var data interface{}
		if err := json.Unmarshal([]byte(body.Data), &data); err != nil {
			this.badResponse(w, r, `Body data JSON parsing error`)
			return err
		}

		message := types.Data{
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

		if this.options.DevMode {
			log.Info(fmt.Sprintf("Channel: %s", this.join(channels, ", ")))
			log.Info(fmt.Sprintf("Event: %s", message.Event))
		}
		for _, channel := range channels {
			// sync
			broadcast(strings.TrimPrefix(channel, this.options.DatabaseConfig.Prefix), message)
		}
	} else {
		this.badResponse(w, r, `Event must include channel, event name and data`)
		return errors.NewError(`Event must include channel, event name and data`)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, fmt.Sprintf(`{"message": "OK"}`))

	return nil
}

/**
 * Handle bad Request.
 *
 * @param  {http.ResponseWriter} w
 * @param  {*http.Request} r
 */
func (this *HttpSubscriber) badResponse(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(400)
	fmt.Fprint(w, fmt.Sprintf(`{"error": "%s"}`, message))
}

/**
 * join
 */
func (this *HttpSubscriber) join(v []string, splite string) string {
	var buf bytes.Buffer
	for _, v := range v {
		if buf.Len() > 0 {
			buf.WriteString(splite)
		}
		buf.WriteString(v)
	}
	return buf.String()
}
