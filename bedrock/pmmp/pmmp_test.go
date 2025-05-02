package pmmp

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
)

func TestPocketmine(t *testing.T) {
	phps := PHPs{}
	versions := Versions{}
	defer func() {
		// Write tmp file
		tmpFile, _ := os.Create("./versions.json")
		defer tmpFile.Close()
		js := json.NewEncoder(tmpFile)
		js.SetIndent("", "  ")
		js.Encode([]any{versions, phps})
	}()

	var lock sync.Mutex
	lock.Lock()
	t.Run("phpBuilds", func(t *testing.T) {
		defer lock.Unlock()
		if err := phps.FetchAllScripts(t.TempDir()); err != nil {
			t.Errorf("Cannot get all PHPs builds: %s", err)
			return
		}
	})

	t.Run("pocketmine_versions", func(t *testing.T) {
		lock.Lock()
		defer lock.Unlock()

		if err := versions.GetVersionsFromGithub(phps); err != nil {
			switch len(versions) {
			case 0:
				t.Errorf("Cannot get PHPs builds: %s", err)
			default:
				t.Skipf("Cannot get all PHPs builds: %s", err)
			}
			return
		}
	})
}
