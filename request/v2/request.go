package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

var (
	Err400 error = errors.New("bad request")
	Err404 error = errors.New("page not found")
	Err409 error = errors.New("request conflict")
	Err502 error = errors.New("bad gateway")
)

type CodeCallback func(res *http.Response) (*http.Response, error)

// Request options
type Options struct {
	Method      string               // Request method, example, Get, Post
	Body        any                  // Struct or io.Reader, if is Struct convert to json
	Header      map[string]string    // Extra request Headers
	CodeProcess map[int]CodeCallback // Process request with callback, -1 call any other status code
	Jar         http.CookieJar
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
	if methodRequest = strings.ToUpper(req.Options.Method); methodRequest != "" {
		methodRequest = "GET"
	}

	var err error
	if _, ok := req.Options.Body.(io.Reader); !ok {
		if data, ok := req.Options.Body.([]byte); ok {
			req.Options.Body = bytes.NewReader(data)
		} else if req.Options.Body != nil {
			var data []byte
			if data, err = json.Marshal(req.Options.Body); err != nil {
				return nil, err
			}
			req.Options.Body = bytes.NewReader(data)
		}
	}

	// Create request
	var request *http.Request
	if request, err = http.NewRequest(methodRequest, req.Url.String(), req.Options.Body.(io.Reader)); err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Options.Header {
		request.Header.Set(key, value)
	}

	var client http.Client
	if req.Options.Jar != nil {
		client.Jar = req.Options.Jar
	}
	return client.Do(request)
}

// Make request and process status code
func (req RequestRoot) MakeRequestWithStatus() (*http.Response, error) {
	request, err := req.MakeRequest()
	if err != nil {
		return request, err
	} else if len(req.Options.CodeProcess) > 0 {
		if codeProcess, ok := req.Options.CodeProcess[request.StatusCode]; ok {
			return codeProcess(request)
		} else if slices.Contains(slices.Collect(maps.Keys(req.Options.CodeProcess)), -1) {
			return req.Options.CodeProcess[-1](request)
		}
	}
	switch request.StatusCode {
	case
		http.StatusContinue,
		http.StatusSwitchingProtocols,
		http.StatusProcessing,
		http.StatusEarlyHints,
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent,
		http.StatusPartialContent,
		http.StatusMultiStatus,
		http.StatusAlreadyReported,
		http.StatusIMUsed:
		return request, nil
	case http.StatusConflict:
		return request, Err409
	case http.StatusBadRequest:
		return request, Err400
	case http.StatusBadGateway:
		return request, Err502
	default:
		return request, fmt.Errorf("request code %d (%s)", request.StatusCode, request.Status)
	}
}
