package express

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/larisgo/laravel-echo-server/options"
	"net"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
)

type Next func(http.ResponseWriter, *http.Request, func())

type muxEntry struct {
	h       http.Handler
	pattern string
}

type Express struct {
	/**
	 * The http router.
	 *
	 * @type {*httprouter.Router}
	 */
	router *httprouter.Router

	/**
	 * The http middlewares.
	 *
	 * @type {[]Next}
	 */
	middlewares []Next

	/**
	 * Configurable server options.
	 */
	options options.Config

	// mu
	mu sync.RWMutex

	m     map[string]muxEntry
	es    []muxEntry // slice of entries sorted from longest to shortest.
	hosts bool       // whether any patterns contain hostnames
}

/**
 * Create a new Express instance.
 */
func NewExpress(Options options.Config) *Express {
	this := &Express{}
	this.router = httprouter.New()
	this.middlewares = []Next{}
	this.options = Options
	return this
}

/**
 * [func description]
 * @Author    ZiShang520@gmail.com
 * @DateTime  2019-07-23T11:17:32+0800
 * @copyright (c) ZiShang520 All Rights Reserved
 * @return    {*httprouter.Router}
 */
func (this *Express) Route() *httprouter.Router {
	return this.router
}

/**
 * [func description]
 * @Author    ZiShang520@gmail.com
 * @DateTime  2019-07-23T11:17:32+0800
 * @copyright (c) ZiShang520 All Rights Reserved
 * @param     {func(http.ResponseWriter, *http.Request, func())} middlewares
 */
func (this *Express) Use(middlewares ...Next) {
	this.mu.Lock()
	defer this.mu.Unlock()

	for _, middleware := range middlewares {
		this.middlewares = append(this.middlewares, middleware)
	}
}

/**
 * [func description]
 * @Author    ZiShang520@gmail.com
 * @DateTime  2019-07-23T11:17:27+0800
 * @copyright (c) ZiShang520 All Rights Reserved
 * @param     {http.ResponseWriter} w
 * @param     {*http.Request} r
 * @param     {*int} index
 * @param     {int} length
 * @return    {func()}
 */
func (this *Express) next(w http.ResponseWriter, r *http.Request, index *int, length int) func() {
	return func() {
		if *index = *index + 1; *index < length {
			this.middlewares[*index](w, r, this.next(w, r, index, length))
		} else {
			this.MuxServeHTTP(w, r)
		}
	}
}

/**
 * [func description]
 * @Author    ZiShang520@gmail.com
 * @DateTime  2019-07-23T11:17:19+0800
 * @copyright (c) ZiShang520 All Rights Reserved
 * @param     {http.ResponseWriter} w
 * @param     {*http.Request} r
 */
func (this *Express) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if length := len(this.middlewares); length > 0 {
		index := 0
		this.middlewares[index](w, r, this.next(w, r, &index, length))
	} else {
		this.MuxServeHTTP(w, r)
	}
}

func (this *Express) match(path string) (h http.Handler, pattern string) {
	v, ok := this.m[path]
	if ok {
		return v.h, v.pattern
	}
	for _, e := range this.es {
		if strings.HasPrefix(path, e.pattern) {
			return e.h, e.pattern
		}
	}
	return nil, ""
}

func (this *Express) redirectToPathSlash(host, path string, u *url.URL) (*url.URL, bool) {
	this.mu.RLock()
	shouldRedirect := this.shouldRedirectRLocked(host, path)
	this.mu.RUnlock()
	if !shouldRedirect {
		return u, false
	}
	path = path + "/"
	u = &url.URL{Path: path, RawQuery: u.RawQuery}
	return u, true
}

func (this *Express) shouldRedirectRLocked(host, path string) bool {
	p := []string{path, host + path}

	for _, c := range p {
		if _, exist := this.m[c]; exist {
			return false
		}
	}

	n := len(path)
	if n == 0 {
		return false
	}
	for _, c := range p {
		if _, exist := this.m[c+"/"]; exist {
			return path[n-1] != '/'
		}
	}

	return false
}

