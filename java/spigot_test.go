package java

import (
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
)

func TestListFromSpigotmc(t *testing.T) {
	versions, err := ListFromSpigotmc()
	if err != nil {
		t.Error(err)
		return
	}

	versionsJSON, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(versionsJSON))
}

func TestBuild(t *testing.T) {
	// Static Version
	version := SpigotMC{
		Version:      "1.21.1",
		ToolVersion:  181,
		JavaVersions: []uint{65, 67},
		Ref: struct {
			Spigot      string `json:"Spigot"`
			Bukkit      string `json:"Bukkit"`
			CraftBukkit string `json:"CraftBukkit"`
		}{
			Spigot:      "a759b629cbf86401aab56b8c3f21a635e9e76c15",
			Bukkit:      "bb4e97c60d2978a1d008f21295a5234228341e14",
			CraftBukkit: "0a7bd6c81a33cfaaa2f4d2456c6b237792f38fe6",
		},
	}

	if err := version.Build(filepath.Join(t.TempDir(), "build"), filepath.Join(t.TempDir(), "output"), io.Discard); err != nil {
		t.Error(err)
		return
	}
}
