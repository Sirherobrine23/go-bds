package adoptium

import (
	"testing"
)

func TestAdoptium(t *testing.T) {
	rels, err := Releases()
	if err != nil {
		t.Error(err)
		return
	}

	if len(rels) == 0 {
		t.Errorf("cannot get java releases")
		return
	}
}
