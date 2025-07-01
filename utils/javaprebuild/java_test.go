package javaprebuild

import (
	"path/filepath"
	"testing"
)

func TestInstallJava(t *testing.T) {
	if err := JavaVersion(21 + 44).InstallLatest(filepath.Join(t.TempDir(), "javatest")); err != nil && err != ErrSystem {
		t.Error(err)
		return
	}
}

func TestLocal(t *testing.T) {
	binPath, err := JavaVersion(21 + 44).Install(filepath.Join(t.TempDir(), "javatest"))
	if err != nil && err != ErrSystem {
		t.Error(err)
		return
	}
	t.Logf("Java path: %q", binPath)
}
