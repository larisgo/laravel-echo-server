package express

import (
	"io"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/zishang520/engine.io/types"
)

type Next func(http.ResponseWriter, *http.Request, func())

type Express struct {
	*types.ServeMux

	// The http router.
	router *httprouter.Router

	// The http middlewares.
	middlewares []Next

	// Configurable server options.
	options *options.Config

	mu sync.RWMutex
}

// Create a new Express instance.
func NewExpress(_options *options.Config) *Express {
	es := &Express{}
	es.router = httprouter.New()
	es.ServeMux = types.NewServeMux(es.router)
	es.middlewares = []Next{}
	es.options = _options
	return es
}

func (es *Express) Route() *httprouter.Router {
	return es.router
}

func (es *Express) Use(middlewares ...Next) {
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, middleware := range middlewares {
		es.middlewares = append(es.middlewares, middleware)
	}
}

func (es *Express) next(w http.ResponseWriter, r *http.Request, index int, length int) func() {
	return func() {
		if index < length {
			es.middlewares[index](w, r, es.next(w, r, index+1, length))
		} else {
			es.ServeMux.ServeHTTP(w, r)
		}
	}
}

func (es *Express) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if length := len(es.middlewares); length > 0 {
		es.next(w, r, 0, length)()
	} else {
		es.ServeMux.ServeHTTP(w, r)
	}
}

// Attach global protection to HTTP routes, to verify the API key.
func (es *Express) AuthorizeRequests(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
		// Get the Basic Authentication credentials
		if es.CanAccess(r, router) {
			handle(w, r, router)
		} else {
			es.UnauthorizedResponse(w, r)
		}
	}
}

// Check is an incoming r can access the api.
func (es *Express) CanAccess(r *http.Request, router httprouter.Params) bool {
	appId := es.GetAppId(router)
	key := es.GetAuthKey(r)

	if appId != "" && key != "" {
		for _, client := range es.options.Clients {
			if client.AppId == appId {
				return client.Key == key
			}
		}
	}

	return false
}

// Get the appId from the URL
func (es *Express) GetAppId(router httprouter.Params) string {
	if appId := router.ByName("appId"); appId != "" {
		return appId
	}
	return ""
}

// Get the api token from the r.
func (es *Express) GetAuthKey(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		return auth[7:]
	}
	if auth_key := r.URL.Query().Get("auth_key"); auth_key != "" {
		return auth_key
	}
	return ""
}

// Handle unauthorized rs.
func (es *Express) UnauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	io.WriteString(w, `{"error":"Unauthorized"}`)
}
