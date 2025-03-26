package bedrock

import (
	"encoding/json"
	"testing"
)

func TestVersionsFetch(t *testing.T) {
	versions := Versions{}
	if err := versions.FetchFromMinecraftDotNet(); err != nil {
		t.Errorf("cannot get server versions from minecraft.net: %s", err)
		return
	}
	d, _ := json.MarshalIndent(versions, "", "  ")
	t.Log(string(d))
}
