package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	// New MapCode test
	_ MapCode = map[int]RequestStatusFunction{}
	_ MapCode = MapCode{}

	ErrCode error = &ResponseError{}
)

// Custom error for http request
type ResponseError struct {
	*http.Response // HTTP Response
	*http.Request  // HTTP Request
}

func (err ResponseError) Error() string {
	return fmt.Sprintf("response from %q, return code status %03d", err.Request.URL.String(), err.Response.StatusCode)
}

// Function callback to process http code status
//
// Deprecated: use [sirherobrine23.com.br/go-bds/go-bds/request/v2.RequestStatusFunction]
type CodeCallback RequestStatusFunction

// Function to request call to process request Status
type RequestStatusFunction func(res *http.Response) (*http.Response, error)

// Map struct to process http code requests
type MapCode map[int]RequestStatusFunction

func (mapCode MapCode) Code(httpCode int) RequestStatusFunction {
	if mapCode == nil {
		return nil
	}
	return mapCode[httpCode]
}

func (mapCode MapCode) HasCode(httpCode int) bool {
	if mapCode == nil {
		return false
	}
	return mapCode[httpCode] != nil && slices.Contains(slices.Collect(maps.Keys(mapCode)), httpCode)
}

// Exists -1 in code
func (mapCode MapCode) HasAll() bool {
	return mapCode.HasCode(-1)
}

// HTTP headers
type Header map[string]string

func (Headers Header) ToHTTP() http.Header {
	httpHeader := http.Header{}
	for key, value := range Headers {
		httpHeader.Set(key, value)
	}
	return httpHeader
}

func (Headers Header) MergeHTTP(ToMerge http.Header) Header {
	if Headers == nil {
		Headers = map[string]string{}
	} else if ToMerge == nil {
		return Headers
	}
	for key := range ToMerge {
		Headers[key] = ToMerge.Get(key)
	}
	return Headers
}

func (Headers Header) Merge(ToMerge Header) Header {
	if Headers == nil {
		return ToMerge
	} else if ToMerge == nil {
		ToMerge = map[string]string{}
	}
	maps.Copy(Headers, ToMerge)
	return Headers
}

// Request options
type Options struct {
	Method            string         `json:"method"`           // Request method, example, Get, Post
	Body              any            `json:"body"`             // nil, io.Reader, another structs converted to json
	Header            Header         `json:"headers"`          // Extra request Headers
	CodeProcess       MapCode        `json:"-"`                // Process request with callback, -1 call any other status code
	NotFollowRedirect bool           `json:"follow_redirects"` // Watcher requests redirects
	Jar               http.CookieJar `json:"-"`                // Cookies storage
	Client            *http.Client   `json:"-"`                // HTTP Client, if nil use default client
}

// Create request and return response
func Request(Url string, Option *Options) (*http.Response, error) {
	requestUrl, err := url.Parse(Url)
	if err != nil {
		return nil, err
	}
	return MakeRequestWithStatus(requestUrl, Option)
}

// Make raw request without process status code
func MakeRequest(Url *url.URL, requestOptions *Options) (*http.Response, error) {
	if requestOptions == nil {
		requestOptions = &Options{}
	}

	// Process body request
	body, err := io.Reader(nil), error(nil)
	switch v := requestOptions.Body.(type) {
	case nil:
		body = http.NoBody // No body
	case []byte:
		body = bytes.NewReader(v) // Convert to reader
	case io.Reader:
		body = v // Only copy
	default:
		// Convert to JSON
		data, err := json.MarshalIndent(requestOptions.Body, "", "  ")
		if err != nil {
			return nil, err
		}
		// Add buffer reader
		body = bytes.NewReader(data)
	}

	// Set method request
	methodRequest, request := "", (*http.Request)(nil)
	if methodRequest = strings.ToUpper(requestOptions.Method); methodRequest == "" {
		methodRequest = "GET"
	}

	// Create new request
	if request, err = http.NewRequest(methodRequest, Url.String(), body); err != nil {
		return nil, err
	}

	// Set headers
	request.Header = requestOptions.Header.MergeHTTP(request.Header).ToHTTP()

	// HTTP Client
	client := &http.Client{Transport: http.DefaultTransport, CheckRedirect: http.DefaultClient.CheckRedirect, Jar: http.DefaultClient.Jar}
	if requestOptions.Client != nil {
		client = requestOptions.Client
	}

	// Client custom cookies jar
	if requestOptions.Jar != nil {
		client.Jar = requestOptions.Jar
	}

	// Don't follow request's redirect
	if requestOptions.NotFollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Process only fist request
		}
	}
	return client.Do(request)
}

// Make request and process status code
func MakeRequestWithStatus(Url *url.URL, requestOptions *Options) (*http.Response, error) {
	response, err := MakeRequest(Url, requestOptions)
	if err != nil {
		return response, err
	}

	if codeStatus := response.StatusCode; requestOptions != nil {
		if requestOptions.CodeProcess.HasCode(codeStatus) {
			return requestOptions.CodeProcess.Code(codeStatus)(response)
		} else if requestOptions.CodeProcess.HasAll() {
			return requestOptions.CodeProcess.Code(-1)(response)
		}
	}

	if code := response.StatusCode; code >= 100 && code <= 399 {
		return response, nil
	}
	return response, &ResponseError{response, response.Request}
}

// Make request and copy all data from response to [*bytes.Buffer]
func Buffer(Url string, Option *Options) (*bytes.Buffer, *http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return nil, res, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res, err
	}
	return bytes.NewBuffer(data), res, err
}

// Make request and save response in Disk
func SaveAs(Url, fileTarget string, Option *Options) (*http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	// Create folder if not exists
	if _, err := os.Stat(filepath.Dir(fileTarget)); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(fileTarget), 0600); err != nil {
			return res, err
		}
	}

	file, err := os.Create(fileTarget)
	if err != nil {
		return res, err
	}
	defer file.Close()
	_, err = io.Copy(file, res.Body)
	return res, err
}

// Make request and save response in temporary file
func SaveTmp(Url, dir string, Option *Options) (*os.File, *http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return nil, res, err
	}
	defer res.Body.Close()

	// Create temp file in system temporary files directory
	local, err := os.CreateTemp(dir, "gobds*requestfile")
	if err != nil {
		return nil, nil, err
	}

	// Copy body
	if _, err := io.Copy(local, res.Body); err != nil {
		return local, res, err
	} else if _, err = local.Seek(0, 0); err != nil {
		return local, res, err
	}

	return local, res, nil
}
