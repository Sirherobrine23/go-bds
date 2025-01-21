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

	if err := proot.DownloadUbuntuRootfs("", ""); err != nil {
		t.Error(err)
		return
	}

	process := ProcExec{
		Arguments: []string{
			"/bin/bash",
			"-c",
			"echo test",
		},
	}

	if err := proot.Start(process); err != nil {
		t.Error(err)
		return
	} else if err := proot.Wait(); err != nil {
		t.Error(err)
		return
	}
}
