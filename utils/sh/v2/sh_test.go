package sh

import (
	"encoding/json"
	"os"
	"testing"
)

var (
	ps1File, _ = os.ReadFile("../_testdata/windows-compile-vs.ps1")
	cmdFile, _ = os.ReadFile("../_testdata/windows-compile-vs.bat")
)

func TestPowershell(t *testing.T) {
	ps1, err := PowershellScript(string(ps1File))
	if err != nil {
		t.Error(err)
		return
	}

	d, _ := json.MarshalIndent(ps1, "", "  ")
	t.Logf("%s", d)

	for _, n1 := range ps1.(*Powershell).Variables {
		if !(n1.Type == VarSetArray && n1.Name == "PHP_VERSIONS") {
			continue
		}
		for value := range n1.ContentArray() {
			t.Logf("Range value: %q", value)
		}
	}
}

func TestCmd(t *testing.T) {
	cmd, err := CommandPromptScript(string(cmdFile))
	if err != nil {
		t.Error(err)
		return
	}

	d, _ := json.MarshalIndent(cmd, "", "  ")
	t.Logf("%s", d)

	for _, n1 := range cmd.(*CommandPrompt).Variables {
		if !(n1.Type == VarSetArray && n1.Name == "PHP_VERSIONS") {
			continue
		}
		for value := range n1.ContentArray() {
			t.Logf("Range value: %q", value)
		}
	}
}
