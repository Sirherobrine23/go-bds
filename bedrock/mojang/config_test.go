package mojang_test

import (
	"testing"

	"sirherobrine23.org/go-bds/go-bds/bedrock/mojang"
)

func TestConfig(t *testing.T) {
	var conf = new(mojang.MojangConfig)
	conf.Gamemode = "survival"
	conf.Difficulty = "normal"
	conf.TickDistance = 12

	if err := conf.Check(); err != nil {
		t.Fatal(err)
	}
}