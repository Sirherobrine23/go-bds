package bedrock

import (
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
	if _, err := FetchFromWebsite(); err != nil {
		t.Error(err)
		return
	}
}
