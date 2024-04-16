package microsoft

import "testing"

func TestMicrosoft(t *testing.T) {
	rels, err := Releases()
	if err != nil {
		t.Error(err)
		return
	} else if len(rels) == 0 {
		t.Errorf("cannot get java releases for microsoft")
		return
	}
}
