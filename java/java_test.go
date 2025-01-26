package java

import (
	"path/filepath"
	"testing"
)

// Test spigot build to install server
func TestSpigotBuild(t *testing.T) {
	// Static Version
	version := SpigotMC{
		Version:      "1.21.1",
		ToolVersion:  181,
		JavaVersions: []uint{65, 67},
		JavaFolder:   filepath.Join(t.TempDir(), "javacs"),
		Ref: struct {
			Spigot      string `json:"Spigot"`
			Bukkit      string `json:"Bukkit"`
			CraftBukkit string `json:"CraftBukkit"`
		}{"a759b629cbf86401aab56b8c3f21a635e9e76c15", "bb4e97c60d2978a1d008f21295a5234228341e14", "0a7bd6c81a33cfaaa2f4d2456c6b237792f38fe6"},
	}

	// Build and install server
	if err := version.Install(filepath.Join(t.TempDir(), "spigotBuild")); err != nil {
		t.Error(err)
		return
	}
}

func TestListVersions(t *testing.T) {
	for _, paperTarget := range PaperProjects {
		t.Run(paperTarget, func(t *testing.T) {
			call, _ := ListPaper(paperTarget)
			versions, err := call()
			if err != nil {
				t.Error(err)
				return
			} else if len(versions) == 0 {
				t.Errorf("cannot get versions to %s in paper", paperTarget)
			}
		})
	}

	if mcVersion, err := ListSpigot("")(); err != nil {
		t.Error(err)
		return
	} else if len(mcVersion) == 0 {
		t.Error("spigotmc return invalid versions list")
		return
	}

	if mcVersion, err := ListMojang(); err != nil {
		t.Error(err)
		return
	} else if len(mcVersion) == 0 {
		t.Error("mojang return invalid versions list")
		return
	}

	if mcVersion, err := ListPurpur(); err != nil {
		t.Error(err)
		return
	} else if len(mcVersion) == 0 {
		t.Error("purpur project return invalid versions list")
		return
	}
}
