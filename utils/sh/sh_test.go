package sh

import (
	"encoding/json"
	"os"
	"testing"
)

var (
	CompileSH, _  = os.ReadFile("./_testdata/compile.sh")
	CompileSH2, _ = os.ReadFile("./_testdata/compile_2.sh")
	WinVS, _      = os.ReadFile("./_testdata/windows-compile-vs.ps1")
	WinVSBat, _   = os.ReadFile("./_testdata/windows-compile-vs.bat")

	Scripts = map[string][]byte{
		"./compile.sh":             CompileSH,
		"./compile_2.sh":           CompileSH2,
		"./windows-compile-vs.ps1": WinVS,
		"./windows-compile-vs.bat": WinVSBat,
	}
)

func TestSh(t *testing.T) {
	for keyName, value := range Scripts {
		t.Logf("Starting:\n%s", keyName)
		d, _ := json.MarshalIndent(ProcessSh(string(value)), "", "  ")
		t.Logf("%s", d)
	}
}
