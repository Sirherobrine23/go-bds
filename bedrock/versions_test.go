package bedrock

import (
	"encoding/json"
	"testing"
)

func TestVersions(t *testing.T) {
	versions, err := FromVersions()
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Get latest", func(t *testing.T) {
		t.Logf("Latest version is %q", versions.GetLatest())
	})
	t.Run("Get latest preview", func(t *testing.T) {
		t.Logf("Latest preview version is %q", versions.GetLatestPreview())
	})
}

func TestFromMojang(t *testing.T) {
	data, err := FetchFromWebsite()
	if err != nil {
		t.Error(err)
		return
	}
	s, _ := json.MarshalIndent(data, "", "  ")
	t.Log(string(s))
}
