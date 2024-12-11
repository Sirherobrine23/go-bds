package exec

import (
	"os/exec"
	"testing"
)

func TestOs(t *testing.T) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("cannot get go version, %s", err.Error())
		return
	}

	sysProc := &Os{}
	if err = sysProc.Start(ProcExec{Arguments: []string{goPath, "version"}}); err != nil {
		t.Error(err)
		return
	} else if err = sysProc.Wait(); err != nil {
		t.Error(err)
		return
	}
}
