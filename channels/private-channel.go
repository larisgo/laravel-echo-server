package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/larisgo/laravel-echo-server/errors"
	"github.com/larisgo/laravel-echo-server/http"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/pschlump/socketio"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type PrivateChannel struct {
	/**
	 * Request client.
	 */
	client *http.Client

	/**
	 * Configurable server options.
	 */
	options options.Config

	// mu
	mu sync.RWMutex
}

/**
 * Create a new private channel instance.
 */
func NewPrivateChannel(Options options.Config) *PrivateChannel {
	this := &PrivateChannel{}
	this.options = Options
	this.client = http.NewClient()
	return this
}

/**
 * Send authentication request to application server.
 */
func (this *PrivateChannel) Authenticate(socket socketio.Socket, data types.Data) (interface{}, error) {
	body, err := json.Marshal(map[string]string{
		"channel_name": data.Channel,
	})
	if err != nil {
		return nil, err
	}
	options := &http.Options{
		Method: "POST",
		Headers: (func() map[string]string {
			data.Auth.Headers["Content-Type"] = "application/json; charset=UTF-8"
			return data.Auth.Headers
		})(),
		Url:  fmt.Sprintf("%s%s", this.authHost(socket), this.options.AuthEndpoint),
		Body: bytes.NewReader(body),
	}

	if this.options.DevMode {
		log.Warning(fmt.Sprintf(`Sending auth request to: %s`, options.Url))
	}

	return this.serverRequest(socket, options, data.Channel)
}

/**
 * Get the auth host based on the Socket.
 */
func (this *PrivateChannel) authHost(socket socketio.Socket) string {
	this.mu.RLock()
	defer this.mu.RUnlock()

	// var authHosts_interface interface{}
	// if this.options.AuthHost != nil {
	// 	authHosts_interface = this.options.AuthHost
	// } else {
	// 	authHosts_interface = this.options.Host
	// }

	authHosts := options.Hosts{}
	switch hosts := this.options.AuthHost.(type) {
	case string:
		authHosts = options.Hosts{hosts}
	case options.Hosts:
		authHosts = hosts
	}

	authHostSelected := "http://localhost"
	if len(authHosts) > 0 {
		authHostSelected = authHosts[0]
	}

	if _, ok := socket.Request().Header["Referer"]; ok {
		if r := socket.Request().Header.Get("Referer"); r != "" {
			if referer, err := url.Parse(r); err != nil {
				for _, authHost := range authHosts {
					authHostSelected = authHost
					if this.hasMatchingHost(referer, authHost) {
						authHostSelected = fmt.Sprintf(`%s//%s`, referer.Scheme, referer.Host)
						break
					}
				}
			}
		}
	}

	if !regexp.MustCompile(`^http(s)?://`).MatchString(authHostSelected) {
		authHostSelected = fmt.Sprintf(`%s://%s`, this.options.AuthProtocol, authHostSelected)
	}

	if this.options.DevMode {
		log.Warning(fmt.Sprintf(`Preparing authentication request to: %s`, authHostSelected))
	}

	return authHostSelected
}

/**
 * Check if there is a matching auth host.
 */
func (this *PrivateChannel) hasMatchingHost(referer *url.URL, host string) bool {
	hostname := referer.Hostname()
	return hostname[strings.Index(hostname, `.`):] == host || fmt.Sprintf(`%s//%s`, referer.Scheme, referer.Host) == host || referer.Host == host
}

/**
 * Send a request to the server.
 */
func (this *PrivateChannel) serverRequest(socket socketio.Socket, options *http.Options, channel_name string) (interface{}, error) {
	options.Headers = this.prepareHeaders(socket, options)
	response, err := this.client.Request(options)
	if err != nil {
		if this.options.DevMode {
			log.Error(fmt.Sprintf(`Error authenticating %s for %s`, socket.Id(), channel_name))
			log.Error(err)
		}
		return nil, err
	}
	if response.StatusCode != 200 {
		if this.options.DevMode {
			log.Warning(fmt.Sprintf(`%s could not be authenticated to %s`, socket.Id(), channel_name))
			log.Error(string(response.BodyBytes))
		}
		return nil, errors.NewError(fmt.Sprintf(`Client can not be authenticated, got HTTP status %d`, response.StatusCode))
	}
	if this.options.DevMode {
		log.Info(fmt.Sprintf(`%s authenticated for: %s`, socket.Id(), channel_name))
	}
	var body interface{}
	if err := json.Unmarshal(response.BodyBytes, &body); err != nil {
		body = string(response.BodyBytes)
		// return nil, err
	}
	return body, nil
}

/**
 * Prepare headers for request to app server.
 */
func (this *PrivateChannel) prepareHeaders(socket socketio.Socket, options *http.Options) map[string]string {
	if cookie, HasCookie := options.Headers[`Cookie`]; !HasCookie && cookie != "" {
		if _, ok := socket.Request().Header["Cookie"]; ok {
			if c := socket.Request().Header.Get("Cookie"); c != "" {
				options.Headers[`Cookie`] = c
			}
		}
	}
	options.Headers[`X-Requested-With`] = `XMLHttpRequest`

	return options.Headers
}