func (this *Express) Handler(r *http.Request) (h http.Handler, pattern string) {
	if r.Method == "CONNECT" {
		if u, ok := this.redirectToPathSlash(r.URL.Host, r.URL.Path, r.URL); ok {
			return http.RedirectHandler(u.String(), http.StatusMovedPermanently), u.Path
		}

		return this.handler(r.Host, r.URL.Path)
	}

	host := stripHostPort(r.Host)
	path := cleanPath(r.URL.Path)

	if u, ok := this.redirectToPathSlash(host, path, r.URL); ok {
		return http.RedirectHandler(u.String(), http.StatusMovedPermanently), u.Path
	}

	if path != r.URL.Path {
		_, pattern = this.handler(host, path)
		url := *r.URL
		url.Path = path
		return http.RedirectHandler(url.String(), http.StatusMovedPermanently), pattern
	}

	return this.handler(host, r.URL.Path)
}

func (this *Express) handler(host, path string) (h http.Handler, pattern string) {
	this.mu.RLock()
	defer this.mu.RUnlock()

	if this.hosts {
		h, pattern = this.match(host + path)
	}
	if h == nil {
		h, pattern = this.match(path)
	}
	if h == nil {
		h, pattern = this.router, ""
	}
	return h, pattern
}

func (this *Express) MuxServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	handler, _ := this.Handler(r)
	handler.ServeHTTP(w, r)
}

func (this *Express) Handle(pattern string, handler http.Handler) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if pattern == "" {
		panic("http: invalid pattern")
	}
	if handler == nil {
		panic("http: nil handler")
	}
	if _, exist := this.m[pattern]; exist {
		panic("http: multiple registrations for " + pattern)
	}

	if this.m == nil {
		this.m = make(map[string]muxEntry)
	}
	e := muxEntry{h: handler, pattern: pattern}
	this.m[pattern] = e
	if pattern[len(pattern)-1] == '/' {
		this.es = appendSorted(this.es, e)
	}

	if pattern[0] != '/' {
		this.hosts = true
	}
}

/**
 * Attach global protection to HTTP routes, to verify the API key.
 */
func (this *Express) AuthorizeRequests(Handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, router httprouter.Params) {
		// Get the Basic Authentication credentials
		if this.CanAccess(r, router) {
			Handle(w, r, router)
		} else {
			this.UnauthorizedResponse(w, r)
		}
	}
}

/**
 * Check is an incoming r can access the api.
 *
 * @param  {*http.Request} r
 * @param  {httprouter.Params} router
 * @return {boolean}
 */
func (this *Express) CanAccess(r *http.Request, router httprouter.Params) bool {
	appId := this.GetAppId(router)
	key := this.GetAuthKey(r)

	if appId != "" && key != "" {
		for _, client := range this.options.Clients {
			if client.AppId == appId {
				return client.Key == key
			}
		}
	}

	return false
}

/**
 * Get the appId from the URL
 *
 * @param  {httprouter.Params} router
 * @return {string}
 */
func (this *Express) GetAppId(router httprouter.Params) string {
	if appId := router.ByName("appId"); appId != "" {
		return appId
	}
	return ""
}

/**
 * Get the api token from the r.
 *
 * @param  {*http.Request} r
 * @return {string}
 */
func (this *Express) GetAuthKey(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		return auth[7:]
	}
	if auth_key := r.URL.Query().Get("auth_key"); auth_key != "" {
		return auth_key
	}
	return ""
}

/**
 * Handle unauthorized rs.
 *
 * @param  {http.ResponseWriter} w
 * @param  {*http.Request} r
 * @return {void}
 */
func (this *Express) UnauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(403)
	fmt.Fprint(w, `{"error": "Unauthorized"}`)
}

func appendSorted(es []muxEntry, e muxEntry) []muxEntry {
	n := len(es)
	i := sort.Search(n, func(i int) bool {
		return len(es[i].pattern) < len(e.pattern)
	})
	if i == n {
		return append(es, e)
	}
	es = append(es, muxEntry{})
	copy(es[i+1:], es[i:])
	es[i] = e
	return es
}

func stripHostPort(h string) string {
	if strings.IndexByte(h, ':') == -1 {
		return h
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h
	}
	return host
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}
