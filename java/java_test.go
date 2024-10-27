package java

import (
	"fmt"
	"path/filepath"
	"testing"
)

// List all targets java versions and return Count
func TestListVersions(t *testing.T) {
	var (
		Mojang MojangSearch
		Spigot SpigotSearch
		Purpur PurpurSearch
	)

	// Mojang
	t.Log("Getting versions to Mojang server")
	if err := Mojang.list(); err != nil {
		t.Error(err)
		return
	}
	t.Logf("Mojang releases counting: %d", len(Mojang.Version))

	// Spigot
	t.Log("Getting versions to Spigot server")
	if err := Spigot.list(); err != nil {
		t.Error(err)
		return
	}
	t.Logf("Spigot releases counting: %d", len(Spigot.Version))

	// Purpur
	t.Log("Getting versions to Purpur server")
	if err := Purpur.list(); err != nil {
		t.Error(err)
		return
	}
	t.Logf("Purpur releases counting: %d", len(Purpur.Version))

	// Paper
	for _, PaperProject := range paperProjects {
		t.Run(fmt.Sprintf("Paper, project %q", PaperProject), func(t *testing.T) {
			var Paper PaperSearch
			Paper.ProjectTarget = PaperProject
			t.Logf("Getting versions to %s server", Paper.ProjectTarget)
			if err := Purpur.list(); err != nil {
				t.Error(err)
				return
			}
			t.Logf("%s releases counting: %d", Paper.ProjectTarget, len(Purpur.Version))
		})
	}
}

// Test spigot build to install server
func TestSpigotBuild(t *testing.T) {
	// Static Version
	version := SpigotMC{
		Version:      "1.21.1",
		ToolVersion:  181,
		JavaVersions: []uint{65, 67},
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
