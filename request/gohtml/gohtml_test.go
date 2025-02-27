package gohtml

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

var HtmlStruct1 string = `<html>
<body>
	<div>
		<a href="https://google.com">2003-01-19T18:50:59.441Z</a>
	</div>
	<div id="float">
		<span>0.1000009</span>
	</div>
	<div id="slice">
		<div>
			<span>Hello</span>
		</div>
		<div>
			<span>Hello 2</span>
		</div>
		<div>
			<span>Create</span>
		</div>
	</div>
</body>
</html>`

func TestHtml(t *testing.T) {
	var body struct {
		Date     time.Time `json:"date" html:"body > div:nth-child(1) > a"`
		Href     string    `json:"href" html:"div > a = href"`
		Float    float32   `json:"float" html:"#float > span"`
		Splited  string    `json:"splited" html:"div[id=float] > span"`
		NilPoint *struct{} `json:"nil_point" html:"cannot"`
		Slice    []struct {
			Text string `json:"text" html:"span"`
			Self string `json:"__self" html:"__self"`
		} `json:"slices" html:"#slice > div"`
		Array [6]struct {
			Text string `json:"text" html:"span"`
			Self string `json:"__self" html:"__self"`
		} `json:"array" html:"#slice > div"`
	}

	if err := Unmarshal([]byte(HtmlStruct1), &body); err != nil {
		t.Error(err)
		return
	}

	d, _ := json.MarshalIndent(body, "", "  ")
	t.Log(string(d))

	if body.Href != "https://google.com" {
		t.Errorf("invalid html destruct")
		return
	} else if len(body.Slice) != 3 {
		data := []string{
			"Hello",
			"Hello 2",
			"Create",
		}
		for _, content := range body.Slice {
			if !slices.Contains(data, content.Text) {
				t.Errorf("Text is invalid, accepts %v, returned %q", data, content.Text)
				return
			}
		}
	}
}
