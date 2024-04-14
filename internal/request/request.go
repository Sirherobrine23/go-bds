package request

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	ErrNoUrl        = errors.New("no url informed")
	ErrPageNotExist = errors.New("page not exists")
)

var DefaultHeader = http.Header{
	"Accept": {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	// "Accept-Encoding":           {"gzip, deflate"},
	"Accept-Language":           {"en-US;q=0.9,en;q=0.8"},
	"Sec-Ch-Ua":                 {`"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123\"`},
	"Sec-Ch-Ua-Mobile":          {"?0"},
	"Sec-Ch-Ua-Platform":        {`"Windows"`},
	"Sec-Fetch-Dest":            {"document"},
	"Sec-Fetch-Mode":            {"navigate"},
	"Sec-Fetch-Site":            {"none"},
	"Sec-Fetch-User":            {"?1"},
	"Upgrade-Insecure-Requests": {"1"},
	"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"},
}

type RequestOptions struct {
	HttpError  bool        // Return if status code is equal or then 300
	Method     string      // Request Method
	Url        string      // Request url
	Body       io.Reader   // Body Reader
	Headers    http.Header // Extra HTTP Headers
	Querys     map[string]string
}

func (w *RequestOptions) ToUrl() (*url.URL, error) {
	if len(w.Url) == 0 {
		return nil, ErrNoUrl
	}

	urlParsed, err := url.Parse(w.Url)
	if err != nil {
		return nil, err
	}

	if len(w.Querys) > 0 {
		query := urlParsed.Query()
		for key, value := range w.Querys {
			query.Set(key, value)
		}
		urlParsed.RawQuery = query.Encode()
	}

	return urlParsed, nil
}

func (w *RequestOptions) String() (string, error) {
	s, err := w.ToUrl()
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

func (opt *RequestOptions) Request() (http.Response, error) {
	if len(opt.Method) == 0 {
		opt.Method = "GET"
	}

	urlRequest, err := opt.ToUrl()
	if err != nil {
		return http.Response{}, err
	}

	// Create request
	req, err := http.NewRequest(opt.Method, urlRequest.String(), opt.Body)
	if err != nil {
		return http.Response{}, err
	}

	// Project headers
	for key, value := range DefaultHeader {
		req.Header[key] = value
	}

	// Set headers
	for key, value := range opt.Headers {
		req.Header[key] = value
	}

	// Create response from request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return *res, err
	}

	if opt.HttpError && res.StatusCode >= 300 {
		if res.StatusCode == 404 {
			err = ErrPageNotExist
		} else {
			err = fmt.Errorf("response non < 299, code %d, url: %q", res.StatusCode, opt.Url)
		}
	}

	// User tratement
	return *res, err
}

// Make custom request and return request, response and error if exist
func Request(opt RequestOptions) (http.Response, error) {
	return opt.Request()
}

func SaveFile(filePath string, opt RequestOptions) (http.Response, error) {
	res, err := Request(opt)
	if err != nil {
		return res, err
	}

	defer res.Body.Close()
	file, err := os.Create(filePath)
	if err != nil {
		return res, err
	}

	defer file.Close()
	_, err = io.Copy(file, res.Body)
	return res, err
}

func HtmlNode(opt RequestOptions) (*goquery.Document, http.Response, error) {
	res, err := Request(opt)
	if err != nil {
		return &goquery.Document{}, res, err
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	return doc, res, err
}

func RequestHtmlLinks(opt RequestOptions) ([]string, http.Response, error) {
	res, err := Request(opt)
	if err != nil {
		return []string{}, res, err
	}

	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		return []string{}, res, err
	}

	urls := []string{}
	var find func(*html.Node)
	find = func(n *html.Node) {
		for _, v := range n.Attr {
			if v.Key == "src" || v.Key == "href" {
				urls = append(urls, v.Val)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)

	urlParsed, _ := url.Parse(opt.Url)
	for urlIndex, value := range urls {
		if strings.HasPrefix(value, "/") {
			urls[urlIndex] = fmt.Sprintf("%s://%s%s", urlParsed.Scheme, urlParsed.Host, value)
		}
	}

	return urls, res, err
}
