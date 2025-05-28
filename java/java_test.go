package java

import (
	"encoding/json"
	"testing"
)

// List versions
func TestListVersions(t *testing.T) {
	t.Run("Mojang", func(t *testing.T) {
		t.Parallel()
		vers := Versions{}
		if err := vers.FetchMojang(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Java versions: %s", d)
	})

	t.Run("Spigot", func(t *testing.T) {
		t.Parallel()
		vers := Versions{}
		if err := vers.FetchSpigotVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Spigot versions: %s", d)
	})

	t.Run("Purpur", func(t *testing.T) {
		t.Parallel()
		vers := Versions{}
		if err := vers.FetchPurpurVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Purpur versions: %s", d)
	})

	t.Run("Paper", func(t *testing.T) {
		t.Parallel()
		vers := Versions{}
		if err := vers.FetchPaperVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Paper versions: %s", d)
	})

	t.Run("Folia", func(t *testing.T) {
		t.Parallel()
		vers := Versions{}
		if err := vers.FetchFoliaVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Folia versions: %s", d)
	})

	t.Run("Velocity", func(t *testing.T) {
		t.Parallel()
		vers := Versions{}
		if err := vers.FetchVelocityVersions(); err != nil {
			t.Error(err)
			return
		}
		d, _ := json.MarshalIndent(vers, "", "  ")
		t.Logf("Velocity versions: %s", d)
	})
}
