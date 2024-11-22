package gohtml

import (
	"slices"
	"testing"
)

var HtmlStruct1 string = `<html>
<body>
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
	<div>
		<a href="https://google.com"></a>
	</div>
</body>
</html>`

func TestHtml(t *testing.T) {
	var body struct {
		Href string `html:"div > a = href"`
		Text []struct {
			Text string `html:"span"`
		} `html:"div > div"`
	}

	if err := Parse([]byte(HtmlStruct1), &body); err != nil {
		t.Error(err)
		return
	}

	if body.Href != "https://google.com" {
		t.Errorf("invalid html destruct")
		return
	} else if len(body.Text) != 3 {
		data := []string{
			"Hello",
			"Hello 2",
			"Create",
		}
		for _, content := range body.Text {
			if !slices.Contains(data, content.Text) {
				t.Errorf("Text is invalid, accepts %v, returned %q", data, content.Text)
				return
			}
		}
	}
}
