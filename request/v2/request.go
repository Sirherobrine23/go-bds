package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

var ErrCode error = new(errResponseCode)

type CodeCallback func(res *http.Response) (*http.Response, error)

type errResponseCode struct {
	Response *http.Response
}

func (err errResponseCode) Error() string {
	return fmt.Sprintf("cannot process request, response code %d", err.Response.StatusCode)
}

type Header map[string]string

func (Headers Header) Merge(ToMerge Header) Header {
	if Headers == nil {
		Headers = map[string]string{}
	}
	if ToMerge == nil {
		ToMerge = map[string]string{}
	}
	n1 := maps.Clone(Headers)
	for key, val := range ToMerge {
		n1[key] = val
	}
	return n1
}

// Request options
type Options struct {
	Method            string               // Request method, example, Get, Post
	Body              any                  // Struct or io.Reader, if is Struct convert to json
	Header            Header               // Extra request Headers
	CodeProcess       map[int]CodeCallback `json:"-"` // Process request with callback, -1 call any other status code
	NotFollowRedirect bool                 // Watcher requests redirects
	Jar               http.CookieJar       `json:"-"`
}

// Request struct
type RequestRoot struct {
	Url     *url.URL // Request URL
	Options *Options // Request options
}

// Mount request struct
func MountRequest(Url string, Option *Options) (*RequestRoot, error) {
	var requestRaw RequestRoot
	var err error
	if requestRaw.Url, err = url.Parse(Url); err != nil {
		return nil, err
	}
	requestRaw.Options = Option
	return &requestRaw, nil
}

// Create request and return response
func Request(Url string, Option *Options) (*http.Response, error) {
	requestRaw, err := MountRequest(Url, Option)
	if err != nil {
		return nil, err
	}
	return requestRaw.MakeRequestWithStatus()
}

// Make raw request without process status code
func (req RequestRoot) MakeRequest() (*http.Response, error) {
	if req.Options == nil {
		req.Options = &Options{}
	}

	var methodRequest string
	if methodRequest = strings.ToUpper(req.Options.Method); methodRequest == "" {
		methodRequest = "GET"
	}

	var err error
	var body io.Reader
	if req.Options.Body != nil {
		switch v := req.Options.Body.(type) {
		case []byte:
			body = bytes.NewReader(v)
		case io.Reader:
			body = v
		default:
			data, err := json.MarshalIndent(req.Options.Body, "", "  ")
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
		}
	}

	// Create request
	var request *http.Request
	if request, err = http.NewRequest(methodRequest, req.Url.String(), body); err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Options.Header {
		request.Header.Set(key, value)
	}

	var client http.Client

	// Client custom cookies jar
	if req.Options.Jar != nil {
		client.Jar = req.Options.Jar
	}

	// Don't follow request's redirect
	if req.Options.NotFollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Process only fist request
		}
	}
	return client.Do(request)
}

// Make request and process status code
func (req RequestRoot) MakeRequestWithStatus() (*http.Response, error) {
	request, err := req.MakeRequest()
	if err != nil {
		return request, err
	} else if req.Options != nil {
		if req.Options.CodeProcess != nil {
			if codeProcess, ok := req.Options.CodeProcess[request.StatusCode]; ok {
				return codeProcess(request)
			} else if slices.Contains(slices.Collect(maps.Keys(req.Options.CodeProcess)), -1) {
				return req.Options.CodeProcess[-1](request)
			}
		}
	}

	if code := request.StatusCode; code >= 100 && code <= 399 {
		return request, nil
	}
	return request, &errResponseCode{request}
}
