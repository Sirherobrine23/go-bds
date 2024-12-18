package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sirherobrine23.com.br/go-bds/go-bds/request/v2"
)

// Generic playit.gg body process
type playitResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

// Process requests to api.playit.gg
func requestAPI[T any](w Api, Path string, Body any, headers request.Header) (T, *http.Response, error) {
	res, err := request.Request(fmt.Sprintf("%s%s", PlayitAPI, Path), &request.Options{
		Method: "POST",
		Body:   Body,
		Header: headers.Merge(request.Header{"Authorization": fmt.Sprintf("Agent-Key %s", w.Secret)}),
		CodeProcess: map[int]request.CodeCallback{
			200: func(res *http.Response) (*http.Response, error) { return res, nil },
			201: func(res *http.Response) (*http.Response, error) { return res, nil },
			-1: func(res *http.Response) (*http.Response, error) {
				defer res.Body.Close()
				var errSta errStatus
				if err := json.NewDecoder(res.Body).Decode(&errSta); err != nil {
					return nil, err
				}
				return nil, errSta.Error()
			},
		},
	})
	if err != nil {
		return *(*T)(nil), res, err
	}
	defer res.Body.Close()
	var bodyProcess playitResponse[T]
	if err := json.NewDecoder(res.Body).Decode(&bodyProcess); err != nil {
		return *(*T)(nil), nil, err
	}
	return bodyProcess.Data, res, nil
}
