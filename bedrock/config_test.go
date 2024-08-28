package mojang

import (
	"testing"
)

func TestConfig(t *testing.T) {
	var conf = new(MojangConfig)
	conf.Gamemode = "survival"
	conf.Difficulty = "normal"
	conf.DefaultPlayerPermission = "member"
	conf.TickDistance = 12

	if err := conf.Check(); err != nil {
		t.Fatal(err)
	}
}
