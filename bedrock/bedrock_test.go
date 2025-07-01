package bedrock

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
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

func TestExtractServer(t *testing.T) {
	versions := Versions{}
	if err := versions.FetchFromMinecraftDotNet(); err != nil {
		t.Errorf("cannot get server versions from minecraft.net: %s", err)
		return
	}

	var versionFolder, cwd, upper, workdir string
	upper = filepath.Join(t.TempDir(), "cwd/storage")
	workdir = filepath.Join(t.TempDir(), "cwd/work")
	cwd = filepath.Join(t.TempDir(), "cwd/exec")
	versionFolder = filepath.Join(t.TempDir(), "versions")

	// Fresh install
	_, err := NewBedrock(versions.LatestStable(), versionFolder, cwd, upper, workdir)
	if err != nil {
		t.Errorf("cannot install server: %s", err)
		return
	}

	if file_checker.FolderIsEmpty(filepath.Join(versionFolder, versions.LatestStable().Version)) {
		t.Error("invalid server installation")
	}

	// Replace old install
	_, err = NewBedrock(versions.LatestStable(), versionFolder, cwd, upper, workdir)
	if err != nil {
		t.Errorf("cannot install server: %s", err)
		return
	}

	if file_checker.FolderIsEmpty(filepath.Join(versionFolder, versions.LatestStable().Version)) {
		t.Error("invalid server installation")
	}
}
