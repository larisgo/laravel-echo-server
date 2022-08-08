package http

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"io"
	"net/http"
	"strings"
	"time"
)

type Response struct {
	*http.Response
	BodyBuffer *bytes.Buffer
}

type Options struct {
	Method  string
	Url     string
	Headers map[string]string
	Timeout time.Duration
	Body    io.Reader
}

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Request(options *Options) (res *Response, _ error) {
	if options == nil {
		options = &Options{}
	}
	client := &http.Client{}
	if options.Timeout == 0 {
		client.Timeout = 30 * time.Second
	} else {
		client.Timeout = options.Timeout
	}
	request, err := http.NewRequest(strings.ToUpper(options.Method), options.Url, options.Body)
	if err != nil {
		return nil, err
	}
	if options.Headers != nil {
		for key, value := range options.Headers {
			request.Header.Set(key, value)
		}
	}
	if _, HasContentType := request.Header["Content-Type"]; options.Body != nil && !HasContentType {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	res = &Response{Response: response}

	// apparently, Body can be nil in some cases
	if response.Body != nil {
		defer response.Body.Close()

		body := bytes.NewBuffer(nil)
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			gz, err := gzip.NewReader(response.Body)
			if err != nil {
				return nil, err
			}
			defer gz.Close()
			io.Copy(body, gz)
			response.Header.Del("Content-Encoding")
			response.Header.Del("Content-Length")
			response.ContentLength = -1
			response.Uncompressed = true
		case "deflate":
			fl := flate.NewReader(response.Body)
			defer fl.Close()
			io.Copy(body, fl)
			response.Header.Del("Content-Encoding")
			response.Header.Del("Content-Length")
			response.ContentLength = -1
			response.Uncompressed = true
		case "br":
			br := brotli.NewReader(response.Body)
			io.Copy(body, br)
			response.Header.Del("Content-Encoding")
			response.Header.Del("Content-Length")
			response.ContentLength = -1
			response.Uncompressed = true
		default:
			io.Copy(body, response.Body)
		}
		res.BodyBuffer = body
	} else {
		res.BodyBuffer = nil
	}
	return res, nil
}

func (c *Client) Get(url string, header map[string]string) (*Response, error) {
	return c.Request(&Options{
		Method:  http.MethodGet,
		Url:     url,
		Headers: header,
	})
}

func (c *Client) Post(url string, body io.Reader, header map[string]string) (*Response, error) {
	return c.Request(&Options{
		Method:  http.MethodPost,
		Url:     url,
		Headers: header,
		Body:    body,
	})
}
