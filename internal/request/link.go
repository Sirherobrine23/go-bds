package request

import (
	"strings"
)

type Link struct {
	URL    string
	Params map[string]string
}

func (w *Link) HasParam(key string) bool {
	_, ok := w.Params[key]
	return ok
}

func (w *Link) HasKey(key string) bool {
	_, ok := w.Params[key]
	return ok
}

func (w *Link) HasKeyValue(key string, values ...string) (string, bool) {
	if value, ok := w.Params[key]; ok {
		for _, inputValue := range values {
			if value == inputValue {
				return inputValue, true
			}
		}
	}
	return "", false
}

func (w *Link) Param(key string) string {
	if k, ok := w.Params[key]; ok {
		return k
	}
	return ""
}

func ParseLink(header string) []Link {
	link := []Link{}
	for _, multi := range strings.Split(header, `,`) {
		multi = strings.TrimSpace(multi)
		linker := Link{URL: "", Params: map[string]string{}}
		for _, ks := range strings.Split(multi, ";") {
			ks = strings.TrimSpace(ks)
			if strings.HasPrefix(ks, "<") && strings.HasSuffix(ks, ">") {
				linker.URL = strings.Trim(ks, "<>")
				continue
			}

			if len(ks) == 0 {
				continue
			}
			k2 := strings.SplitN(ks, "=", 2)
			if len(k2) == 1 {
				linker.Params[k2[0]] = ""
			} else if len(k2) == 2 {
				if strings.HasPrefix(k2[1], `"`) && strings.HasSuffix(k2[1], `"`) {
					k2[1] = strings.Trim(k2[1], `"`)
				}
				linker.Params[k2[0]] = k2[1]
			}
		}

		link = append(link, linker)
	}

	return link
}

func ParseMultipleLinks(ks ...string) []Link {
	link := []Link{}
	for _, k := range ks {
		link = append(link, ParseLink(k)...)
	}
	return link
}
