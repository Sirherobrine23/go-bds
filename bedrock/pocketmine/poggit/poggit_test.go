package poggit

import (
	"testing"
)

func TestPoggit(t *testing.T) {
	poggit, err := NewPoggitClient("https://poggit.pmmp.io")
	if err != nil {
		t.Error(err)
		return
	}

	err = poggit.ListPlugins()
	if err != nil {
		t.Error(err)
		return
	} else if len(poggit.Plugins) == 0 {
		t.Errorf("Plugins list not correct parsed")
		return
	}
}
