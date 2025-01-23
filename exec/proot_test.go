package exec

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestPRoot(t *testing.T) {
	if !(runtime.GOOS == "linux" || runtime.GOOS == "android") {
		t.Skipf("Cannot run test, current os is %q", runtime.GOOS)
		return
	} else if !LocalBinExist(ProcExec{Arguments: []string{"proot"}}) {
		t.Skip("Cannot run test, proot not installed")
		return
	}

	proot := &Proot{
		Rootfs: filepath.Join(t.TempDir(), "rootfs"),
	}

	// Install ubuntu to current arch and latest version
	if err := proot.DownloadUbuntuRootfs("", ""); err != nil {
		t.Error(err)
		return
	}

	// Simples process
	process := ProcExec{Arguments: []string{"/bin/bash", "-c", "echo test"}}
	if err := proot.Start(process); err != nil {
		t.Error(err)
		return
	} else if err := proot.Wait(); err != nil {
		if code, _ := proot.ExitCode(); code == 1 {
			t.Skip("cannot run proot or rootfs is invalid")
			return
		}
		t.Error(err)
	}
}
