package request

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var codeInt int
		fmt.Sscan(strings.Split(r.URL.Path, "/")[1], &codeInt)
		w.WriteHeader(codeInt)
		w.Write(nil)
	})
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Errorf("http listen: %s", err)
		return
	}
	defer ln.Close()
	go http.Serve(ln, mux)

	for _, requestData := range [][]any{
		{200, "GET"},
		{200, "POST"},
		{200, "DELETE"},
		{300, "GET"},
		{300, "POST"},
		{300, "DELETE"},
		{404, "GET"},
		{404, "POST"},
		{404, "DELETE"},
	} {
		var codeInt int = requestData[0].(int)
		var method string = requestData[1].(string)
		t.Run(fmt.Sprintf("%s %d", method, codeInt), func(t *testing.T) {
			res, err := Request(fmt.Sprintf("http://%s/%d", ln.Addr().String(), codeInt), &Options{Method: method})
			if err != nil {
				if nerr, ok := err.(errResponseCode); ok {
					if nerr.Response.StatusCode == codeInt {
						return
					}
				}
				t.Error(err)
			} else if res.StatusCode != codeInt {
				t.Error("Invalid request response")
			}
		})
	}
}
