package purpur

import (
	"testing"
)

func TestReleases(t *testing.T) {
	version, err := Releases()
	if err != nil {
		t.Error(err)
		return
	}

	if len(version) == 0 {
		t.Errorf("cannot get Purpur versions")
		return
	}
}
