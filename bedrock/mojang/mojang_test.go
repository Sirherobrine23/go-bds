package mojang

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMojang(t *testing.T) {
	testRoot, err := os.MkdirTemp(os.TempDir(), "testMCBds*")
	if err != nil {
		t.Skip("cannot make temp folder")
		return
	}
	defer os.RemoveAll(testRoot)

	var bedrock Mojang
	bedrock.VersionsFolder = filepath.Join(testRoot, "versions")
	bedrock.Path = filepath.Join(testRoot, "bdsdata")
	bedrock.Version = "latest"
	if err := bedrock.Start(); err != nil {
		t.Fatal(err)
		return
	}

	<-time.After(time.Second * 5)
	if err := bedrock.Close(); err != nil {
		t.Fatal(err)
		return
	}
}
