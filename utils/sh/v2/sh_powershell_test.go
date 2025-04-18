package sh

import (
	"encoding/json"
	"os"
	"testing"
)

var ps1File, _ = os.ReadFile("../_testdata/windows-compile-vs.ps1")

func TestPowershell(t *testing.T) {
	ps1, err := PowershellScript(string(ps1File))
	if err != nil {
		t.Error(err)
		return
	}

	d, _ := json.MarshalIndent(ps1, "", "  ")
	t.Logf("%s", d)
}
