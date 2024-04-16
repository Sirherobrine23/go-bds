package liberica

import "testing"

func TestLiberica(t *testing.T) {
	rels, err := Releases()
	if err != nil {
		t.Error(err)
		return
	} else if len(rels) == 0 {
		t.Errorf("cannot get releases to liberica")
		return
	}
}
