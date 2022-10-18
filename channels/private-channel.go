package channels

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	_http "github.com/larisgo/laravel-echo-server/http"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/types"
	"github.com/zishang520/engine.io/utils"
	"github.com/zishang520/socket.io/socket"
)

type PrivateChannel struct {

	// Request client.
	client *_http.Client

	// Configurable server options.
	options *options.Config
}

// Create a new private channel instance.
func NewPrivateChannel(_options *options.Config) *PrivateChannel {
	pch := &PrivateChannel{}
	pch.options = _options
	pch.client = _http.NewClient()
	return pch
}

// Send authentication request to application server.
func (pch *PrivateChannel) Authenticate(_socket *socket.Socket, data *types.Data) (any, int, error) {
	body, err := json.Marshal(map[string]string{
		"channel_name": data.Channel,
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	data.Auth.Headers["Content-Type"] = "application/json; charset=UTF-8"
	options := &_http.Options{
		Method:  http.MethodPost,
		Headers: data.Auth.Headers,
		Url:     pch.authHost(_socket) + pch.options.AuthEndpoint,
		Body:    bytes.NewReader(body),
	}

	if pch.options.DevMode {
		utils.Log().Warning(`Sending auth request to: %s`, options.Url)
	}

	return pch.serverRequest(_socket, options, data.Channel)
}

// Get the auth host based on the Socket.
func (pch *PrivateChannel) authHost(_socket *socket.Socket) string {
	_authHosts := pch.options.AuthHost
	if _authHosts == nil {
		_authHosts = pch.options.Host
	}
	authHosts := options.Hosts{}
	switch hosts := _authHosts.(type) {
	case string:
		authHosts = options.Hosts{hosts}
	case options.Hosts:
		authHosts = hosts
	}

	authHostSelected := "http://localhost"
	if len(authHosts) > 0 {
		authHostSelected = authHosts[0]
	}

	if r := _socket.Request().Headers().Peek("Referer"); r != "" {
		if referer, err := url.Parse(r); err != nil {
			for _, authHost := range authHosts {
				authHostSelected = authHost
				if pch.hasMatchingHost(referer, authHost) {
					authHostSelected = referer.Scheme + "//" + referer.Host
					break
				}
			}
		}
	}

	if pch.options.DevMode {
		utils.Log().Warning(`Preparing authentication request to: %s`, authHostSelected)
	}

	return authHostSelected
}

// Check if there is a matching auth host.
func (pch *PrivateChannel) hasMatchingHost(referer *url.URL, host string) bool {
	hostname := referer.Hostname()
	return (hostname != "" && hostname[strings.Index(hostname, `.`):] == host) || (referer.Scheme+"//"+referer.Host) == host || referer.Host == host
}

// Send a request to the server.
func (pch *PrivateChannel) serverRequest(_socket *socket.Socket, options *_http.Options, channel_name string) (any, int, error) {
	options.Headers = pch.prepareHeaders(_socket, options)
	response, err := pch.client.Request(options)
	if err != nil {
		if pch.options.DevMode {
			utils.Log().Error(`Error authenticating %s for %s`, _socket.Id(), channel_name)
			utils.Log().Error("%v", err)
		}
		return nil, http.StatusBadGateway, errors.New("Error sending authentication request.")
	}
	if response.StatusCode != http.StatusOK {
		if pch.options.DevMode {
			utils.Log().Warning(`%s could not be authenticated to %s`, _socket.Id(), channel_name)
			utils.Log().Error("%s", response.BodyBuffer.String())
		}
		return nil, response.StatusCode, errors.New(fmt.Sprintf(`Client can not be authenticated, got HTTP status %d`, response.StatusCode))
	}
	if pch.options.DevMode {
		utils.Log().Info(`%s authenticated for: %s`, _socket.Id(), channel_name)
	}
	if response.BodyBuffer == nil {
		return nil, http.StatusBadGateway, errors.New("Error sending authentication request.")
	}
	var res_channel_data *types.AuthenticateData = nil
	if err := json.Unmarshal(response.BodyBuffer.Bytes(), &res_channel_data); err != nil {
		var res_bool bool
		if err := json.Unmarshal(response.BodyBuffer.Bytes(), &res_bool); err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return res_bool, response.StatusCode, nil
	}

	return res_channel_data, response.StatusCode, nil
}

// Prepare headers for request to app server.
func (pch *PrivateChannel) prepareHeaders(_socket *socket.Socket, options *_http.Options) map[string]string {
	if cookie, HasCookie := options.Headers[`Cookie`]; !HasCookie || cookie == "" {
		if c := _socket.Request().Headers().Peek("Cookie"); c != "" {
			options.Headers[`Cookie`] = c
		}
	}
	options.Headers[`X-Requested-With`] = `XMLHttpRequest`

	return options.Headers
}
