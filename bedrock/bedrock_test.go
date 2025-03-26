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

	t.Run("Preview", func(t *testing.T) {
		ver := versions.LatestStable()
		if ver == nil {
			t.Error("cannot get latest minecraft version")
			return
		}
		d, _ := json.MarshalIndent(ver, "", "  ")
		t.Log(string(d))
	})

	t.Run("Latest", func(t *testing.T) {
		ver := versions.LatestPreview()
		if ver == nil {
			t.Error("cannot get latest minecraft preview version")
			return
		}
		d, _ := json.MarshalIndent(ver, "", "  ")
		t.Log(string(d))
	})
}
