package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var (
	httpHandler = http.NewServeMux()
	server      = httptest.NewServer(httpHandler)
)

func init() {
	httpHandler.HandleFunc("/{$}", func(res http.ResponseWriter, req *http.Request) {
		responseBody := map[string]any{
			"uri":     req.RequestURI,
			"headers": req.Header,
			"method":  req.Method,
			"body":    struct{}{},
		}
		if req.Method != "GET" {
			switch req.Header.Get("Content-Type") {
			case "application/json":
				var body any
				if err := json.NewDecoder(req.Body).Decode(body); err != nil {
					res.WriteHeader(500)
					res.Write([]byte(err.Error()))
					return
				}
				responseBody["body"] = body
			case "multipart/form-data":
				if err := req.ParseMultipartForm(1024 ^ ^2); err != nil {
					res.WriteHeader(500)
					res.Write([]byte(err.Error()))
					return
				}
				responseBody["body"] = req.MultipartForm
			default:
				req.ParseForm()
				responseBody["body"] = req.Form
			}

		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(200)
		jsonEncode := json.NewEncoder(res)
		jsonEncode.SetIndent("", "  ")
		jsonEncode.Encode(responseBody)
	})

	// HTTP Status code
	httpHandler.HandleFunc("/status/{code}/{requestPath...}", func(res http.ResponseWriter, req *http.Request) {
		httpStatus, err := strconv.Atoi(req.PathValue("code"))
		if err != nil {
			fmt.Printf("client error code convert: %s\n", err)
			res.WriteHeader(400)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(httpStatus) // Write status code
		responseData := json.NewEncoder(res)
		responseData.SetIndent("", "  ")
		responseData.Encode(map[string]any{
			"header": req.Header,
			"path":   req.RequestURI,
			"method": req.Method,
		})
	})
}

func TestRequestRaw(t *testing.T) {
	process := func(t *testing.T, res *http.Response, err error) {
		if err != nil {
			t.Error(err)
			return
		}
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("response data total: %d", len(data))
	}

	t.Run("GET", func(t *testing.T) {
		res, err := Request(server.URL, &Options{Client: server.Client(), Method: "GET"})
		process(t, res, err)
	})

	t.Run("POST", func(t *testing.T) {
		res, err := Request(server.URL, &Options{Client: server.Client(), Method: "POST", Body: map[string]string{"from": "JSON maped"}})
		process(t, res, err)
	})
}

func TestRequestJSON(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		_, _, err := JSON[map[string]any](server.URL, &Options{Client: server.Client(), Method: "GET"})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("POST", func(t *testing.T) {
		_, _, err := JSON[map[string]any](server.URL, &Options{Client: server.Client(), Method: "POST", Body: map[string]string{"from": "JSON maped"}})
		if err != nil {
			t.Error(err)
			return
		}
	})
}
