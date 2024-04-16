package adoptium

import (
	"encoding/json"
	"testing"
)

func TestXxx(t *testing.T) {
	rels, err := Releases()
	if err != nil {
		t.Error(err)
		return
	}

	if len(rels) == 0 {
		t.Errorf("cannot get java releases")
		return
	}

	data, _ := json.MarshalIndent(&rels, "", "  ")
	t.Log(string(data))
}