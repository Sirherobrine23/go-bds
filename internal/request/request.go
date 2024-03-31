package request

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/html"
)

func GetJson(url string, body interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return err;
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return err;
	}

	return nil
}

func GetHtml(url string) (*html.Node, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err;
	}
	
	defer res.Body.Close()
	return html.Parse(res.Body)
}