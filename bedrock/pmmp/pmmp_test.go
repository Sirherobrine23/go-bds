package pmmp

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPHPBuilds(t *testing.T) {
	phps := PHPs{}
	if err := phps.FetchAllScripts(t.TempDir()); err != nil {
		t.Errorf("Cannot get all PHPs builds: %s", err)
		return
	}

	// Write tmp file
	tmpFile, _ := os.Create("./phps_versions.json")
	defer tmpFile.Close()
	js := json.NewEncoder(tmpFile)
	js.SetIndent("", "  ")
	js.Encode(phps)
}

func TestVersionsGithub(t *testing.T) {
	versions := Versions{}
	if err := versions.GetVersionsFromGithub(); err != nil {
		switch len(versions) {
		case 0:
			t.Errorf("Cannot get PHPs builds: %s", err)
		default:
			t.Skipf("Cannot get all PHPs builds: %s", err)
		}
		return
	}

	// Write tmp file
	tmpFile, _ := os.Create("./versions.json")
	defer tmpFile.Close()
	js := json.NewEncoder(tmpFile)
	js.SetIndent("", "  ")
	js.Encode(versions)
}
