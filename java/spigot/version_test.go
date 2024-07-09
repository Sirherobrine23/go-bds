package spigot

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {
	rel, err := GetReleases()
	if err != nil {
		t.Error(err)
		return
	}
	d,_:=json.MarshalIndent(rel, "", "  ")
	fmt.Println(string(d))
}