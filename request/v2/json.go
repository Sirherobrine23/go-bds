package request

import (
	"encoding/json"
	"net/http"
)

// Make request and return struct body
func JSON[T any](Url string, Option *Options) (T, *http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return *new(T), res, err
	}
	defer res.Body.Close()

	var data T // Data
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return *new(T), res, err
	}
	return data, res, err
}

// Make request and process Body in target var
func JSONDo(Url string, Target any, Option *Options) (*http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()
	return res, json.NewDecoder(res.Body).Decode(Target)
}