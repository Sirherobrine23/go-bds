package paper

import (
	"testing"
)

func TestReleases(t *testing.T) {
	b := Versions{Project: "paper"}
	if err := b.Releases(); err != nil {
		t.Error(err)
		return
	}

	if len(b.Versions) == 0 {
		t.Errorf("cannot get releases to %s", b.Project)
		return
	}
}
