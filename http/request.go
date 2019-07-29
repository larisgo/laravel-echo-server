package http

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Response struct {
	*http.Response
	BodyBytes  []byte
	BodyLength int64
}

type Options struct {
	Method  string
	Url     string
	Headers map[string]string
	Timeout int
	Body    io.Reader
}

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (this *Client) extractBody(r io.Reader) (int64, []byte, io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	length, err := buf.ReadFrom(r)
	return length, buf.Bytes(), ioutil.NopCloser(buf), err
}

func (this *Client) Request(options *Options) (res *Response, err error) {
	// if options == nil {
	// 	options = &Options{}
	// }
	client := &http.Client{}
	if options.Timeout <= 0 {
		client.Timeout = 30 * time.Second
	} else {
		client.Timeout = time.Duration(options.Timeout) * time.Second
	}
	request, err := http.NewRequest(strings.ToUpper(options.Method), options.Url, options.Body)
	if err != nil {
		return nil, err
	}
	for key, value := range options.Headers {
		request.Header.Set(key, value)
	}
	if _, HasContentType := request.Header["Content-Type"]; options.Body != nil && !HasContentType {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	res = &Response{Response: response}
	// apparently, Body can be nil in some cases
	if response.Body != nil {
		// 解压gzio
		if _, HasContentEncoding := response.Header["Content-Encoding"]; HasContentEncoding && response.Header.Get("Content-Encoding") == "gzip" {
			response.Body, err = gzip.NewReader(response.Body)
			if err != nil {
				return nil, err
			}
		}
		res.BodyLength, res.BodyBytes, res.Body, err = this.extractBody(response.Body)
		if err != nil {
			return nil, err
		}
	} else {
		res.BodyLength = 0
		res.BodyBytes = []byte{}
	}
	return res, nil
}

func (this *Client) Get(url string, headers ...map[string]string) (res *Response, err error) {
	headers = append(headers, map[string]string{})
	return this.Request(&Options{
		Method:  "GET",
		Url:     url,
		Headers: headers[0],
	})
}

func (this *Client) Post(url string, body io.Reader, headers ...map[string]string) (res *Response, err error) {
	headers = append(headers, map[string]string{})
	return this.Request(&Options{
		Method:  "POST",
		Url:     url,
		Headers: headers[0],
		Body:    body,
	})
}
