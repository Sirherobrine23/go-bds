package request

import (
	"net/http"

	"sirherobrine23.com.br/go-bds/go-bds/request/gohtml"
)

// Make request and parse HTML pages/content
func GoHTML[T any](Url string, Option *Options) (v T, res *http.Response, err error) {
	res, err = DoGoHTML(Url, &v, Option)
	return v, res, err
}

// Make request and append data to current struct
func DoGoHTML(Url string, v any, Option *Options) (*http.Response, error) {
	res, err := Request(Url, Option)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()
	return res, gohtml.NewDecode(res.Body, v)
}
