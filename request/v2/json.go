package request

import (
	"encoding/json"
	"net/http"
)

// Make request and process Body in target
func DoJSON(Url string, v any, Option *Options) (*http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()
	return res, json.NewDecoder(res.Body).Decode(v)
}

// Make request and return struct body
func JSON[T any](Url string, Option *Options) (v T, res *http.Response, err error) {
	res, err = DoJSON(Url, &v, Option)
	return
}
