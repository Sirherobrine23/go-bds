package request_test

import (
	"testing"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/request"
)

func TestLink(t *testing.T) {
	links := request.ParseLink(`<https://api.adoptium.net/v3/assets/version/%5B1.0%2C100.0%5D?project=jdk&image_type=jdk&semver=true&heap_size=normal&sort_method=DEFAULT&sort_order=DESC&page=1&page_size=20>; rel="next"`)
	for _, k := range links {
		t.Logf("%+v", k)
	}
}