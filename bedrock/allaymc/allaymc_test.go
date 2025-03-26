package allaymc

import (
	"encoding/json"
	"testing"
)

func TestListVersions(t *testing.T) {
	vers := Versions{}
	if err := vers.FetchFromGithub(); err != nil {
		t.Error(err)
		return
	}

	d, _ := json.MarshalIndent(vers, "", "  ")
	t.Log(string(d))
}
